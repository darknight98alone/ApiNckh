package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

type idstruct struct {
	ID string `json:"id"`
}

func getAllContents(w http.ResponseWriter, r *http.Request) {
	var newEvent idstruct
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "data is not enable")
	} else {
		json.Unmarshal(reqBody, &newEvent)
		w.WriteHeader(http.StatusCreated)
		if exists("./saved/" + newEvent.ID) {
			listFile := returnAllFileName("./saved/" + newEvent.ID)
			for _, value := range listFile {
				if strings.Contains(value, "txt") {
					content, _ := ioutil.ReadFile("./saved/" + newEvent.ID + "/" + value)
					fmt.Fprint(w, string(content))
					break
				}
			}
		}
	}
}

func downloadFile(writer http.ResponseWriter, r *http.Request) {
	var newEvent idstruct
	var fileName string
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(writer, "data is not enable")
	} else {
		json.Unmarshal(reqBody, &newEvent)
		writer.WriteHeader(http.StatusCreated)
		if exists("./saved/" + newEvent.ID) {
			listFile := returnAllFileName("./saved/" + newEvent.ID)
			for _, value := range listFile {
				if len(listFile) > 1 {
					if !strings.Contains(value, "txt") && strings.Contains(value, ".") {
						fileName = value
						break
					}
				} else {
					if strings.Contains(value, ".") {
						fileName = value
						break
					}
				}

			}
		}
	}
	Filename := "./saved/" + newEvent.ID + "/" + fileName
	//Check if file exists and open
	Openfile, err := os.Open(Filename)
	defer Openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(writer, "File not found.", 404)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	writer.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	writer.Header().Set("Content-Type", FileContentType)
	writer.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(writer, Openfile) //'Copy' the file to the client
	return
}

func getExt(writer http.ResponseWriter, r *http.Request) {
	var newEvent idstruct
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(writer, "data is not enable")
	} else {
		json.Unmarshal(reqBody, &newEvent)
		writer.WriteHeader(http.StatusCreated)
		if exists("./saved/" + newEvent.ID) {
			listFile := returnAllFileName("./saved/" + newEvent.ID)
			for _, value := range listFile {
				if len(listFile) > 1 {
					if !strings.Contains(value, "txt") && strings.Contains(value, ".") {
						fmt.Fprintf(writer, strings.Split(value, ".")[1])
						break
					}
				} else {
					if strings.Contains(value, ".") {
						fmt.Fprintf(writer, strings.Split(value, ".")[1])
						break
					}
				}
			}
		}
	}
}
func setupRoutes() {
	http.HandleFunc("/getRootFileExtension", getExt)
	http.HandleFunc("/download", downloadFile)
	http.HandleFunc("/getAllContents", getAllContents)
	http.HandleFunc("/search", search)
	http.HandleFunc("/pushtextandid", receiveContentAndID)
	http.ListenAndServe(":8080", nil)
}

func main() {
	setupRoutes()
}
