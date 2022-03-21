package main

import (
    "os"
	"fmt"
    "io/ioutil"
	"compress/gzip"
	"github.com/bradfitz/gomemcache/memcache"	
)

func main() {
	mc := memcache.New("127.0.0.1:11211")
    fmt.Println(mc)

	file, err := os.Open("20170929000300.tsv.gz")
    if err != nil {
        fmt.Println(err)
    }
    defer file.Close()

	fz, err := gzip.NewReader(file)
	if err != nil {
	  fmt.Println(err)
	}
    defer fz.Close()

	s, err := ioutil.ReadAll(fz)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(string(s))   


}