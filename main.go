package main

import (
	// "bytes"
    "flag"
	"fmt"
	"io"
	"log"
	"os"

	"io/ioutil"
	"compress/gzip"
	"strings"
    "sync"

	"github.com/bradfitz/gomemcache/memcache"
)


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
        // fmt.Println("head!!!")
		// fmt.Println(string(head))
		head = append(head, []byte(smass[0])...)
        cacher(head, mc)
        starter += 1
	}
    for starter < strings_in_batch - 1 {
        cacher( []byte(smass[starter]), mc)
        starter ++
	}
	return []byte(smass[strings_in_batch - 1])
}


func fileProcessor(fileName string, memcacheClient *memcache.Client, w *sync.WaitGroup) {
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

	buf := make([]byte, 0, 4*1024*1024)

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
        // fmt.Println("#########################################")
        if err != nil && err != io.EOF {
            log.Fatal(err)
        }
    }
    log.Println("Prosessed:", fileName, "Bytes:", nBytes, "Chunks:", nChunks)
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
    wg := new(sync.WaitGroup)
	for _, file := range filesFromDir {
        // if !strings.HasPrefix(file.Name(), ".") && strings.HasSuffix(file.Name(), ".tsv.gz") && file.Name() == "20170929000300.tsv.gz" {
		if !strings.HasPrefix(file.Name(), ".") && strings.HasSuffix(file.Name(), ".tsv.gz") {
			// Проходим по всем найденным файлам и печатаем их имя и размер
            wg.Add(1)
			go fileProcessor(file.Name(), mc, wg)
			fmt.Printf("name: %s, size: %d\n", file.Name(), file.Size())
		}
	}
    wg.Wait()
}