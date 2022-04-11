package main

import (
	// "bytes"
    "flag"
	"fmt"
	"io"
	"log"
	"os"
    "sort"
	"io/ioutil"
	"compress/gzip"
	"strings"
    "sync"

	"github.com/bradfitz/gomemcache/memcache"
)

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

func cacher(buf []byte, mc *memcache.Client) {
	s := strings.Split(string(buf), "\n")
	for _, st := range s {
		words := strings.Fields(st)
		if len(words) > 1 {
			key := words[0] + ":" + words[1]
			value := strings.Join( words[2:], ",")
			// fmt.Println(key)
			// fmt.Println(value)
			mc.Set(&memcache.Item{Key: key, Value: []byte(value)})
		}
	}
}


func buferHandler(head []byte, chank []byte, mc *memcache.Client) []byte {
	smass := strings.Split(string(chank), "\n")
	strings_in_batch := len(smass)
    starter := 0
	if len(head) != 0 {
		head = append(head, []byte(smass[0])...)
        go cacher(head, mc)
        starter += 1
	}
    for starter < strings_in_batch - 1 {
        go cacher( []byte(smass[starter]), mc)
        starter ++
	}
	return []byte(smass[strings_in_batch - 1])
}

// func fileProcessor(fileName string, memcacheClient *memcache.Client, w *sync.WaitGroup, ch chan(int), done chan(int), idx, length int) {
func fileProcessor(fileName string, memcacheClient *memcache.Client, w *sync.WaitGroup, ch chan(int), done chan(int), idx int) {
	defer w.Done()
 
	nBytes, nChunks := int64(0), int64(0)
    file, err := os.Open(fileName)
    if err != nil {
        fmt.Println(err)
    }
    defer file.Close()

	zipReader, err := gzip.NewReader(file)
	if err != nil {
	  fmt.Println(err)
	}
    defer zipReader.Close()

	buf := make([]byte, 0, 8*1024*1024)

	head := []byte{}

	for {
        n, err := zipReader.Read(buf[:cap(buf)])

        buf = buf[:n]

        if n == 0 {
            if err == nil {
                continue
            }
            if err == io.EOF {
                break
            }
            log.Fatal(err)
        }
        nChunks++
        nBytes += int64(len(buf))

        head = buferHandler(head, buf, memcacheClient)

		if err != nil && err != io.EOF {
            log.Fatal(err)
        }
    }
	ch <- idx
	log.Println("Prosessed file:", fileName, "Bytes:", nBytes, "Chunks:", nChunks)
	_, ok := <-done 
    if ok {
        close(done)
		// return
	}
}


func main() {
    flushAll := flag.Bool("flushAll", true, "Drop all cached values before the program start")
	flag.Parse()
	mc := memcache.New("127.0.0.1:11211")
	if *flushAll {
        mc.FlushAll()
	}

	filesFromDir, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}


    var targetFiles []os.FileInfo
	for _, file := range filesFromDir {
		if !strings.HasPrefix(file.Name(), ".") && strings.HasSuffix(file.Name(), ".tsv.gz") {
        targetFiles = append(targetFiles, file)
		}
	}
    // Здесь сортируем targetFiles
    sort.Slice(targetFiles, func(i, j int) bool { return targetFiles[i].Name() < targetFiles[j].Name() })
    fmt.Println(targetFiles)

	readyChan := make(chan int, len(targetFiles)) // канал для обмена значениями между горитнами воркерами и горутиной буффером
    done := make(chan(int)) // канал для завершения работы горутины - буффера
	var min int  // указатель на текущее минимальное значение
	var counter int // количество отработанных горитин из числа fileCount
	buff := []int{} 

	caching_group := new(sync.WaitGroup)
	for idx, file := range targetFiles {
		caching_group.Add(1)
		// go fileProcessor(file.Name(), mc, wg, readyChan, idx, fileCount)
		go fileProcessor(file.Name(), mc, caching_group, readyChan, done, idx)
	}

	// Здесь распологаем конкурентный буффер
	buffer_group := new(sync.WaitGroup)
    mu := new(sync.Mutex)

	go func(w2 *sync.WaitGroup, ctr int, mu *sync.Mutex) {
		defer w2.Done()
		for msg := range readyChan {
            if msg == min {	
                fmt.Println("Prefixed current file: ", targetFiles[msg].Name())
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
                    fmt.Println("Prefixed file from buffer while goroutine working: ", targetFiles[value].Name())
					mu.Lock()
					min++
					buff = remove(buff, value)
                    mu.Unlock()
				}
			}
			mu.Lock()
			ctr++
            mu.Unlock()
            // fmt.Println("counter:", ctr)
            // fmt.Println(fileCount)
			if ctr == len(targetFiles) {
				done <- ctr
				// close(done)
				// close(readyChan)
                // return
			}
		}
	}(buffer_group, counter, mu)
	caching_group.Wait()
	buffer_group.Wait()

	for _, value := range buff {
        fmt.Println("Prefixed buffer after goroutine done", targetFiles[value].Name())
	}

	// for _, file := range filesFromDir {
	// 	if !strings.HasPrefix(file.Name(), ".") && strings.HasSuffix(file.Name(), ".tsv.gz") {
	// 		// Здесь нужно применить сортировку - по имени или по времени
	// 		wg.Add(1)
	// 		go fileProcessor(file.Name(), mc, wg)
	// 		log.Printf("name: %s, size: %d\n", file.Name(), file.Size())
	// 	}
	// }
    // wg.Wait()
}