package main

import (
	// "bytes"
	"fmt"
	"io"
	"log"
	"os"

	// "io/ioutil"
	"compress/gzip"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)


func cacher(buf []byte, mc *memcache.Client) {
	s := strings.Split(string(buf), "\n")
	for _, st := range s {
		words := strings.Fields(st)
		if len(words) > 1 {
			key := words[0] + ":" + words[1]
			value := strings.Join( words[2:], ",")
			fmt.Println(key)
			fmt.Println(value)
			mc.Set(&memcache.Item{Key: key, Value: []byte(value)})
		}
	}
}


func bufer_handler(head []byte, chank []byte, mc *memcache.Client) []byte {
	smass := strings.Split(string(chank), "\n")
	strings_in_batch := len(smass)
	// last_string := smass[strings_in_batch - 1]
    starter := 0
	if len(head) != 0 {
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


func main() {
	mc := memcache.New("127.0.0.1:11211")
    fmt.Printf("%T\n", mc)

    nBytes, nChunks := int64(0), int64(0)
    file, err := os.Open("20170929000300.tsv.gz")
    if err != nil {
        fmt.Println(err)
    }
    defer file.Close()

	zipReader, err := gzip.NewReader(file)
	if err != nil {
	  fmt.Println(err)
	}
    defer zipReader.Close()

	buf := make([]byte, 0, 4*1024)

	head := []byte{}

	for {
        n, err := zipReader.Read(buf[:cap(buf)])
        // n = n + corrected_bytes_len

        buf = buf[:n]

        // fmt.Println(string(buf))

        head = bufer_handler(head, buf, mc)

		// smass := strings.Split(string(buf), "\n")
		// cacher(buf, mc)
		// strings_in_batch := len(smass)
        // last_string := smass[strings_in_batch - 1]
		// fmt.Printf("%T\n", last_string)
		// fmt.Println(last_string)
        // fmt.Println(len([]byte(last_string)))
        fmt.Println("#########################################")

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

        // buf = buf[:n]
        // process buf
        if err != nil && err != io.EOF {
            log.Fatal(err)
        }
    }
    log.Println("Bytes:", nBytes, "Chunks:", nChunks)

}