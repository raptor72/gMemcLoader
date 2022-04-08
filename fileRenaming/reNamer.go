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
    // функция иммитирующая загрузку файлов в мемкеш. Добавляет произвольную задержку на каждое выполнение
	defer w.Done()
	latency := rand.Intn(2000) + 300
	time.Sleep(time.Duration(latency) * time.Millisecond)
	fmt.Println("Done id: ", idx)
    ch <- idx
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

    // fmt.Println("length: ", length)

	readyChan := make(chan int, len(targetFiles2)) // канал для обмена значениями между горитнами воркерами и горутиной буффером
    done := make(chan(int)) // канал для завершения работы горутины - буффера
	var min int  // указатель на текущее минимальное значение
	var counter int // количество отработанных горитин из числа length
	buff := []int{} // буфер куда складываются отработавшие значения превышающие минимальное

	for idx := range targetFiles2 {
		wg.Add(1)
		go someLatency(idx, length, wg, readyChan, done)
	}

	buffer_group := new(sync.WaitGroup)
    mu := new(sync.Mutex)

	buffer_group.Add(1)
    go func(w2 *sync.WaitGroup, ctr int, mu *sync.Mutex) {
        // функция конкурентный буффер. Если из канала получается минимальное значение оно сразу обрабатывается
		// в данном случае печатается. Иначе значение добавляется в буфер
        // затем буфер проверяется на наличие в нем текущего минимального значения и в случае его присутствия оно 
        // так же будет обработано
		defer w2.Done()
		for msg := range readyChan {
            if msg == min {
				fmt.Println("Current min value done", msg)		
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
                    fmt.Println("Remove from buffer while goroutine working", value)
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
	}(buffer_group, counter, mu)
	wg.Wait()
	buffer_group.Wait()
    for _, value := range buff {
        fmt.Println("buffer after goroutine done", value)
	}
}