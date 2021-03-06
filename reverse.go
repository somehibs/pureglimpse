package pureglimpse

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pierrre/archivefile/zip"
)

var ApktoolUrl = "https://bitbucket.org/iBotPeaches/apktool/downloads/apktool_2.3.4.jar"
var ApktoolPath = fmt.Sprintf("%s/util/apktool.jar", WorkingDir)
var ReversedRoot = "data/rev/"

// Reverser can reverse APK files (through unzipping or APKtool where Java is available)
// Should automatically fetch APKtool where possible
type Reverser struct {
	AppStream chan AppItem
}

// give us a path to the original apk, give us a reference for this process
func NewReverser() Reverser {
	r := Reverser{make(chan AppItem, 20)}
	os.MkdirAll(ReversedRoot, 0755)
	os.MkdirAll(WorkingDir+"/util", 0755)
	r.CheckApkTool()
	return r
}

func (r *Reverser) StreamAppsForever(appReversed chan AppItem) {
	for {
		app := <-r.AppStream
		if app.PackageId == "" {
			// dead app, shut down
			appReversed <- app
			return
		}
		r.ReverseApp(app)
		appReversed <- app
	}
}

func (r Reverser) ReverseApp(app AppItem) {
	apkPath := ApkPath(app.PackageId, app.CurrentVersion)
	outputPath := ReversedRoot + app.PackageId + "/" + app.CurrentVersion
	_, e := os.Open(outputPath + ".zip")
	if e == nil {
		//fmt.Println("Already reversed "+app.PackageId+" ", f)
		return
	}
	fmt.Println("Reverse app requested: ", app)
	// Run java
	cmd := exec.Command("java", "-jar", ApktoolPath, "d", "-o", outputPath, apkPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		os.RemoveAll(outputPath)
		fmt.Printf("Could not decode %s\n", app.PackageId)
		return
	}
	fmt.Println(out.String())
	fmt.Println("Compressing result...")
	r.ZipPath(outputPath)
	fmt.Println("Destroying result folder...")
	e = os.RemoveAll(outputPath)
	fmt.Printf("Deleted result folder %s (err: %s)\n", outputPath, e)
	time.Sleep(5 * time.Second)
}

func (r Reverser) ZipPath(path string) {
	// Take the output path and turn it into a zip before deleting the original output path
	zipPath := path + ".zip"
	e := zip.ArchiveFile(path, zipPath, nil)
	if e != nil {
		panic(e.Error())
	}
}

func (r Reverser) CheckApkTool() bool {
	_, e := os.Open(ApktoolPath)
	if e == nil {
		// Found APK tool ok
		fmt.Println("apktool.jar found, no need to continue")
		return true
	} else {
		//		panic("Could not open APKtool.jar: " + e.Error())
	}
	fmt.Println("Fetching apktool.jar")
	body := ReadUrl(ApktoolUrl)
	if body == nil {
		panic("Could not fetch APKtool.jar")
	}
	WriteFile(ApktoolPath, body)
	return true
}
