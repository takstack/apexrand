package apexdb

import (
	//"database/sql"
	//"errors"
	//"fmt"
	_ "github.com/go-sql-driver/mysql" //comment
	//"log"
	"sort"
	"time"
)

//Tourney exp
type Tourney struct {
	T []Tourn
	P string
	G []Game
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
		p.Avgdmg = p.Sumdmg / p.Numgames
		pSL = append(pSL, p)
	}

	sort.SliceStable(pSL, func(i, j int) bool {
		return pSL[i].Sumdmg > pSL[j].Sumdmg
	})
	return pSL
}
func getplayersgames(player string) []Game {
	qry := "select id, username, dmg, place, placedmg, totaldmg, gametime from games where username=? order by totaldmg desc limit 30"
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
