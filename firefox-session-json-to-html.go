package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"time"
)

type TimeValue struct {
	Time *time.Time
}

func (t TimeValue) String() string {
	if t.Time != nil {
		return t.Time.String()
	}

	return ""
}

func MakeTime(str string) (time.Time, error) {
	const fmt = "2006-01-02T15:04:05"
	reh := regexp.MustCompile(`.+[tT](\d\d)`)
	rem := regexp.MustCompile(`.+[tT](\d\d):(\d\d)`)
	ret := regexp.MustCompile(`.+[tT](\d\d):(\d\d):(\d\d)`)
	rez := regexp.MustCompile(`.+([zZ]|([+\-](\d\d):(\d\d)))`)
	tnow := time.Now()
	location := tnow.Location()
	strs := ""
	if rez.MatchString(str) {
		if tm, err := time.Parse(time.RFC3339, str); err != nil {
			return tnow, err
		} else {
			return tm, nil
		}

	} else if ret.MatchString(str) {
		strs = str
	} else if rem.MatchString(str) {
		strs = str + ":00"
	} else if reh.MatchString(str) {
		strs = str + ":00:00"
	} else {
		strs = str + "T00:00:00"
	}

	if tm, err := time.ParseInLocation(fmt, strs, location); err != nil {
		return tnow, err
	} else {
		return tm, nil
	}
}

func (t TimeValue) Set(str string) error {
	if tm, err := MakeTime(str); err != nil {
		return err
	} else {
		*t.Time = tm
	}

	return nil
}

var tstart = &time.Time{}
var tend = &time.Time{}

var (
	inf    = flag.String("in", "/dev/stdin", "input file in JSON")
	outf   = flag.String("out", "/dev/stdout", "output file in HTML")
	drange = flag.Bool("range", false, "print range of dates represented in the dump")
	rev    = flag.Bool("reverse", false, "sort reverse-chronologically (most recent first)")
	upat   = flag.String("url-regex", "", "print only tabs where URL matches regex")
	tpat   = flag.String("title-regex", "", "print only tabs where title matches regex")
	max    = flag.Int("max", -1, "maximum number of tabs printed, -1 for unlimited")
)

type TabSession []struct {
	Windows map[string]map[string]Tab `json:"windows"` // map[string]TabList
	WindowsNumber int `json:"windowsNumber"`
	WindowsInfo map[string]WindowInfo `json:"windowsInfo"`
	TabsNumber int `json:"tabsNumber"`
	Name string `json:"name"`
	Date int64 `json:"date"`
	LastEditedTime int64 `json:"lastEditedTime"`
	Tag []interface{} `json:"tag"`
	SessionStartTime int64 `json:"sessionStartTime"`
	ID string `json:"id"`
}

type MutedInfo struct {
	Muted bool `json:"muted"`
}

type SharingState struct {
	Camera bool `json:"camera"`
	Microphone bool `json:"microphone"`
}

type TabList struct {
	Tabs map[string]Tab
}

type Tab struct {
	ID int `json:"id"`
	Index int `json:"index"`
	WindowID int `json:"windowId"`
	Highlighted bool `json:"highlighted"`
	Active bool `json:"active"`
	Attention bool `json:"attention"`
	Pinned bool `json:"pinned"`
	Status string `json:"status"`
	Hidden bool `json:"hidden"`
	Discarded bool `json:"discarded"`
	Incognito bool `json:"incognito"`
	Width int `json:"width"`
	Height int `json:"height"`
	LastAccessed int64 `json:"lastAccessed"`
	Audible bool `json:"audible"`
	MutedInfo MutedInfo `json:"mutedInfo"`
	IsArticle bool `json:"isArticle"`
	IsInReaderMode bool `json:"isInReaderMode"`
	SharingState SharingState `json:"sharingState"`
	SuccessorTabID int `json:"successorTabId"`
	CookieStoreID string `json:"cookieStoreId"`
	URL string `json:"url"`
	Title string `json:"title"`
	FavIconURL string `json:"favIconUrl"`
}

type WindowInfo struct {
	ID int `json:"id"`
	Focused bool `json:"focused"`
	Top int `json:"top"`
	Left int `json:"left"`
	Width int `json:"width"`
	Height int `json:"height"`
	Incognito bool `json:"incognito"`
	Type string `json:"type"`
	State string `json:"state"`
	AlwaysOnTop bool `json:"alwaysOnTop"`
	Title string `json:"title"`
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	*tend = time.Now()
	flag.Var(&TimeValue{tstart}, "start", "dump tabs last accessed on this date and after, RFC 3339 format with optional time and time zone, default to local time (2017-11-01[T00:00:00[-07:00]]) (Default is beginning of file)")
	flag.Var(&TimeValue{tend}, "end", "dump tabs last accessed on this date and before, in RFC 3339 format with optional time and time zone, default to local time (2017-11-01[T00:00:00[-07:00]]) (Default is end of file)")
	flag.Parse()

	if tend.Before(*tstart) {
		panic("range is nonsensical")
	}

	if (*max == 0) || (*max < -1) {
		panic("maximum is nonsensical")
	}

	var ret *regexp.Regexp
	if len(*tpat) > 0 {
		ret = regexp.MustCompile(*tpat)
	}

	var reu *regexp.Regexp
	if len(*upat) > 0 {
		reu = regexp.MustCompile(*upat)
	}

	input, err := ioutil.ReadFile(*inf)
	check(err)

	output, erro := os.Create(*outf)
	check(erro)
	defer output.Close()

	writer := bufio.NewWriter(output)

	var dump TabSession
	err = json.Unmarshal([]byte(input), &dump)
	check(err)

	items := make(map[int64]Tab)
	var keys []int64
	for i, v := range dump[0].Windows {
		for j, w := range v {
			stamp := w.LastAccessed/1000
			items[stamp] = w
			keys = append(keys, stamp)
		}
	}

	if *rev {
		sort.Slice(keys, func(i, j int) bool { return keys[i] > keys[j] })
	} else {
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	}

	st := keys[0]
	if !tstart.IsZero() {
		st = tstart.Unix()
	}

	et := keys[len(keys)-1]
	if tend.Before(time.Unix(et, 0)) {
		et = tend.Unix()
	}

	fmt.Fprintf(writer, "<!DOCTYPE html><html>\n<head><meta charset=\"utf-8\"><title>Session Dump</title></head>\n")
	fmt.Fprintf(writer, "<body>\n")
	if *drange {
		fmt.Fprintf(writer, "<h1>%s - %s</h1>\n", time.Unix(keys[0], 0), time.Unix(keys[len(keys)-1], 0));
	}

	fmt.Fprintf(writer, "<ol>\n")
	for _, key := range keys {
		if key < st {
			continue
		}

		if key > et {
			continue
		}

		v := items[key]

		if reu != nil {
			if !reu.Match([]byte(v.URL)) {
				continue
			}
		}

		if *max != -1 {
			if *max > 0 {
				*max--
			} else {
				break
			}
		}

		when := time.Unix(key, 0)

		if ret != nil {
			if !ret.Match([]byte(v.Title)) {
				continue
			}
		}

		fmt.Fprintf(writer, "<li>%s <a href=\"%s\">%s</a></li>\n", when.Format(time.UnixDate), v.URL, html.EscapeString(v.Title))
	}

	fmt.Fprintf(writer, "</ol>\n</body></html>")
	writer.Flush()
}
