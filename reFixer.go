package main

import (
    "os"
	"log"
	"fmt"
	"flag"
	"io/ioutil"
    "strings"
)


func main() {
	prefix := flag.String("prefix", ".", "prefix shiuld be deleted from file names")
    suffix := flag.String("suffix", ".tsv.gz", "suffix to catch target files")
    dir := flag.String("dir", ".", "directpory to finding files")
	flag.Parse()
    fmt.Printf("rename files from directory %v started from %v which has suffix %v\n", *dir, *prefix, *suffix)
	filesFromDir, err := ioutil.ReadDir(*dir)
	if err != nil {
		log.Fatal(err)
	}
    renamed_files := []string{}
	for _, file := range filesFromDir {
		if strings.HasPrefix(file.Name(), *prefix) && strings.HasSuffix(file.Name(), *suffix) {
			renamed_files = append(renamed_files, file.Name())
			newName := strings.TrimPrefix(file.Name(), *prefix)      
    		osErr := os.Rename(file.Name(), newName)
			if osErr != nil {
			    log.Fatal(osErr)
		    }
		}
	}
    fmt.Printf("Renamed files: %v\n", renamed_files)
}