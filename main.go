package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
)


type Tracker struct {
	Key, Uuid, Lat, Lon string
    Tail []string
}


func parseBuff(buf []byte) ([]Tracker, int, int) {
	tracks := []Tracker{}
	goodCounter, errCounter := 0,0
	s := strings.Split(string(buf), "\n")
	for _, st := range s {
		words := strings.Fields(st) 
        if len(words) < 5 {
			errCounter ++
            break
		}
        strLat := words[2]
        strLon := words[3]

		lat, err := strconv.ParseFloat(strLat, 64) 
		if err != nil {
			errCounter ++
            break
		}

		lon, err := strconv.ParseFloat(strLon, 64) 
		if err != nil {
			errCounter ++
            break
		}	

		if (lat < -180 || lat > 180) || (lon < -180 || lon > 180) {
			errCounter ++
			break
		}

		track := Tracker{words[0], words[1], strLat, strLon, words[4:]}
        tracks = append(tracks, track)
        goodCounter ++
	}
	return tracks, goodCounter, errCounter
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


func cacher(tracks []Tracker, mGrid map[string]*memcache.Client) {
	for _, track := range tracks {
		var sb strings.Builder
		sb.WriteString(track.Key)
		sb.WriteString(":")
		sb.WriteString(track.Uuid)
        tail := []string{}
        tail = append(tail, track.Lat)
        tail = append(tail, track.Lon)
        tail = append(tail, track.Tail...)
		value := strings.Join(tail, ",")
        mGrid[track.Key].Set(&memcache.Item{Key: sb.String(), Value: []byte(value)})
	}
}


func prefix(f os.FileInfo, prefix, where string) {
	var b strings.Builder
    b.WriteString(prefix)
    b.WriteString(f.Name())
    os.Rename(f.Name(), b.String())
    switch where {
	case "current":
		fmt.Printf("Prefixed current handling file %v\n", f.Name())
    case "while":
		fmt.Printf("Prefixed file %v from buffer while goroutine working\n", f.Name())
	default:
		fmt.Printf("Prefixed file %v from buffer after goroutine done\n", f.Name())
	}
}


func buferHandler(head []byte, chank []byte, mGrid	map[string]*memcache.Client) ([]byte, int, int, int) {
	smass := strings.Split(string(chank), "\n")
	strings_in_batch := len(smass)
    var starter, goodValues, Errors int
	if len(head) != 0 {
		head = append(head, []byte(smass[0])...) // Здесь слепляем полноценный chank
        track, goodCounter, errCounter := parseBuff(head)
		cacher(track, mGrid)
        starter ++
		goodValues += goodCounter
		Errors += errCounter
	}
    for starter < strings_in_batch - 1 {
        track, goodCounter, errCounter := parseBuff([]byte(smass[starter]))
		cacher(track, mGrid)
        starter ++
		goodValues += goodCounter
		Errors += errCounter
	}
	return []byte(smass[strings_in_batch - 1]), starter, goodValues, Errors
}


func fileProcessor(fileName string, mGrid map[string]*memcache.Client, w *sync.WaitGroup, ch, done chan(int), idx int) {
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
	var numAll, numGood, numErr int

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

        curHead, curAll, curGood, curErr := buferHandler(head, buf, mGrid)
        head = curHead
		numAll += curAll
		numGood += curGood
		numErr += curErr

		if nChunks % 8000 == 0 {
			log.Printf("%v Chunks and %v values done for file %v\n", nChunks, numAll, fileName)
		}

		if err != nil && err != io.EOF {
            log.Fatal(err)
        }
    }
	ch <- idx
	log.Println("Prosessed file:", fileName, "Bytes:", nBytes, "Chunks:", nChunks, 
	            "AllValues:", numAll, "Good Values:", numGood, "Err values:", numErr)
	_, ok := <-done 
    if ok {
        close(done)
	}
}


func main() {
    flushAll := flag.Bool("flushAll", true, "Drop all cached values before the program start")
    maxConns := flag.Int("maxConns", 15, "Number of max idle connections for each client instance")
	Adid := flag.String("adid", "11211", "port for storing info about adid klient device")
    Dvid := flag.String("dviv", "11211", "port for storing info about dvid klisnt device")
    Gaid := flag.String("gaid", "11211", "port for storing info about gaid client device")
    Idfa := flag.String("idfa", "11211", "port for storing info about idfa client device")
	flag.Parse()        
	mGrid := make(map[string]*memcache.Client)
	ss := new(memcache.ServerList)
    for _, name_and_port := range [][]string{ 
		[]string{"adid", *Adid}, 
		[]string{"dvid", *Dvid}, 
		[]string{"gaid", *Gaid}, 
		[]string{"idfa", *Idfa}}{
		err := ss.SetServers("127.0.0.1:" + name_and_port[1])
        if err != nil {
            log.Fatalf("Coud not set memcache server on address 127.0.0.1 and port %v", name_and_port[1])
		}
        c := memcache.NewFromSelector(ss) // type of c is: *memcache.Client

		err = c.Ping()
		if err != nil {
            log.Fatalf("Coud not connect to memcache instance on 127.0.0.1 and port %v", name_and_port[1])
		}

		c.MaxIdleConns = *maxConns
	    if *flushAll {
            c.FlushAll()
    	}

		mGrid[name_and_port[0]] = c
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
    log.Println("Start caching files: ")
	for _, file := range targetFiles {
		fmt.Printf("%v ", file.Name())
	}
    fmt.Println()

	readyChan := make(chan int, len(targetFiles)) // канал для обмена значениями между горитнами воркерами и горутиной буффером
    done := make(chan(int)) // канал для завершения работы горутины - буффера
	var min int  // указатель на текущее минимальное значение
	var counter int // количество отработанных горитин из числа fileCount
	buff := []int{} 

	caching_group := new(sync.WaitGroup)
	for idx, file := range targetFiles {
		caching_group.Add(1)
		go fileProcessor(file.Name(), mGrid, caching_group, readyChan, done, idx)
	}

	// Здесь распологаем конкурентный буффер
	buffer_group := new(sync.WaitGroup)
    mu := new(sync.Mutex)

	go func(w2 *sync.WaitGroup, ctr int, mu *sync.Mutex) {
		defer w2.Done()
		for msg := range readyChan {
            if msg == min {	
                prefix(targetFiles[msg], ".", "current")
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
					prefix(targetFiles[value], ".", "while")
					mu.Lock()
					min++
					buff = remove(buff, value)
                    mu.Unlock()
				}
			}
			mu.Lock()
			ctr++
            mu.Unlock()
			if ctr == len(targetFiles) {
				done <- ctr
			}
		}
	}(buffer_group, counter, mu)
	caching_group.Wait()
	buffer_group.Wait()

    sort.Ints(buff)
	for _, value := range buff {
		prefix(targetFiles[value], ".", "") // %T of targetFiles[value] is *os.fileStat
	}

}