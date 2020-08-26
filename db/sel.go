package apexdb

import (
	//"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" //comment
)

//Curse holds id, adjustment factor and type for all potential curses
type Curse struct {
	ID         int
	Descrip    string
	Adj        float64
	Assigntype string
}

//Stat holds row level stat data
type Stat struct {
	ID      string
	Descrip string
	Adj     float64
	Count   int
}

//Stats holds combined stat data
type Stats struct {
	Playercount []Stat
	Cursecount  []Stat
	Totalrolls  int
	Thresh      int
}

//Selcursestats will return []string of mode requested
func Selcursestats() []Stat {
	qry := "select a.curse_id, b.descrip, b.adjfactor, count(*) from assign as a left join curse as b on a.curse_id=b.id group by curse_id order by curse_id;"
	res, err := db.Query(qry)
	handleError(err)
	var sl []Stat
	for res.Next() {
		var c Stat
		// for each row, scan the result into our tag composite object
		err := res.Scan(&c.ID, &c.Descrip, &c.Adj, &c.Count)
		handleError(err)

		sl = append(sl, c)
	}
	res.Close()
	return sl
}

//Selplayerstats will return []string of mode requested
func Selplayerstats() []Stat {
	qry := "select player_id, count(*) from assign group by player_id order by player_id;"
	res, err := db.Query(qry)
	handleError(err)
	var sl []Stat
	for res.Next() {
		var c Stat
		// for each row, scan the result into our tag composite object
		err := res.Scan(&c.ID, &c.Count)
		handleError(err)

		sl = append(sl, c)
	}
	res.Close()
	return sl
}

//Wipestats will wipe stat table
func Wipestats() {
	res, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("foreign constraint check = 0", res)
	}
	res, err = db.Exec("TRUNCATE TABLE assign")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Successfully truncated assign..", res)
	}
	res, err = db.Exec("SET FOREIGN_KEY_CHECKS = 1;")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("foreign constraint check = 1", res)
	}
}

//Selcurse will return []string of mode requested
func Selcurse() []Curse {
	qry := "select id, descrip, adjfactor, assigntype from curse"
	res, err := db.Query(qry)
	handleError(err)
	var sl []Curse
	for res.Next() {
		var c Curse
		// for each row, scan the result into our tag composite object
		err := res.Scan(&c.ID, &c.Descrip, &c.Adj, &c.Assigntype)
		handleError(err)

		sl = append(sl, c)
	}
	res.Close()
	return sl
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
