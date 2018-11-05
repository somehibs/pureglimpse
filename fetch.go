package pureglimpse

import (
	"encoding/json"
	"fmt"
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
	return f
}

func (f *Fetcher) StreamList(listStream chan AppItem, max int) {
	fetched := 0
	i := 1
	for fetched < max {
		apps := f.ListApps(i)
		i += 1
		for _, app := range apps {
			listStream <- f.FetchApp(app)
		}
		fetched += len(apps)
	}
}

var re = regexp.MustCompile(`.*<a id="download_link" .+ href="(.+)">click here</a>`)
var apkRe = regexp.MustCompile(`>(.+)_v(.+)_apkpure\.com\.apk`)
var ApkDir = "data/apks/"

func ApkPath(pkgId, version string) string {
	return fmt.Sprintf("%s%s/%s.apk", ApkDir, pkgId, version)
}

func ApkManifestPath(pkgId string) string {
	return fmt.Sprintf("data/apks/%s/manifest.json", pkgId)
}

func (f *Fetcher) FetchApp(app AppItem) AppItem {
	// Check if manifest exists
	apkJsonPath := ApkManifestPath(app.PackageId)
	data := ReadFile(apkJsonPath)
	if data != nil {
		fmt.Printf("found apk at path %s - not downloading\n", apkJsonPath)
		var decoded AppItem
		json.Unmarshal(data, &decoded)
		return decoded
	}

	// fetch download page
	apkUrl := fmt.Sprintf("%s%s/%s", ApkPureUrl, app.Url, ApkPureDl)
	resp := ReadUrl(apkUrl)
	// Found URL in page (probably)
	matches := re.FindStringSubmatch(string(resp))
	dlUrl := matches[1]
	// found version in page
	matches = apkRe.FindStringSubmatch(string(resp))
	if len(matches) < 2 {
		panic("could not match in resp: " + string(resp))
	}
	app.CurrentVersion = matches[2]
	// prepare path
	os.Mkdir(ApkDir+app.PackageId, 0755)
	apkPath := ApkPath(app.PackageId, app.CurrentVersion)
	fmt.Printf("Downloading pkg: %s (title: %s)\n", app.PackageId, app.Title)
	// Dump the metadata about the APK
	apkBytes := ReadUrl(dlUrl)
	fmt.Printf("Writing pkg %s to %s\n", app.PackageId, apkPath)
	WriteFile(apkPath, apkBytes)
	appJson, _ := json.Marshal(app)
	WriteFile(apkJsonPath, appJson)
	return app
}

type AppItem struct {
	Title          string
	Url            string
	PackageId      string
	CurrentVersion string
}

func (li *AppItem) GenPackageId() string {
	li.PackageId = strings.Split(li.Url, "/")[2]
	return li.PackageId
}

func (f *Fetcher) ListGames(index int) []AppItem {
	return f.List("game", index)
}

func (f *Fetcher) ListApps(index int) []AppItem {
	return f.List("app", index)
}

func (f *Fetcher) ParseList(body string) []AppItem {
	reader := strings.NewReader(body)
	doc := etree.NewDocument()
	_, err := doc.ReadFrom(reader)
	if err != nil {
		panic(err)
	}
	var items []AppItem
	for _, app := range doc.SelectElements("li") {
		li := AppItem{}
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

func (f *Fetcher) List(cat string, page int) []AppItem {
	url := fmt.Sprintf(ApkPureUrl+"/"+cat+ApkPureArgs, page)
	body := ReadUrlCached(url)
	return f.ParseList(string(body))
}
