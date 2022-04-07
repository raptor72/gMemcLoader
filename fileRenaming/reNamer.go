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



func someLatency(idx, length int, w *sync.WaitGroup, ch chan(int), done chan(int)) { 
    defer w.Done()
	latency := rand.Intn(2000) + 300
	time.Sleep(time.Duration(latency) * time.Millisecond)
	fmt.Println("Done id: ", idx)
    ch <- idx
    // Здесь по идее надо каунтером высчитывать последний элемент
	// if idx == 8 {
	// 	close(done)
	// }
    if <-done == length {
        close(done)
	}
}


func remove(slice []int, s int) []int {
    newSlise := []int{}
    for _, value := range slice {
		if value == s {
			continue
		} else {
			newSlise = append(newSlise, value)
		}
	}
    return newSlise
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

	targetFiles2 := []int{11,21,13,41,51,16,17,18,29}
    length := len(targetFiles2)

    fmt.Println(length)

	readyChan := make(chan int, len(targetFiles2))
    done := make(chan(int))
	min := 0
    buff := []int{}


	for idx, _ := range targetFiles2 {
		wg.Add(1)
		go someLatency(idx, length, wg, readyChan, done)
	}


	wg2 := new(sync.WaitGroup)

    mu := new(sync.Mutex)

	counter := 0
	wg2.Add(1)
    go func(w2 *sync.WaitGroup, ctr int, mu *sync.Mutex) {
        defer w2.Done()
		for {
			select {
			case msg := <-readyChan:
                if msg == min {
					fmt.Println(msg, true)		
                    mu.Lock()
                    min++
                    mu.Unlock()
				} else {
                    mu.Lock()
					buff = append(buff, msg)
                    mu.Unlock()
				}
				for _, value := range buff {
					if value == min {
                        fmt.Println("Remove from buffer while goroutine wirging", value)
						mu.Lock()
						min++
					    buff = remove(buff, value)
                        mu.Unlock()
					}
				}
                
				
				mu.Lock()
				ctr++
                mu.Unlock()
                fmt.Println("counter:", ctr)
                if ctr == length {
					done <- ctr
                    return
				}
			}
		}
	}(wg2, counter, mu)
	wg.Wait()
	wg2.Wait()
    for _, value := range buff {
        fmt.Println("buffer after goroutine done", value)
	}
}