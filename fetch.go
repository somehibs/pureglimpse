package pureglimpse

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/beevik/etree"
)

var ApkPureUrl = "https://apkpure.com"
var ApkPureArgs = "?ajax=1&page=%d"
var ApkPureDl = "download?from=details"

type Fetcher struct {
	appFetched chan int
}

func NewFetcher() Fetcher {
	f := Fetcher{}
	apps := f.ListApps(1)
	for _, app := range apps {
		fmt.Println(app)
		f.FetchApp(app)
	}
	return f
}

var re = regexp.MustCompile(`.*<a id="download_link" .+ href="(.+)">click here</a>`)
var apkRe = regexp.MustCompile(`>(.+)_([v\.0-9]+)_apkpure\.com\.apk`)

func (f *Fetcher) FetchApp(app ApkItem) bool {
	// Check if file exists
	apkPath := fmt.Sprintf("data/apks/%s.apk", app.PackageId)
	_, e := os.Open(apkPath)
	if e == nil {
		fmt.Printf("found apk at path %s - not downloading", apkPath)
		return true
	}

	// Fetch APK redirect from download page
	apkUrl := fmt.Sprintf("%s%s/%s", ApkPureUrl, app.Url, ApkPureDl)
	resp := f.Get(apkUrl)
	matches := re.FindStringSubmatch(string(resp))
	// Found URL in page (probably)
	dlUrl := matches[1]
	matches = apkRe.FindStringSubmatch(string(resp))
	//apkData :=
	return false
	fmt.Printf("Downloading pkg: %s (title: %s)\n", app.PackageId, app.Title)
	// Dump the metadata about the APK
	apkBytes := f.Get(dlUrl)
	fmt.Printf("Writing pkg %s to %s\n", app.PackageId, apkPath)
	f.WriteFile(apkPath, apkBytes)
	return true
}

type ApkItem struct {
	Title          string
	Url            string
	PackageId      string
	CurrentVersion string
}

func (li *ApkItem) GenPackageId() string {
	li.PackageId = strings.Split(li.Url, "/")[2]
	return li.PackageId
}

func (f *Fetcher) ListGames(index int) []ApkItem {
	return f.List("game", index)
}

func (f *Fetcher) ListApps(index int) []ApkItem {
	return f.List("app", index)
}

func (f *Fetcher) ParseList(body string) []ApkItem {
	reader := strings.NewReader(body)
	doc := etree.NewDocument()
	_, err := doc.ReadFrom(reader)
	if err != nil {
		panic(err)
	}
	var items []ApkItem
	for _, app := range doc.SelectElements("li") {
		li := ApkItem{}
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

func (f *Fetcher) Get(url string) []byte {
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
	return body
}

func (f *Fetcher) Fetch(url string) []byte {
	cachePath := f.CachePath(url)
	cached := f.ReadFile(cachePath)
	if cached != nil {
		return cached
	}
	body := f.Get(url)
	f.WriteFile(cachePath, body)
	return body
}

func (f *Fetcher) WriteFile(path string, body []byte) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Printf("Could not write response to cache %s\n", path)
		return
	}
	defer file.Close()
	file.Write(body)
}

func (f *Fetcher) CachePath(url string) string {
	return fmt.Sprintf("data/cache/%s", strings.Replace(url, "/", "_", -1))
}

func (f *Fetcher) ReadFile(path string) []byte {
	file, err := os.Open(path)
	if err == nil {
		dat, err := ioutil.ReadAll(file)
		if err != nil {
			return nil
		}
		fmt.Printf("read from cache (%s)\n", path)
		return dat
	}
	fmt.Println(err.Error())
	return nil
}

func (f *Fetcher) List(cat string, page int) []ApkItem {
	url := fmt.Sprintf(ApkPureUrl+"/"+cat+ApkPureArgs, page)
	body := f.Fetch(url)
	return f.ParseList(string(body))
}
