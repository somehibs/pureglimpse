package pureglimpse

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func ReadUrl(url string) []byte {
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

func ReadUrlCached(url string) []byte {
	cachePath := CachePath(url)
	cached := ReadFile(cachePath)
	if cached != nil {
		return cached
	}
	body := ReadUrl(url)
	WriteFile(cachePath, body)
	return body
}

func WriteFile(path string, body []byte) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Printf("Could not write response to cache %s\n", path)
		return
	}
	defer file.Close()
	file.Write(body)
}

func CachePath(url string) string {
	return fmt.Sprintf("data/cache/%s", strings.Replace(url, "/", "_", -1))
}

func ReadFile(path string) []byte {
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
