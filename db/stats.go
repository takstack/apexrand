package apexdb

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql" //comment
)

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
