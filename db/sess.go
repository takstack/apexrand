package apexdb

import (
	//"database/sql"
	//"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" //comment
	"log"
	"time"
)

//Selsess will return session info for sess id
func Selsess(sessid string) Creds {
	var res Creds
	qry := fmt.Sprintf("select username, pass, sess_id, sess_exp from user where sess_id = '%s';", sessid)
	err := db.QueryRow(qry).Scan(&res.Username, &res.Pass, &res.Sessid, &res.Exp)
	handleError(err)

	return res
}

//Logip will record successful ip addresses
func Logip(sessid string, ip string) {
	form, err := db.Prepare("UPDATE user SET ip = ? WHERE sess_id = ?;")
	handleError(err)
	_, err = form.Exec(ip, sessid)
	handleError(err)
	return
}

//Getuserfromip will return username for an ip address
func Getuserfromip(ip string) string {
	var res string
	qry := fmt.Sprintf("select username from user where ip = '%s';", ip)
	err := db.QueryRow(qry).Scan(&res)
	handleError(err)

	return res
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
	log.Println("logout: ", Getuserfromsess(sessid))
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
