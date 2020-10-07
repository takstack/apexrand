package qrw

import (
	//"fmt"
	//"strconv"
	//"strings"

	"bufio"
	"log"
	//"messingaround/folder1"
	//"bytes"
	"encoding/csv"
	//"io"
	"os"
	"path/filepath"
	//"sync"
	//"time"
	//LOG "logger"
)

//StartCSVwriter returns csv writer
func StartCSVwriter(f *os.File) *csv.Writer {
	CSVw := csv.NewWriter(f)
	log.Println("CSV writer started")
	return CSVw
}

//Startbufwriter returns buf writer
func Startbufwriter(f *os.File) *bufio.Writer {
	w := bufio.NewWriter(f)
	log.Println("bufio writer started")
	return w
}

//Openwritefile opens and clears file for writing
func Openwritefile(s string) *os.File {
	path := os.Getenv("GOPATH")
	f, err := os.OpenFile(filepath.Join(path, "/src/", s), os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	if err != nil {
		log.Println("Error opening file:", err)
		//os.Exit(1)
	}
	return f
}
