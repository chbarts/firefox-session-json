package main

import (
	"bufio"
	"encoding/json"
//	"flag"
	"fmt"
//	"html"
	"io/ioutil"
	"os"
//	"regexp"
//	"sort"
//	"strconv"
//	"time"
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
	input, err := ioutil.ReadFile("/dev/stdin")
	check(err)

	output, erro := os.Create("/dev/stdout")
	check(erro)
	defer output.Close()

	writer := bufio.NewWriter(output)

	var dump TabSession
	err = json.Unmarshal([]byte(input), &dump)
	check(err)

	for i, v := range dump[0].Windows {
		for j, w := range v {
			// LastAccessed, URL, Title
			fmt.Fprintf(writer, "%s\t%s\t%d\t%s\t%s\n", i, j, w.LastAccessed, w.URL, w.Title)
		}
	}

	writer.Flush()
}
