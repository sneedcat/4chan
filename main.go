package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

const baseurl = "https://boards.4channel.org/g/"

func getImages(path string, wg *sync.WaitGroup, url string) {
	resp, err := http.Get(url)
	if err != nil {
		wg.Done()
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		wg.Done()
		return
	}
	reg := regexp.MustCompile("i.4cdn.org/g/[0-9]*[.][a-zA-Z]*")
	list := reg.FindAll(body, -1)
	var waitGroup sync.WaitGroup
	for i := 0; i < len(list); i++ {
		waitGroup.Add(1)
		go downloadImages(path, &waitGroup, "https://"+string(list[i]))
	}
	waitGroup.Wait()
	wg.Done()
}

func downloadImages(path string, wg *sync.WaitGroup, url string) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()
	name := url[21:]
	f, err := os.Create(filepath.Join(path, name))
	if err != nil {
		return
	}
	defer f.Close()
	io.Copy(f, resp.Body)
}

func getThreads(path string, data []byte) {
	reg := regexp.MustCompile("\"[0-9]*?\":")
	var waitGroup sync.WaitGroup
	list := reg.FindAll(data, -1)
	for i := 0; i < len(list); i++ {
		id := list[i][1 : len(list[i])-2]
		waitGroup.Add(1)
		go getImages(path, &waitGroup, baseurl+"thread/"+string(id))
	}
	waitGroup.Wait()
}

func main() {
	resp, err := http.Get(baseurl + "catalog")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	path, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}
	path = filepath.Join(path, "4chan")
	err = os.RemoveAll(path)
	if err != nil {
		log.Println(err)
	}
	os.Mkdir(path, 0755)
	getThreads(path, body)
}
