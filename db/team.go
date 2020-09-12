package apexdb

import (
	//"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" //comment
	"log"
	"time"
)

//Team is list of players assigned to each team
type Team struct {
	Team1 string
	Team2 string
}

//Getteamassigns will get the team assignments
func Getteamassigns(teamassign int) []string {
	now := time.Now()
	res, err := db.Query("select propername from user where teamassign=? and sess_exp > ?;", teamassign, now)

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

//Assignteam updates team assignment number
func Assignteam(username string, teamassignment int) {
	form, err := db.Prepare("UPDATE user SET teamassign = ? WHERE username = ?;")
	handleError(err)
	_, err = form.Exec(teamassignment, username)
	handleError(err)
	return
}

//Switchteams switch to other team
func Switchteams(propername string) error {
	num, s := getteamnum(propername)
	t1 := Getteamassigns(1)
	t2 := Getteamassigns(2)
	if num == 1 {
		if len(t2) < 3 {
			//log.Println("len(t2)==", len(t2))
			Assignteam(s, 2)
		} else {
			return errors.New("Team 2 full")
		}
	} else if num == 2 {
		if len(t1) < 3 {
			//log.Println("len(t1)==", len(t1))
			Assignteam(s, 1)
		} else {
			return errors.New("Team 1 full")
		}
	} else {
		log.Println("error: no team found")
		return errors.New("no team found")
	}
	return nil
}

//Removeplayer will remove from any team
func Removeplayer(propername string) {
	_, s := getteamnum(propername)
	Assignteam(s, 0)

}
func getteamnum(propername string) (int, string) {
	qry := fmt.Sprintf("select teamassign,username from user where propername = '%s';", propername)
	var num int
	var s string
	err := db.QueryRow(qry).Scan(&num, &s)
	handleError(err)

	return num, s
}

//Getbothteams will put all assigned players in struct
func Getbothteams() []Team {

	t1 := Getteamassigns(1)
	t2 := Getteamassigns(2)

	var l int
	if len(t1) > len(t2) {
		l = len(t1)
	} else {
		l = len(t2)
	}

	var sl []Team
	for i := 0; i < l; i++ {
		var t Team

		if len(t1) > i {
			t.Team1 = t1[i]
		} else {
			t.Team1 = "-"
		}

		if len(t2) > i {
			t.Team2 = t2[i]
		} else {
			t.Team2 = "-"
		}

		sl = append(sl, t)
	}
	return sl
}
