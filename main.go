package main

import (
    "io"
    "os"
	"fmt"
    "log"
    // "io/ioutil"
    "strings"
    "compress/gzip"
	"github.com/bradfitz/gomemcache/memcache"	
)



func main() {
	mc := memcache.New("127.0.0.1:11211")
    fmt.Println(mc)

    nBytes, nChunks := int64(0), int64(0)
    file, err := os.Open("20170929000300.tsv.gz")
    if err != nil {
        fmt.Println(err)
    }
    defer file.Close()
    ff := "ап"
    fmt.Println(len([]byte(ff)))

	zipReader, err := gzip.NewReader(file)
	if err != nil {
	  fmt.Println(err)
	}
    defer zipReader.Close()

    buf := make([]byte, 0, 4*1024)

    // corrected_bytes_len := 0

    for {
        n, err := zipReader.Read(buf[:cap(buf)])
        // n = n + corrected_bytes_len

        buf = buf[:n]

        // fmt.Println(string(buf))
        smass := strings.Split(string(buf), "\n")
        fmt.Println(smass)
        strings_in_batch := len(smass)
        last_string := smass[strings_in_batch - 1]
        fmt.Println(last_string)
        fmt.Println(len([]byte(last_string)))
        fmt.Println("#########################################")
        // corrected_bytes_len = len([]byte(last_string))
//        }
//        fmt.Printf("%T", buf)
//        s, err := reader.ReadString('\n')
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


    
	// s, err := ioutil.ReadAll(zipReader)
    // if err != nil {
    //     fmt.Println(err)
    // }
    // fmt.Println(string(s))   


}