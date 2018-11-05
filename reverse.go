package pureglimpse

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var ApktoolUrl = "https://bitbucket.org/iBotPeaches/apktool/downloads/apktool_2.3.4.jar"
var ApktoolPath = fmt.Sprintf("%s/util/apktool.jar", WorkingDir)

// Reverser can reverse APK files (through unzipping or APKtool where Java is available)
// Should automatically fetch APKtool where possible
type Reverser struct {
	AppStream chan AppItem
}

// give us a path to the original apk, give us a reference for this process
func NewReverser() Reverser {
	r := Reverser{make(chan AppItem, 20)}
	r.CheckApkTool()
	return r
}

func (r *Reverser) StreamAppsForever(appReversed chan AppItem) {
	for {
		app := <-r.AppStream
		r.ReverseApp(app)
		appReversed <- app
	}
}

var ReversedRoot = "data/rev/"

func (r Reverser) ReverseApp(app AppItem) {
	apkPath := ApkPath(app.PackageId, app.CurrentVersion)
	outputPath := ReversedRoot + app.PackageId + "/" + app.CurrentVersion
	f, e := os.Open(outputPath)
	if e == nil {
		fmt.Println("Already reversed "+app.PackageId+" ", f)
		return
	}
	fmt.Println("Reverse app requested: ", app)
	// Run java
	cmd := exec.Command("java", "-jar", ApktoolPath, "d", "-o", outputPath, apkPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		panic(err.Error())
	}
	time.Sleep(15 * time.Second)
	fmt.Println(out.String())
}

func (r Reverser) CheckApkTool() bool {
	_, e := os.Open(ApktoolPath)
	if e == nil {
		// Found APK tool ok
		fmt.Println("apktool.jar found, no need to continue")
		return true
	} else if e != os.ErrNotExist {
		panic("Could not open APKtool.jar: " + e.Error())
	}
	fmt.Println("Fetching apktool.jar")
	body := ReadUrl(ApktoolUrl)
	if body == nil {
		panic("Could not fetch APKtool.jar")
	}
	WriteFile(ApktoolPath, body)
	return true
}
