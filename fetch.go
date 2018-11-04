package pureglimpse

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/beevik/etree"
)

var ApkPureUrl = "https://apkpure.com/"
var ApkPureArgs = "?ajax=1&page=%d"

type Fetcher struct {
}

type ListItem struct {
	Title     string
	Url       string
	PackageId string
}

func (li *ListItem) GenPackageId() string {
	li.PackageId = strings.Split(li.Url, "/")[1]
	return li.PackageId
}

func NewFetcher() Fetcher {
	f := Fetcher{}
	fmt.Printf("Found apps: %+v\n", f.ListApps(1))
	return f
}

func (f *Fetcher) ListApps(index int) []ListItem {
	return f.List("app", index)
}

func (f *Fetcher) ParseList(body string) []ListItem {
	reader := strings.NewReader(body)
	doc := etree.NewDocument()
	_, err := doc.ReadFrom(reader)
	if err != nil {
		panic(err)
	}
	var items []ListItem
	for _, app := range doc.SelectElements("li") {
		li := ListItem{}
		divs := app.SelectElements("div")
		if len(divs) != 4 {
			fmt.Printf("Diag: app: %+v div: %+v\n", app, divs)
			panic(fmt.Sprintf("Expected div length was 4, got %d", len(divs)))
		}
		titleLinks := divs[1].SelectElement("a")
		li.Title = titleLinks.Text()
		for _, attr := range titleLinks.Attr {
			if attr.Key == "href" {
				li.Url = attr.Value
			}
		}
		li.GenPackageId()
		items = append(items, li)
	}
	return items
}

func (f *Fetcher) Fetch(url string) []byte {
	cached := f.Cache(url)
	if cached != nil {
		return cached
	}
	resp, err := http.Get(url)
	if err != nil {
		panic(fmt.Sprintf("Could not fetch page for list url: %s due to ", url, err.Error()))
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("Could not read page for url %s - status was %s", url, resp.Status))
	}

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		panic(fmt.Sprintf("Failed to read all bytes as %s", e.Error()))
	}
	f.PutCache(url, body)
	return body
}

func (f *Fetcher) PutCache(url string, body []byte) {
	cachePath := fmt.Sprintf("data/cache/%s", strings.Replace(url, "/", "_", -1))
	file, err := os.Create(cachePath)
	if err != nil {
		fmt.Printf("Could not write response to cache %s\n", cachePath)
		return
	}
	defer file.Close()
	file.Write(body)
}

func (f *Fetcher) Cache(url string) []byte {
	cachePath := fmt.Sprintf("data/cache/%s", strings.Replace(url, "/", "_", -1))
	file, err := os.Open(cachePath)
	if err == nil {
		dat, err := ioutil.ReadAll(file)
		if err != nil {
			return nil
		}
		fmt.Printf("read from cache for %s\n", url)
		return dat
	}
	fmt.Println(err.Error())
	return nil
}

func (f *Fetcher) List(cat string, page int) []ListItem {
	url := fmt.Sprintf(ApkPureUrl+cat+ApkPureArgs, page)
	body := f.Fetch(url)
	return f.ParseList(string(body))
}
