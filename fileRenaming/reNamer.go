package main

import (
    "os"
	"fmt"
	"io/ioutil"
	"log"
    "strings"
    "sort"
    "sync"
	"time"
    "math/rand"
)



func someLatency(Name string, w *sync.WaitGroup) { 
    defer w.Done()
	latency := rand.Intn(500) + 500
	time.Sleep(time.Duration(latency) * time.Millisecond)
	fmt.Println(Name)
}



func main() {
	filesFromDir, err := ioutil.ReadDir("../")
	if err != nil {
		log.Fatal(err)
	}
    // fmt.Println(filesFromDir)
    var targetFiles []os.FileInfo
    wg := new(sync.WaitGroup)
	for _, file := range filesFromDir {
        // if !strings.HasPrefix(file.Name(), ".") && strings.HasSuffix(file.Name(), ".tsv.gz") && file.Name() == "20170929000300.tsv.gz" {
		if !strings.HasPrefix(file.Name(), ".") && strings.HasSuffix(file.Name(), ".tsv.gz") {
            targetFiles = append(targetFiles, file)
			// log.Printf("Type: %T, name: %s, size: %d\n", file, file.Name(), file.Size())
			sort.Slice(targetFiles, func(i, j int) bool { return targetFiles[i].Name() < targetFiles[j].Name() })
			// sort.Slice(targetFiles, func(i, j int) bool { return targetFiles[i].ModTime().After(targetFiles[j].ModTime() )})
		}
	}
	for _, file := range targetFiles {
		wg.Add(1)
		go someLatency(file.Name(), wg)
	}
	wg.Wait()
}