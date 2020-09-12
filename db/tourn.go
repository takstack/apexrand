package apexdb

import (
	_ "github.com/go-sql-driver/mysql" //comment
	"strconv"
	//"sync"
	"errors"
	"log"
	"sort"
	"time"
)

//Tourney exp
type Tourney struct {
	T           []Tourn
	P           string
	G           []Game
	Activeusers []string
	Errcode     string
}

//Tourn is dataset to show for tournament
type Tourn struct {
	Player   string
	Numgames int
	Avgdmg   int
	Sumdmg   int
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
	form, err := db.Prepare("INSERT INTO games(username,dmg,place,placedmg,totaldmg,inc_tourn,gametime) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		panic(err.Error())
	}
	_, err = form.Exec(username, d, p, pdmg, tdmg, 1, now)
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
	if len(playerlist) == 0 {
		t := Tourn{}
		pSL = append(pSL, t)
		return pSL
	}

	for _, player := range playerlist {
		var p Tourn
		p.Player = Getproper(player)
		p.Games = getplayersgames(player)

		for _, g := range p.Games {
			p.Sumdmg = p.Sumdmg + g.Totdmg
			p.Numgames++
		}
		if p.Numgames > 0 {
			p.Avgdmg = p.Sumdmg / p.Numgames
			pSL = append(pSL, p)
		}

	}

	sort.SliceStable(pSL, func(i, j int) bool {
		return pSL[i].Sumdmg > pSL[j].Sumdmg
	})
	return pSL
}
func getplayersgames(player string) []Game {
	qry := "select id, username, dmg, place, placedmg, totaldmg, gametime from games where username=? and inc_tourn='1' order by totaldmg desc limit 10"
	res, err := db.Query(qry, player)
	handleError(err)
	var sl []Game
	for res.Next() {
		var g Game
		// for each row, scan the result into our tag composite object
		err := res.Scan(&g.ID, &g.Username, &g.Dmg, &g.Place, &g.Placedmg, &g.Totdmg, &g.Gametime)
		g.Gametime = convertutc(g.Gametime)
		handleError(err)

		sl = append(sl, g)
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
