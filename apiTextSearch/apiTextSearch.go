package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// json for recieve file contents and id
type event struct {
	Mac      string `json:"mac"`
	ID       string `json:"id"`
	Contents string `json:"contents"`
}

func receiveContentAndID(w http.ResponseWriter, r *http.Request) {
	var newEvent event
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "data is not enable")
	} else {
		json.Unmarshal(reqBody, &newEvent)
		w.WriteHeader(http.StatusCreated)
		if moveFolderToSaved(newEvent.ID) {
			pushToDb(newEvent)
		}
	}
}

func pushToDb(item event) {
	url := "http://localhost:9200/" + item.Mac + "/documents/"
	payload := strings.NewReader("{\n\t\"id\":\"" + item.ID + "\",\n\t\"contents\":\"" + item.Contents + "\"\n}")
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("postman-token", "a842a97b-8d7b-9fe6-e304-f8420e7e1892")
	res, _ := http.DefaultClient.Do(req)
	res.Body.Close()
}

func makeNewDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
}

func returnAllFileName(root string) []string {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		temp := strings.Split(path, "/")
		files = append(files, temp[len(temp)-1])
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files[1:]
}

func moveFolderToSaved(folderID string) bool {
	newDir := "./saved/" + folderID
	oldDir := "./temp/" + folderID
	if exists(oldDir) {
		makeNewDir(newDir)
		files := returnAllFileName(oldDir)
		for _, f := range files {
			err := os.Rename(oldDir+"/"+f, newDir+"/"+f)
			if err != nil {
				log.Fatal(err)
			}
		}
		os.RemoveAll(oldDir)
		return true
	}
	return false
}

// exists returns whether the given file or directory exists
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func setupRoutes() {
	http.HandleFunc("/pushtextandid", receiveContentAndID)
	http.ListenAndServe(":8080", nil)
}

func main() {
	setupRoutes()
}
