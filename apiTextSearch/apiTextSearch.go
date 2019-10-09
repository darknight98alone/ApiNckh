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
	url := "http://localhost:9200/" + item.Mac + "/document/"
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

type searchItem struct {
	Mac            string `json:"mac"`
	SearchContents string `json:"search_contents"`
}

func search(w http.ResponseWriter, r *http.Request) {
	var newEvent searchItem
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "data is not enable")
	} else {
		json.Unmarshal(reqBody, &newEvent)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, searchElastic(newEvent))
	}
}

func searchElastic(item searchItem) string {
	url := "http://localhost:9200/" + item.Mac + "/document/_search?filter_path=hits.hits._source"
	limit := "5"
	payload := strings.NewReader("{\n  \"from\" : 0, \"size\" :" + limit + ",\n  \"query\": {\n    \"match\": {\n      \"contents\": {\n        \"query\": \"" + item.SearchContents + "\",\n        \"fuzziness\": 2\n      }\n    }\n  }\n}")
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("postman-token", "c5da7799-0243-0446-8a9d-0c59bd95a810")
	res, _ := http.DefaultClient.Do(req)
	body, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return string(handleRet(body, 100))
}

type inner struct {
	A string `json:"id"`
	B string `json:"contents"`
	C string `json:"filename"`
}
type outer struct {
	X inner `json:"_source"`
}

type out struct {
	Y []outer `json:"hits"`
}

type outest struct {
	Z out `json:"hits"`
}

func handleRet(b []byte, limit int) []byte {
	var cont outest
	json.Unmarshal(b, &cont)
	for i := 0; i < len(cont.Z.Y); i++ {
		cont.Z.Y[i].X.C = returnAllFileName("./saved/" + cont.Z.Y[i].X.A)[0]
		if len(cont.Z.Y[i].X.B) > limit {
			cont.Z.Y[i].X.B = cont.Z.Y[i].X.B[:limit]
		}
	}
	js, _ := json.Marshal(cont.Z.Y)
	return js
}

func setupRoutes() {
	http.HandleFunc("/search", search)
	http.HandleFunc("/pushtextandid", receiveContentAndID)
	http.ListenAndServe(":8080", nil)
}

func main() {
	setupRoutes()
}
