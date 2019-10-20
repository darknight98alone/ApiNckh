package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// json for recieve file contents and id
type event struct {
	Mac      string `json:"mac"`
	ID       string `json:"id"`
	Contents string `json:"contents"`
}

func retError(w http.ResponseWriter, message string) {
	fmt.Fprintf(w, message)
	w.WriteHeader(406)
}

func receiveContentAndID(w http.ResponseWriter, r *http.Request) {
	var newEvent event
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil || len(reqBody) == 0 {
		retError(w, "wrong data")
	} else {
		json.Unmarshal(reqBody, &newEvent)
		if standardString(newEvent.Mac) == "" || standardString(newEvent.ID) == "" {
			retError(w, "wrong data")
			return
		}
		if exists("./temp/" + newEvent.Mac + "/" + newEvent.ID) {
			w.WriteHeader(http.StatusCreated)
			err := pushToDb(newEvent)
			if !err {
				retError(w, "elastic disconnect")
				return
			}
			if !moveFolderToSaved(newEvent.Mac, newEvent.ID) {
				retError(w, "save fail")
				return
			}
			os.Remove("./saved/" + newEvent.Mac + "/" + newEvent.ID + "/text.txt")
			f, _ := os.OpenFile("./saved/"+newEvent.Mac+"/"+newEvent.ID+"/text.txt", os.O_CREATE|os.O_WRONLY, 0600)
			f.WriteString(newEvent.Contents)
		} else {
			retError(w, "wrong data")
			return
		}
	}
}

func pushToDb(item event) bool {
	url := "http://localhost:9200/" + item.Mac + "/document/"
	payload := strings.NewReader("{\n\t\"id\":\"" + item.ID + "\",\n\t\"contents\":\"" + item.Contents + "\"\n}")
	req, err1 := http.NewRequest("POST", url, payload)
	if err1 != nil {
		return false
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("postman-token", "a842a97b-8d7b-9fe6-e304-f8420e7e1892")
	res, err2 := http.DefaultClient.Do(req)
	if res == nil {
		return false
	}
	res.Body.Close()
	if err2 != nil {
		return false
	}
	return true
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

func moveFolderToSaved(mac string, folderID string) bool {
	newDir := "./saved/" + mac + "/" + folderID
	oldDir := "./temp/" + mac + "/" + folderID
	if exists(oldDir) {
		if !exists("./saved/" + mac) {
			makeNewDir("./saved/" + mac)
		}
		if !exists("./saved/" + mac + "/" + folderID) {
			makeNewDir(newDir)
		}
		files := returnAllFileName(oldDir)
		for _, f := range files {
			err := os.Rename(oldDir+"/"+f, newDir+"/"+f)
			if err != nil {
				log.Fatal(err)
			}
		}
		os.RemoveAll("./temp/" + mac)
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
	if err != nil || len(reqBody) == 0 {
		retError(w, "wrong data")
	} else {
		json.Unmarshal(reqBody, &newEvent)
		if standardString(newEvent.Mac) != "" && exists("./saved/"+newEvent.Mac) {
			result := searchElastic(newEvent)
			if result != "err" {
				fmt.Fprintf(w, result)
				w.WriteHeader(http.StatusCreated)
				return
			}
		}
		retError(w, "wrong data")
		return
	}
}

// one person have one database for searching use mac id

func searchElastic(item searchItem) string {
	url := "http://localhost:9200/" + item.Mac + "/document/_search?filter_path=hits.hits._source"
	limit := "5"
	payload := strings.NewReader("{\n  \"from\" : 0, \"size\" :" + limit + ",\n  \"query\": {\n    \"match\": {\n      \"contents\": {\n        \"query\": \"" + item.SearchContents + "\",\n        \"fuzziness\": 2\n      }\n    }\n  }\n}")
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("postman-token", "c5da7799-0243-0446-8a9d-0c59bd95a810")
	res, _ := http.DefaultClient.Do(req)
	if res == nil {
		return "err"
	}
	body, _ := ioutil.ReadAll(res.Body)
	temp := string(handleRet(body, 100, item.Mac))
	res.Body.Close()
	return temp
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

func handleRet(b []byte, limit int, mac string) []byte {
	var cont outest
	json.Unmarshal(b, &cont)
	for i := 0; i < len(cont.Z.Y); i++ {
		if exists("./saved/" + mac + "/" + cont.Z.Y[i].X.A) {
			cont.Z.Y[i].X.C = returnAllFileName("./saved/" + mac + "/" + cont.Z.Y[i].X.A)[0]
			if len(cont.Z.Y[i].X.B) > limit {
				cont.Z.Y[i].X.B = cont.Z.Y[i].X.B[:limit]
			}
		}
	}
	js, _ := json.Marshal(cont.Z.Y)
	return js
}

type idstruct struct {
	ID  string `json:"id"`
	Mac string `json:"mac"`
}

func standardString(input string) string {
	temp := strings.ToLower(input)
	temp = strings.ReplaceAll(temp, " ", "")
	return temp
}

func getAllContents(w http.ResponseWriter, r *http.Request) {
	var newEvent idstruct
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil || len(reqBody) == 0 {
		retError(w, "wrong data")
	} else {
		json.Unmarshal(reqBody, &newEvent)
		if standardString(newEvent.ID) == "" || standardString(newEvent.Mac) == "" {
			retError(w, "wrong data")
			return
		}
		dir := "./saved/" + newEvent.Mac + "/" + newEvent.ID
		if exists(dir) {
			listFile := returnAllFileName(dir)
			for _, value := range listFile {
				if strings.Contains(value, "txt") {
					content, _ := ioutil.ReadFile(dir + "/" + value)
					fmt.Fprint(w, string(content))
					break
				}
			}
			w.WriteHeader(http.StatusCreated)
		} else {
			retError(w, "wrong data")
			return
		}
	}
}

func downloadFile(writer http.ResponseWriter, r *http.Request) {
	var newEvent idstruct
	var fileName string
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil || len(reqBody) == 0 {
		retError(writer, "wrong data")
	} else {
		json.Unmarshal(reqBody, &newEvent)
		if standardString(newEvent.ID) != "" && standardString(newEvent.Mac) != "" {
			dir := "./saved/" + newEvent.Mac + "/" + newEvent.ID
			if exists(dir) {
				listFile := returnAllFileName(dir)
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
				writer.WriteHeader(http.StatusCreated)
			} else {
				retError(writer, "wrong data")
				return
			}
		} else {
			retError(writer, "wrong data")
			return
		}
	}
	Filename := "./saved/" + newEvent.Mac + "/" + newEvent.ID + "/" + fileName
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
	if err != nil || len(reqBody) == 0 {
		retError(writer, "wrong data")
	} else {
		json.Unmarshal(reqBody, &newEvent)
		if standardString(newEvent.ID) == "" || standardString(newEvent.Mac) == "" {
			retError(writer, "wrong data")
			return
		}
		dir := "./saved/" + newEvent.Mac + "/" + newEvent.ID
		if exists(dir) {
			listFile := returnAllFileName(dir)
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
			writer.WriteHeader(http.StatusCreated)
			return
		}
	}
	retError(writer, "wrong data")
	return
}

type option struct {
	Deskew       string `json:"deskew"`
	Deblur       string `json:"deblur"`
	TableBasic   string `json:"table_basic"`
	TableAdvance string `json:"table_advance"`
}

type customOptions struct {
	Mac string `json:"mac"`
	// FileType string   `json:"file_type"`
	// FileName string   `json:"file_name"`
	// Options  []option `json:"options"`
}

func putOptions(writer http.ResponseWriter, r *http.Request) {
	var newEvent customOptions
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil || len(reqBody) == 0 {
		retError(writer, "wrong data")
	} else {
		json.Unmarshal(reqBody, &newEvent)
		// dir := ""
		if standardString(newEvent.Mac) == "" {
			retError(writer, "wrong data")
			return
		}
		if !exists("./temp/" + newEvent.Mac) {
			os.Mkdir("./temp/"+newEvent.Mac, os.ModePerm)
		}
		id, err := uuid.NewUUID()
		for {
			if err == nil {
				if !exists("./saved/"+newEvent.Mac+"/"+id.String()) && !exists("./temp/"+newEvent.Mac+"/"+id.String()) {
					os.Mkdir("./temp/"+newEvent.Mac+"/"+id.String(), os.ModePerm)
					// dir = "./temp/" + newEvent.Mac + "/" + id.String()
					break
				}
			}
			id, err = uuid.NewUUID()
		}
		// file, _ := os.Create(path.Join(dir, "option.txt"))
		// file.WriteString(newEvent.FileName)
		// file.WriteString("\n" + newEvent.FileType)
		// for _, val := range newEvent.Options {
		// 	file.WriteString("\n" + val.Deskew)
		// 	file.WriteString("\n" + val.Deblur)
		// 	file.WriteString("\n" + val.TableBasic)
		// 	file.WriteString("\n" + val.TableAdvance)
		// }
		fmt.Fprintf(writer, id.String())
		writer.WriteHeader(http.StatusCreated)
	}
}

// localhost:8080/?id=1234&mac=1

func getInformationRequest(r *http.Request, key string) (string, bool) {
	temp, err := r.URL.Query()[key]
	id := strings.Join(temp, "")
	return id, err
}

func putFile(writer http.ResponseWriter, r *http.Request) {
	id, ok1 := getInformationRequest(r, "id")
	mac, ok2 := getInformationRequest(r, "mac")
	skew, ok3 := getInformationRequest(r, "skew")
	blur, ok4 := getInformationRequest(r, "blur")
	basic, ok5 := getInformationRequest(r, "basic")
	advance, ok6 := getInformationRequest(r, "advance")
	fileType, ok7 := getInformationRequest(r, "filetype")
	fileName, ok8 := getInformationRequest(r, "filename")
	tempdir := "./temp/" + mac + "/" + id + "/"
	// savedir = "./saved/" + mac + "/" + id
	reqBody, err := ioutil.ReadAll(r.Body)
	if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8) || !exists(tempdir) || err != nil || len(reqBody) == 0 {
		retError(writer, "wrong data")
		return
	}
	// save file
	if exists("./" + path.Join(tempdir, fileName+"."+fileType)) {
		retError(writer, "save fail")
		return
	}
	file, _ := os.Create("./" + path.Join(tempdir, fileName+"."+fileType))
	file.Write(reqBody)
	file.Close()
	if err != nil {
		retError(writer, "wrong data")
		return
	}
	folder := tempdir
	fileTextToSave := "text.txt"
	// python3 detai.py -ft pdf -fd ./save/ -fs text.txt
	var mutex = &sync.Mutex{}
	mutex.Lock()
	cmd := exec.Command("python3", "detai.py", "-ft", fileType, "-fd", folder, "-fs", fileTextToSave,
		"-skew", skew, "-blur", blur, "-basic", basic, "-advance", advance)
	mutex.Unlock()
	// return text to client
	log.Println(cmd.Run())
	writer.WriteHeader(http.StatusCreated)
	data, _ := ioutil.ReadFile(path.Join(tempdir, "text.txt"))
	fmt.Fprintf(writer, string(data))
}

func setupRoutes() {
	http.HandleFunc("/recieveFile/", putFile)              // 2. put file,mac,id,options return text
	http.HandleFunc("/putOptions", putOptions)             // 1. mac return id
	http.HandleFunc("/getRootFileExtension", getExt)       // 6. get extension before download
	http.HandleFunc("/download", downloadFile)             // 7. nhận id và cho phép download
	http.HandleFunc("/getAllContents", getAllContents)     // 5. API nhận ID, chỉ trả về file text
	http.HandleFunc("/search", search)                     // 4. search file
	http.HandleFunc("/pushtextandid", receiveContentAndID) // 3. push to elastic search
	http.ListenAndServe(":8080", nil)
}

func main() {
	setupRoutes()
}
