package apexdb

import (
	_ "github.com/go-sql-driver/mysql" //comment
	"strconv"
	//"sync"
	//"encoding/csv"
	"apexrand/qrw"
	"errors"
	"log"
	"sort"
	"time"
)

//Tourney exp
type Tourney struct {
	T           []Tourn
	P           string
	G           []Game //holds game vals for individual games
	Activeusers []Onlineuser
	Errcode     string //err val for incorrect dmg input
}

//Tourn is dataset to show standings
type Tourn struct {
	Player   string
	Numgames int
	Avgdmg   int
	Sumdmg   int
	Handicap int
	Adjavg   int
	Adjsum   int
	Games    []Game
}

//Game hold indv game vals
type Game struct {
	ID       int
	Username string
	Dmg      int
	Place    int
	Placedmg int
	Totdmg   int
	Handicap int
	Adjdmg   int
	Gametime time.Time
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
	handi := Gethandifromuser(username)
	log.Println("handi:", handi)
	pdmg := 0

	//determine if place should award additional damage
	switch {
	case p > 30 || p < 1 || d < 0:
		log.Println("loggame parameters out of bounds:", p, d)
		return errors.New("FUCKING IDIOT - parameters out of bounds")
	case p == 1:
		pdmg = 400
	case p < 4 && p > 1:
		pdmg = 200
	}
	tdmg := pdmg + d
	adjdmg := int(float64(tdmg) * ((10000 - float64(handi)) / 10000))
	tourndate := time.Date(2020, time.Month(10), 30, 0, 0, 0, 0, time.UTC)

	form, err := db.Prepare("INSERT INTO games(username,dmg,place,placedmg,totaldmg,handicap, adj_dmg,inc_tourn,tourn,gametime) VALUES (?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		panic(err.Error())
	}
	_, err = form.Exec(username, d, p, pdmg, tdmg, handi, adjdmg, 1, tourndate, now)
	if err != nil {
		panic(err.Error())
	}
	return nil
}

//Wipetourn will wipe game table
func Wipetourn(sessid string) {
	username := Getuserfromsess(sessid)
	_, err := db.Exec("UPDATE games SET inc_tourn=0 WHERE username=?;", username)
	if err != nil {
		log.Println(err.Error())
	}

}

//Seltourngames will get game data
func Seltourngames() []Tourn {
	playerlist := getlistoftournplayers()
	var pSL []Tourn

	for _, player := range playerlist {
		var p Tourn
		p.Player = Getproper(player)
		p.Games = getplayersgames(player)

		for _, g := range p.Games {
			p.Sumdmg = p.Sumdmg + g.Totdmg
			p.Adjsum = p.Adjsum + g.Adjdmg
			p.Numgames++
		}
		if p.Numgames > 0 {
			p.Avgdmg = p.Sumdmg / p.Numgames
			p.Adjavg = p.Adjsum / p.Numgames
			pSL = append(pSL, p)
		}

	}

	//if there are no games, add empty set
	if len(pSL) == 0 {
		t := Tourn{}
		pSL = append(pSL, t)
		log.Println("seltourngames end len==0. psl:", pSL)
		return pSL
	}

	sort.SliceStable(pSL, func(i, j int) bool {
		return pSL[i].Adjsum > pSL[j].Adjsum
	})

	return pSL
}
func getplayersgames(player string) []Game {

	qry := "select id, username, dmg, place, placedmg, totaldmg, handicap, adj_dmg, gametime from games where username=? and inc_tourn='1' order by totaldmg desc limit 10"
	res, err := db.Query(qry, player)
	handleError(err)
	var sl []Game
	for res.Next() {
		var g Game
		// for each row, scan the result into our tag composite object
		err := res.Scan(&g.ID, &g.Username, &g.Dmg, &g.Place, &g.Placedmg, &g.Totdmg, &g.Handicap, &g.Adjdmg, &g.Gametime)
		g.Gametime = convertutc(g.Gametime)
		handleError(err)

		sl = append(sl, g)
	}
	res.Close()

	return sl
}
func getplayersgamesspecdate(player string, tourndate time.Time) [][]string {

	qry := "select username, totaldmg, tourn from games where username=? and tourn>? order by totaldmg desc limit 20"
	res, err := db.Query(qry, player, tourndate)
	handleError(err)
	var sl [][]string
	for res.Next() {
		var u string
		var d int
		var t time.Time
		// for each row, scan the result into our tag composite object
		err := res.Scan(&u, &d, &t)
		dconv := strconv.Itoa(d)
		tconv := t.String()
		handleError(err)

		sl = append(sl, []string{u, dconv, tconv})
	}
	res.Close()

	return sl
}
func convertutc(t time.Time) time.Time {
	var local time.Time
	location, err := time.LoadLocation("America/Los_Angeles")
	if err == nil {
		local = t.In(location)
	}

	return local

}
func getlistoftournplayers() []string {
	qry := "select username from games group by username"
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

//Sethandicap will set new handicaps
func Sethandicap() {
	sl := Seltourngames()
	for _, elem := range sl {
		player := Getuser(elem.Player)
		handi := int((1 - (1000 / float64(elem.Sumdmg))) * 10000)
		//log.Println("capture:", 1000/float64(elem.Sumdmg))
		//log.Println("handi calc player, sumdmg,handi:", player, elem.Sumdmg, handi)
		Sethandifromuser(player, handi)

	}
}

//Writetourngamescsv will get previous tournament games and export to csv
func Writetourngamescsv() {
	log.Println("writing game data to csv")
	cw := qrw.StartCSVwriter(qrw.Openwritefile("apexrand/file/gamedata.csv"))
	t := Seltourngames()
	for _, elem := range t {
		for _, game := range elem.Games {
			sl := []string{game.Username, strconv.Itoa(game.Totdmg)}
			log.Printf("sl: %v, type: %T\n", sl, sl[1])
			err := cw.Write(sl)
			if err != nil {
				log.Println("csv write error:", err)
			}
			cw.Flush()
		}
	}

}

//Writetourngamescsv2 will get previous tournament games and export to csv
func Writetourngamescsv2() {
	log.Println("writing game data to csv")
	cw := qrw.StartCSVwriter(qrw.Openwritefile("apexrand/file/gamedata.csv"))

	playerlist := getlistoftournplayers()
	tourndate := time.Date(2020, time.Month(1), 1, 0, 0, 0, 0, time.UTC)

	for _, elem := range playerlist {
		sl := getplayersgamesspecdate(elem, tourndate)
		for _, game := range sl {
			g := []string{game[0], game[1], game[2]}
			log.Printf("sl: %v", g)
			err := cw.Write(g)
			if err != nil {
				log.Println("csv write error:", err)
			}
			cw.Flush()
		}
	}
	/*
		tourndate = time.Date(2020, time.Month(9), 11, 0, 0, 0, 0, time.UTC)

		for _, elem := range playerlist {
			sl := getplayersgamesspecdate(elem, tourndate)
			for _, game := range sl {
				g := []string{game[0], game[1], game[2]}
				log.Printf("sl: %v", g)
				err := cw.Write(g)
				if err != nil {
					log.Println("csv write error:", err)
				}
				cw.Flush()
			}
		}
	*/
}
