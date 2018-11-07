package pureglimpse

import (
	"os"
	"path/filepath"
	"regexp"
)

var ignorePaths = []string{
	"res/drawable.*",
	"smali.*/android/support/.*",
	"smali.*/com.google/.*",
	"smali.*/com/android/support/.*",
}

type Scanner struct {
	ScanChan   chan AppItem
	CurrentDir string
	Ignores    []*regexp.Regexp
}

func compileRe(re []string) []*regexp.Regexp {
	comp := make([]*regexp.Regexp, len(re))
	for i, r := range re {
		comp[i] = regexp.MustCompile(".*" + r)
	}
	return comp
}

func NewScanner() Scanner {
	sc := Scanner{
		make(chan AppItem, 100),
		"",
		compileRe(ignorePaths),
	}
	return sc
}

func (s *Scanner) ScanAppsForever() {
	for {
		app := <-s.ScanChan
		if app.PackageId == "" {
			return
		}
		s.ScanApp(app)
	}
}

func (s *Scanner) ScanApp(app AppItem) {
	s.CurrentDir = ReversedRoot + app.PackageId + "/" + app.CurrentVersion + "/"
	filepath.Walk(s.CurrentDir, s.checkFile)
}

func (s *Scanner) checkFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	for _, re := range s.Ignores {
		if re.Match([]byte(path)) {
			return filepath.SkipDir
		}
	}
	if !info.IsDir() {
		//fmt.Println("Found file: ", path)
	}
	return nil
}
