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



func someLatency(idx int, w *sync.WaitGroup, ch chan(int)) { 
    defer w.Done()
	latency := rand.Intn(1500) + 500
	time.Sleep(time.Duration(latency) * time.Millisecond)
	// fmt.Println(idx)
    ch <- idx
    // if idx == 3 {
	// 	close(ch)
	// }
}


func remove(slice []int, s int) []int {
    return append(slice[:s], slice[s+1:]...)
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
			// sort.Slice(targetFiles, func(i, j int) bool { return targetFiles[i].Name() < targetFiles[j].Name() })
			sort.Slice(targetFiles, func(i, j int) bool { return targetFiles[i].ModTime().After(targetFiles[j].ModTime() )})
		}
	}

    // for _, file := range targetFiles {
	// 	fmt.Println(file.Name())
	// }


	readyChan := make(chan int, len(targetFiles))
	min := 0
    buff := []int{}

	for idx, _ := range targetFiles {
		wg.Add(1)
		go someLatency(idx, wg, readyChan) 
	}
	wg.Wait()
	close(readyChan)

	// Возможно запустить цикл в горутине
	for value := range readyChan {
		// fmt.Println(value)
        if value == min {
            fmt.Println("min from channel")
			fmt.Println(min)
            min +=1
		} else {
			buff = append(buff, value)  
            // fmt.Println("min from saved buffer")
		}
	}

	fmt.Println("from buffer: ")
	for _, buff_value := range buff {
        fmt.Println(buff_value)
	}

}