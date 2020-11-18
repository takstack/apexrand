package apexdb

import (
	//"database/sql"
	//"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" //comment
	"log"
	//"math"
	"time"
)

//Creds holds user login info
type Creds struct {
	Username string
	Pass     string
	Sessid   string
	Exp      time.Time
}

//User holds display info for user status
type User struct {
	Activeusers []Onlineuser
	Teams       []Team
}

//Onlineuser holds pairs of users
type Onlineuser struct {
	Username string
	Online   string
	Handiflt float64
	Handiint int
}

//Insuserfromfile imports the curses
func Insuserfromfile() {
	sl := readsecurefile("apexrand/file/impapexusers.csv")
	//log.Println(sl)
	tx, err := db.Begin() //get connection
	handleError(err)
	//LOG.GL.Println("in batch insert after tx begin")
	qry := "INSERT INTO user (eaddr, username,pass,sess_id,sess_exp,propername,teamassign,psnid) VALUES(?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE username=?, pass=?, propername=?, psnid=?"
	//LOG.GL.Println("in batch insert after qry gotten", qry)
	stmt, err := tx.Prepare(qry)
	handleError(err)

	for _, elem := range sl {
		now := time.Now()
		oldexp := now.AddDate(0, -1, 0)
		_, err = stmt.Exec(elem[3], elem[0], elem[1], "", oldexp, elem[2], 0, elem[4], elem[0], elem[1], elem[2], elem[4])

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

//Updpropername allows users to update their own name
func Updpropername(sessid string, newname string) {
	form, err := db.Prepare("UPDATE user SET propername = ? WHERE sess_id = ?;")
	handleError(err)
	_, err = form.Exec(newname, sessid)
	handleError(err)
	return
}

//Getuser returns the actual username for a players proper name
func Getuser(propername string) string {
	var res string
	qry := fmt.Sprintf("select username from user where propername = '%s';", propername)
	err := db.QueryRow(qry).Scan(&res)
	handleError(err)

	return res
}

//Getuserfromsess returns the actual username for a players proper name
func Getuserfromsess(sessid string) string {
	var res string
	qry := fmt.Sprintf("select username from user where sess_id = '%s';", sessid)
	err := db.QueryRow(qry).Scan(&res)
	handleError(err)

	return res
}

//Getuserfrompsn returns the actual username for a players proper name
func Getuserfrompsn(psnid string) string {
	var res string
	qry := fmt.Sprintf("select username from user where psnid = '%s';", psnid)
	err := db.QueryRow(qry).Scan(&res)
	handleError(err)

	return res
}

//Gethandifromuser returns the handicap for actual username
func Gethandifromuser(username string) int {
	var res int
	qry := fmt.Sprintf("select handicap from user where username = '%s';", username)
	err := db.QueryRow(qry).Scan(&res)
	handleError(err)

	return res
}

//Sethandifromuser returns the handicap for actual username
func Sethandifromuser(username string, handicap int) {
	form, err := db.Prepare("UPDATE user SET handicap= ? WHERE username =?;")
	handleError(err)
	_, err = form.Exec(handicap, username)
	handleError(err)
	log.Println("set handicap for:", username)
	return

}

//Getproper returns the actual username for a players proper name
func Getproper(username string) string {
	var res string
	qry := fmt.Sprintf("select propername from user where username = '%s';", username)
	err := db.QueryRow(qry).Scan(&res)
	handleError(err)

	return res
}

//Getactiveusers will return []string of mode requested
func Getactiveusers() []Onlineuser {
	now := time.Now()
	qry := "select propername,sess_exp,handicap from user where sess_exp > ? and CHAR_LENGTH(sess_id)>5 "
	res, err := db.Query(qry, now)
	handleError(err)
	var sl []Onlineuser
	for res.Next() {
		var u Onlineuser
		var t time.Time
		// for each row, scan the result into our tag composite object
		err := res.Scan(&u.Username, &t, &u.Handiint)
		handleError(err)

		u.Online = isonline(t)
		if u.Online == "" {
			Assignteam(Getuser(u.Username), 0) //remove from team assignment if offline
		}
		if u.Handiint == 0 {
			u.Handiflt = float64(0)
		} else {
			u.Handiflt = float64(u.Handiint) / 100
		}

		sl = append(sl, u)
	}
	res.Close()
	return sl
}
func isonline(exp time.Time) string {
	now := time.Now()
	d1 := now.AddDate(0, 1, 0)
	d1 = d1.Add(-time.Hour / 2)

	if exp.After(d1) {
		return "Online"
	}
	return ""
}

//Seluser will return user session info
func Seluser(u string) Creds {
	var res Creds
	qry := fmt.Sprintf("select username, pass, sess_id, sess_exp from user where username = '%s';", u)
	err := db.QueryRow(qry).Scan(&res.Username, &res.Pass, &res.Sessid, &res.Exp)
	handleError(err)

	return res
}

//Updlogin updates user/pass for given sessid
func Updlogin(username string, pass string, sessid string) {
	form, err := db.Prepare("UPDATE user SET username = ? WHERE sess_id = ?;")
	handleError(err)
	_, err = form.Exec(username, sessid)
	handleError(err)

	form, err = db.Prepare("UPDATE user SET pass = ? WHERE sess_id = ?;")
	handleError(err)
	_, err = form.Exec(pass, sessid)
	handleError(err)
	log.Println("user/pass updated:", username, pass)
	return
}
