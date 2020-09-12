package apexdb

import (
	//"encoding/csv"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" //comment
	"io"
	"keys"
	"log"
	"qrw"
	//LOG "logger"
	"os"
	"strings"
	//"quotes/format"
	"time"
)

var db *sql.DB

func init() {
	str := key.Getkeys("Mysql")
	var err error
	db, err = sql.Open("mysql", strings.Join([]string{str[1], ":", str[2], str[3], "/apexdb?parseTime=true"}, ""))
	handleError(err)
	db.SetMaxOpenConns(11)
	db.SetMaxIdleConns(11)
	db.SetConnMaxLifetime(time.Second * 11 * 3)
	log.Println("Database Opened")
	f, err := os.OpenFile("file/sqlerr.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err)
	}
	f.Close()

}

//Insvarfromfile imports the current game vars
func Insvarfromfile() {
	sl := readfile("importvar.csv")
	//log.Println(sl)
	tx, err := db.Begin() //get connection
	handleError(err)
	//LOG.GL.Println("in batch insert after tx begin")
	qry := "INSERT INTO var (descrip,cat) VALUES(?,?) ON DUPLICATE KEY UPDATE descrip=?"
	//LOG.GL.Println("in batch insert after qry gotten", qry)
	stmt, err := tx.Prepare(qry)
	handleError(err)

	for _, elem := range sl {
		_, err = stmt.Exec(elem[0], elem[1], elem[0])

		if err != nil {
			log.Println("DB Error on this row: ", elem)
			tx.Rollback()
			handleError(err)
		}

	}
	err = tx.Commit()
	if err != nil {
		log.Fatalln("Commit Error")
	}
}

//Selvars will return []string of mode requested
func Selvars(mode string) []string {
	qry := getvarqry(mode)
	res, err := db.Query(qry)
	handleError(err)
	var sl []string
	for res.Next() {
		var s string
		// for each row, scan the result into our tag composite object
		err := res.Scan(&s)
		handleError(err)

		sl = append(sl, s)
	}
	res.Close()
	return sl
}
func getvarqry(mode string) string {
	qry := fmt.Sprintf("SELECT descrip from var WHERE cat='%s';", mode)
	return qry
}
func readfile(s string) [][]string {
	r := qrw.StartCSVreader(qrw.Getreadfile("apexrand/file/"+s, 0))
	i := 0
	var resSL [][]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		resSL = append(resSL, record)
		i++
	}
	log.Printf("read %d items:", i)
	return resSL
}
func readsecurefile(s string) [][]string {
	r := qrw.StartCSVreader(qrw.Getreadfile(s, 0))
	i := 0
	var resSL [][]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		resSL = append(resSL, record)
		i++
	}
	log.Printf("read %d items:", i)
	return resSL
}
func handleError(err error) {
	if err != nil {
		log.Println(err)
	}
}
