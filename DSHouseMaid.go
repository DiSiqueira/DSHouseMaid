package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
    "os/user"
	"bufio"
)

type Format struct {
	Items []struct {
		Items []struct {
			Name string `json:"Name"`
		} `json:"Items"`
		Name string `json:"Name"`
	} `json:"Items"`
}

//https://gist.github.com/albrow/5882501
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

func posString(slice []string, element string) int {
	for index, elem := range slice {
		if strings.ToLower(elem) == strings.ToLower(element) {
			return index
		}
	}
	return -1
}

func askQuestion(question string) bool {

	fmt.Println(question + " (yes/no)")

	var response string
	_, err := fmt.Scanln(&response)

	if err != nil {
		fmt.Println(err)
	}

	okayResponses := []string{"y", "yes"}
	nokayResponses := []string{"n", "no"}

	if containsString(okayResponses, response) {
		return true
	}
	if containsString(nokayResponses, response) {
		return false
	}
	fmt.Println("Please type yes or no and then press enter:")
	return askQuestion(question)
}

//https://gist.github.com/albrow/5882501

//https://stackoverflow.com/questions/11692860/how-can-i-efficiently-download-a-large-file-using-go
func downloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

//https://stackoverflow.com/questions/11692860/how-can-i-efficiently-download-a-large-file-using-go

func move(oldfile, newFile string, link bool) error {
	if link {
		return os.Link(oldfile, newFile)
	}

	return os.Rename(oldfile, newFile)
}

func organize(formats Format, output, input string, link bool) {
	files, _ := ioutil.ReadDir(input)

	var wg sync.WaitGroup
	wg.Add(len(files))

	for _, f := range files {

		go func(file string) {

			ext := filepath.Ext(file)

			for _, folder := range formats.Items {
				for _, extension := range folder.Items {

					if ext == extension.Name {
						os.MkdirAll(output+string(filepath.Separator)+folder.Name, 0777)

						oldfile := input + string(filepath.Separator) + file
						newFile := output + string(filepath.Separator) + folder.Name + string(filepath.Separator) + file

						err := move(oldfile, newFile, link)

						fmt.Println(oldfile + " --> " + newFile)

						if err != nil {
							fmt.Println(err)
						}
					}

				}
			}

			defer wg.Done()

		}(f.Name())
	}

	wg.Wait()
}

func saveLib(formats Format, libDir string) {
	b, _ := json.Marshal(formats) 
	
	fileHandle, _ := os.Create(libDir + "formats.json")
	writer := bufio.NewWriter(fileHandle)
	fmt.Fprintln(writer, string(b))
	writer.Flush()
}

func downloadCommunity(libDir string) {
	downloadFile(libDir + "formats.json", "https://raw.githubusercontent.com/DiSiqueira/DSHouseMaid/master/formats.json")
}

func createLib(libDir string) {
	fmt.Println("No extension database found.")
	if askQuestion("Want to download a community version?") {
		
		downloadCommunity(libDir)
		return
	}
	
	var formats Format
	
	saveLib(formats, libDir)
	return 
}

func loadLib(libDir string) Format {

	os.MkdirAll(libDir, 0777)
	if _, err := os.Stat(libDir + "formats.json"); os.IsNotExist(err) {
		createLib(libDir)
	}

	content, _ := ioutil.ReadFile(libDir + "formats.json")
	
	var formats Format
	json.Unmarshal(content, &formats)

	return formats
}

func main() {

	usr, _ := user.Current()

	libDir := usr.HomeDir + string(filepath.Separator) + ".DSHouseMaid" + string(filepath.Separator)
	fmt.Println(libDir)

	formats := loadLib(libDir)

	directory := flag.String("directory", ".", "Directory to be organized")
	output := flag.String("output", ".", "Main directory to put organized folders")
	link := flag.Bool("link", false, "Create a Symbolic link instead of moving files")

	flag.Parse()

	organize(formats, *output, *directory, *link)
}
