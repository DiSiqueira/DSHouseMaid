package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"os"
	"flag"
    "sync"
    "encoding/json"
)

type Format struct {
	Items []struct {
		Items []struct {
			Name string `json:"Name"`
		} `json:"Items"`
		Name string `json:"Name"`
	} `json:"Items"`
}

func move (oldfile, newFile string, link bool) error {
	if (link) {
		return os.Link(oldfile, newFile)
	}
		
	return os.Rename(oldfile, newFile)
}

func organize(formats Format, output, input string, link bool) {
	files, _ := ioutil.ReadDir(input)
	
    var wg sync.WaitGroup
	wg.Add(len(files))	
	
	for _, f := range files {
		
		go func(file string){
			
			ext := filepath.Ext(file)
			
			for _, folder := range formats.Items {
				for _, extension := range folder.Items {
					
					if (ext == extension.Name) {						
						os.MkdirAll(output + string(filepath.Separator) + folder.Name,0777)
				
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

func loadLib() Format {

	if _, err := os.Stat("~/.DSHouseMaid/formats.json"); os.IsNotExist(err) {

	}
	
	content, err := ioutil.ReadFile("formats.json")
    if err!=nil{
        fmt.Print("Error:",err)
    }
	var formats Format
	json.Unmarshal(content, &formats)
	 
}

func main() {
	
	formats := loadLib()
	
	directory := flag.String("directory", ".", "Directory to be organized")
	output := flag.String("output", ".", "Main directory to put organized folders")
	link := flag.Bool("link", false, "Create a Symbolic link instead of moving files")
	
	flag.Parse()
	
	organize(formats, *output, *directory, *link)
}
