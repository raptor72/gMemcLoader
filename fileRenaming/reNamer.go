package main

import (
    "os"
	"fmt"
	"io/ioutil"
	"log"
    "strings"
    "sort"
)



func main() {
	filesFromDir, err := ioutil.ReadDir("../")
	if err != nil {
		log.Fatal(err)
	}
    fmt.Println(filesFromDir)
    var targetFiles []os.FileInfo

	for _, file := range filesFromDir {
        // if !strings.HasPrefix(file.Name(), ".") && strings.HasSuffix(file.Name(), ".tsv.gz") && file.Name() == "20170929000300.tsv.gz" {
		if !strings.HasPrefix(file.Name(), ".") && strings.HasSuffix(file.Name(), ".tsv.gz") {
            targetFiles = append(targetFiles, file)
			log.Printf("Type: %T, name: %s, size: %d\n", file, file.Name(), file.Size())
		}
	}
    fmt.Println(targetFiles)
	// sort.Slice(targetFiles, func(i, j int) bool { return targetFiles[i].Name() < targetFiles[j].Name() })
	sort.Slice(targetFiles, func(i, j int) bool { return targetFiles[i].ModTime().After(targetFiles[j].ModTime() )})
	fmt.Println(targetFiles)
}