package qrw

import (
	"fmt"
	//"strconv"
	//"strings"
	"bufio"
	"io/ioutil"
	"log"
	//"messingaround/folder1"
	//"bytes"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"sync"
	//"time"
	//LOG "logger"
)

//StartCSVreader opens bufio reader and returns csv reader
func StartCSVreader(f *os.File) *csv.Reader {
	r := bufio.NewReader(f)
	CSVr := csv.NewReader(r)
	return CSVr
}

//Startbufscanner opens bufio scanner
func Startbufscanner(f *os.File) *bufio.Scanner {
	r := bufio.NewScanner(f)
	r.Split(bufio.ScanLines)
	return r
}

//Startioreadall reads and returns res
func Startioreadall(f *os.File) []byte {
	res, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("Error reading file:", f)
	}
	return res
}

//StreamCSVreader reads reader and sends to csv channel
func StreamCSVreader(CSVr *csv.Reader, c1 chan<- []string, wg *sync.WaitGroup) {
	i := 0
	firstline, err := CSVr.Read()
	if err != nil {
		log.Println("error CSV read: ", i)
	}
	log.Println("First CSV line discarded: ", firstline)
	for {
		record, err := CSVr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		//log.Println("reader send:", i, record)
		c1 <- record
		i++
	}
	log.Println("Completed send items:", i-1)
	close(c1)
	log.Println("streamcsvreader: c1 channel closed")
}

//PrintCSVreader prints to screen from CSV reader
func PrintCSVreader(CSVr *csv.Reader) {
	i := 0
	for {
		record, err := CSVr.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Println("i: ", i, record)
		i++

	}
}

//Printtofile will write []string to file
func Printtofile(sl []string) {
	f, err := os.OpenFile("files/sym.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	for elem := range sl {
		f.Write([]byte(sl[elem]))
		f.Write([]byte("\r\n"))
	}
}

//Getreadfile opens file for reading without clearing, 0 for include gopath, anything for don't include
func Getreadfile(fs string, pathselect int) *os.File {
	//dbtest.WriteTest()
	var path string
	if pathselect == 0 {
		path = os.Getenv("GOPATH") + "/src/"
	} else {
		path = ""
	}
	f, err := os.Open(filepath.Join(path, fs))

	if err != nil {
		log.Println(err)
	}
	return f
}
