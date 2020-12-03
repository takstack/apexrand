package apexdb

import (
	_ "github.com/go-sql-driver/mysql" //comment
	//"strconv"
	//"sync"
	//"apexrand/qrw"
	"encoding/json"
	//"errors"
	"log"
	//"sort"
	//"database/sql"
	"fmt"
	"time"
)

//Cats are tracker categories, initialized in init
type Cats struct {
	Cat1 string
	Cat2 string
	Cat3 string
}

/*
//Chars are allowed characters, initialized in init
type Chars struct {
	Char1 string
	Char2 string
	Char3 string
	Char4 string
}
*/

//Apimain exp
type Apimain struct {
	Apiseries     []Apigames
	Timesincepull time.Duration
	Timeselect    time.Duration
}

//Apigames exp
type Apigames struct {
	UID         string `json:"uid"`
	Userid      int
	Username    string
	Player      string `json:"player"`
	Timestamp   int    `json:"timestamp"`
	Stampconv   time.Time
	Legend      string       `json:"legendPlayed"`
	Tracker     []Apitracker `json:"-"`
	Seltrackers Pulltracker
	Throwaway   string
	Totdmg      int
	Handi       int
	Adjdmg      int
	Inctourn    int
	Rawtracker  json.RawMessage `json:"event"`
	Importdate  time.Time
}

//Pulltracker exp
type Pulltracker struct {
	Val1 string
	Val2 string
	Val3 string
}

//Apitracker exp
type Apitracker struct {
	Val  int    `json:"value"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

//Apievent exp
type Apievent struct {
	A1 string `json:"action"`
}

//Cat is global var for 3 tracker categories currently in use
var Cat Cats

//Char is global var for characters currently allowed
var Char []string

func initcats() {
	Cat = Cats{"smg_kills", "shotgun_kills", "top_3"}
	Char = []string{"Gibraltar", "Caustic", "Wattson", "Rampart"}
}

//Logapigame will enter data from games
func Logapigame(g Apigames) error {
	form, err := db.Prepare("INSERT INTO apigames(uid,username,psnid,tstamp,legend,totaldmg,handicap,adjdmg,inctourn,importdate) VALUES (?,?,?,?,?,?,?,?,?,?) on DUPLICATE KEY UPDATE uid=?")
	if err != nil {
		log.Println("logapigame err:", err.Error())
		return err
	}
	_, err = form.Exec(g.Userid, g.Username, g.Player, g.Stampconv, g.Legend, g.Totdmg, g.Handi, g.Adjdmg, g.Importdate, g.Inctourn, g.Userid)
	if err != nil {
		log.Println("logapigame err:", err.Error())
		return err
	}
	form.Close()
	return nil

}

//Logtracker exp
func Logtracker(g Apigames, tracker Apitracker) error {
	form, err := db.Prepare("INSERT INTO apitracker(uid,tstamp,val,keyid,nameid) VALUES (?,?,?,?,?) on DUPLICATE KEY UPDATE uid=?")
	if err != nil {
		log.Println("logtracker prep err:", err.Error())
		return err
	}
	_, err = form.Exec(g.Userid, g.Stampconv, tracker.Val, tracker.Key, tracker.Name, g.Userid)
	if err != nil {
		log.Println("logtracker exec err:", err.Error())
		return err
	}
	form.Close()
	return nil
}

//SeltopAPImatches gets most recent match list for any user from api
func SeltopAPImatches(username string) Apimain {
	qry := fmt.Sprintf("select uid,username,psnid,tstamp,legend,totaldmg,handicap,adjdmg,importdate from apigames where username='%s' order by tstamp desc limit 3;", username)
	res, err := db.Query(qry)
	handleError(err)

	var sl Apimain
	for res.Next() {
		var game Apigames
		// for each row, scan the result into our tag composite object
		err := res.Scan(&game.Userid, &game.Username, &game.Player, &game.Stampconv, &game.Legend, &game.Totdmg, &game.Handi, &game.Adjdmg, &game.Importdate)
		handleError(err)

		sl.Apiseries = append(sl.Apiseries, game)
	}
	res.Close()
	return sl
}

//Sellatestimport gets most recent match list for any user from api
func Sellatestimport() time.Time {
	var t time.Time
	qry := "select max(importdate) from apigames;"
	err := db.QueryRow(qry).Scan(&t)
	handleError(err)
	return t
}

//Seltrackers gets trackers for corresponding game
func Seltrackers(u int, t time.Time) []Apitracker {
	//log.Println("in db seltrackers, time:", t)
	qry := fmt.Sprintf("select val,keyid,nameid from apitracker where uid='%d' and tstamp='%v';", u, t)
	res, err := db.Query(qry)
	handleError(err)

	var sl []Apitracker
	for res.Next() {
		var game Apitracker
		// for each row, scan the result into our tag composite object
		err := res.Scan(&game.Val, &game.Key, &game.Name)
		handleError(err)

		sl = append(sl, game)
	}
	res.Close()
	return sl
}
func he(err error) {
	if err != nil {
		log.Println("error:", err)
	}
}
