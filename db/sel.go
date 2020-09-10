package apexdb

import (
	//"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" //comment
	"log"
	"time"
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

//Creds holds user login info
type Creds struct {
	Username string
	Pass     string
	Sessid   string
	Exp      time.Time
}

//User holds display info for user status
type User struct {
	Activeuser []string
	Teams      []Team
	Errcode    string
}

//Team is list of players assigned to each team
type Team struct {
	Team1 string
	Team2 string
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

//Wipetourn will wipe game table
func Wipetourn(sessid string) {
	username := Getuserfromsess(sessid)
	_, err := db.Exec("DELETE FROM games WHERE username=?;", username)
	if err != nil {
		fmt.Println(err.Error())
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

//Seluser will return user session info
func Seluser(u string) Creds {
	var res Creds
	qry := fmt.Sprintf("select username, pass, sess_id, sess_exp from user where username = '%s';", u)
	err := db.QueryRow(qry).Scan(&res.Username, &res.Pass, &res.Sessid, &res.Exp)
	handleError(err)

	return res
}

//Selsess will return session info for sess id
func Selsess(sessid string) Creds {
	var res Creds
	qry := fmt.Sprintf("select username, pass, sess_id, sess_exp from user where sess_id = '%s';", sessid)
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

//Updsessid updates sessid for given user
func Updsessid(username string, sessid string) {
	form, err := db.Prepare("UPDATE user SET sess_id = ? WHERE username = ?;")
	handleError(err)
	_, err = form.Exec(sessid, username)
	handleError(err)
	return
}

//Updsessexp will return session info for sess id
func Updsessexp(sessid string, newexp time.Time) {
	form, err := db.Prepare("UPDATE user SET sess_exp = ? WHERE sess_id = ?;")
	handleError(err)
	_, err = form.Exec(newexp, sessid)
	handleError(err)
	return
}

//Checkallvalidsess removes sess_id for all expired sessions
func Checkallvalidsess() {
	now := time.Now()
	form, err := db.Prepare("UPDATE user SET sess_id = '' WHERE sess_exp < ?;")
	handleError(err)
	_, err = form.Exec(now)
	handleError(err)
	return
}

//Delsess removes sess_id for all expired sessions
func Delsess(sessid string) {
	//delete sess exp
	now := time.Now()
	oldexp := now.AddDate(0, -1, 0)
	form, err := db.Prepare("UPDATE user SET sess_exp = ?,sess_id='',teamassign=0 WHERE sess_id=?")
	handleError(err)
	_, err = form.Exec(oldexp, sessid)
	handleError(err)
	/*
		//delete sess id
		form, err = db.Prepare("UPDATE user SET sess_id = '' WHERE sess_id=?")
		handleError(err)
		_, err = form.Exec(sessid)
		handleError(err)
	*/
	return
}

//Delallsess removes sess_id for all expired sessions
func Delallsess() {
	//delete sess exp
	now := time.Now()
	oldexp := now.AddDate(0, -1, 0)
	form, err := db.Prepare("UPDATE user SET sess_exp = ?,sess_id='',teamassign=0")
	handleError(err)
	_, err = form.Exec(oldexp)
	handleError(err)
	log.Println("all sessions deleted")
	return
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

//Updpropername allows users to update their own name
func Updpropername(sessid string, newname string) {
	form, err := db.Prepare("UPDATE user SET propername = ? WHERE sess_id = ?;")
	handleError(err)
	_, err = form.Exec(newname, sessid)
	handleError(err)
	return
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

//Getproper returns the actual username for a players proper name
func Getproper(username string) string {
	var res string
	qry := fmt.Sprintf("select propername from user where username = '%s';", username)
	err := db.QueryRow(qry).Scan(&res)
	handleError(err)

	return res
}

//Getactiveusers will return []string of mode requested
func Getactiveusers() []string {
	now := time.Now()
	qry := "select propername from user where sess_exp > ? and CHAR_LENGTH(sess_id)>5 "
	res, err := db.Query(qry, now)
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
