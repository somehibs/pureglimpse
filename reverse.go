package pureglimpse

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var ApktoolUrl = "https://bitbucket.org/iBotPeaches/apktool/downloads/apktool_2.3.4.jar"
var ApktoolPath = fmt.Sprintf("%s/util/apktool.jar", WorkingDir)

// Reverser can reverse APK files (through unzipping or APKtool where Java is available)
// Should automatically fetch APKtool where possible
type Reverser struct {
}

// give us a path to the original apk, give us a reference for this process
func NewReverser() Reverser {
	r := Reverser{}
	r.CheckApkTool()
	return r
}

func (r Reverser) CheckApkTool() bool {
	f, e := os.Open(ApktoolPath)
	if e == nil {
		// Found APK tool ok
		fmt.Println("apktool.jar found, no need to continue")
		return true
	} else if e != os.ErrNotExist {
		panic("Could not open APKtool.jar: " + e.Error())
	}
	fmt.Println("Fetching apktool.jar")
	resp, e := http.Get(ApktoolUrl)
	if e != nil {
		panic("Could not fetch APKtool.jar: " + e.Error())
	}
	defer resp.Body.Close()

	f, e = os.Create(ApktoolPath)
	if e != nil {
		panic("Could not create file for apktool.jar: " + e.Error())
	}

	_, e = io.Copy(f, resp.Body)
	if e != nil {
		panic("Could not copy data into file for apktool.jar: " + e.Error())
	}
	return true
}

func (r Reverser) Reverse(apkPath, workingDir string) bool {
	return false
}
