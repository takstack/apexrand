package apexdb

import (
	//"database/sql"
	//"errors"
	_ "github.com/go-sql-driver/mysql" //comment
	"log"
	//"log"
	"time"
)

//Curse holds id, adjustment factor and type for all potential curses
type Curse struct {
	ID         int
	Descrip    string
	Adj        float64
	Assigntype string
}

//Inscursefromfile imports the curses
func Inscursefromfile() {
	sl := readfile("importcurse.csv")
	//log.Println(sl)
	tx, err := db.Begin() //get connection
	handleError(err)
	//LOG.GL.Println("in batch insert after tx begin")
	qry := "INSERT INTO curse (id,descrip,adjfactor,assigntype) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE adjfactor=?"
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
