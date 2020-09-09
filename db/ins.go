package apexdb

import (
	//"encoding/csv"
	"database/sql"
	"io"
	"log"
	"qrw"
	//"fmt"
	_ "github.com/go-sql-driver/mysql" //comment
	"keys"
	//LOG "logger"
	"os"
	"strings"
	//"quotes/format"
	"strconv"
	//"sync"
	"errors"
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
	qry := qryins("var")
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

//Insuserfromfile imports the curses
func Insuserfromfile() {
	sl := readsecurefile("impapexusers.csv")
	//log.Println(sl)
	tx, err := db.Begin() //get connection
	handleError(err)
	//LOG.GL.Println("in batch insert after tx begin")
	qry := qryins("user")
	//LOG.GL.Println("in batch insert after qry gotten", qry)
	stmt, err := tx.Prepare(qry)
	handleError(err)

	for _, elem := range sl {
		now := time.Now()
		oldexp := now.AddDate(0, -1, 0)
		_, err = stmt.Exec(elem[0], elem[1], "", oldexp, elem[2], 0, elem[0], elem[1], elem[2])

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

//Inscursefromfile imports the curses
func Inscursefromfile() {
	sl := readfile("importcurse.csv")
	//log.Println(sl)
	tx, err := db.Begin() //get connection
	handleError(err)
	//LOG.GL.Println("in batch insert after tx begin")
	qry := qryins("curse")
	//LOG.GL.Println("in batch insert after qry gotten", qry)
	stmt, err := tx.Prepare(qry)
	handleError(err)

	for _, elem := range sl {
		_, err = stmt.Exec(elem[1], elem[0], elem[2], elem[3], elem[2])
		log.Println("inscurse elem[0],elem[2]", elem[0], elem[2])
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

//Logcurse enters all curses into db
func Logcurse(cid int, player int, team int, assigntype string, rollcounter int) {
	pid := player
	switch assigntype {
	case "player":
		if team == 2 {
			pid = player + 3
		}
	case "team":
		pid = team * 10
	}

	now := time.Now()
	form, err := db.Prepare("INSERT INTO assign(curse_id, player_id, adate,roll) VALUES (?,?,?,?)")
	if err != nil {
		panic(err.Error())
	}
	form.Exec(cid, pid, now, rollcounter)
}

//Loggame will enter data from games
func Loggame(player string, dmg string, place string) error {
	username := Getuser(player) //get username from current proper name
	d, err := strconv.Atoi(dmg)
	if err != nil {
		log.Println("Error: strconv.Atoi(dmg)", err)
		return errors.New("Error parsing damage")
	}

	p, err := strconv.Atoi(place)
	handleError(err)
	now := time.Now()
	pdmg := 0

	//determine if place should award additional damage
	switch {
	case p > 30 || p < 1 || d < 0:
		log.Println("loggame parameters out of bounds:", p, d)
		return errors.New("FUCKING IDIOT - parameters out of bounds")
	case p == 1:
		pdmg = 500
	case p < 4 && p > 1:
		pdmg = 200
	}
	tdmg := pdmg + d
	form, err := db.Prepare("INSERT INTO games(username,dmg,place,placedmg,totaldmg,gametime) VALUES (?,?,?,?,?,?)")
	if err != nil {
		panic(err.Error())
	}
	_, err = form.Exec(username, d, p, pdmg, tdmg, now)
	if err != nil {
		panic(err.Error())
	}
	return nil
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
func qryins(mode string) string {
	var s string
	switch mode {
	case "var":
		s = "INSERT INTO var (descrip,cat) VALUES(?,?) ON DUPLICATE KEY UPDATE descrip=?"
		return s
	case "curse":
		s = "INSERT INTO curse (id,descrip,adjfactor,assigntype) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE adjfactor=?"
		return s
	case "user":
		s = "INSERT INTO user (username,pass,sess_id,sess_exp,propername,teamassign) VALUES(?,?,?,?,?,?) ON DUPLICATE KEY UPDATE username=?, pass=?, propername=?"
		return s
	}
	return s
}
func handleError(err error) {
	if err != nil {
		log.Println(err)
	}
}
