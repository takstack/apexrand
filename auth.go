package main

import (
	"apexrand/api"
	apexdb "apexrand/db"
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

type formlogin struct {
	user string
	pass string
}

func login(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/html/login.html"))
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}
	det := formlogin{
		user: r.FormValue("fielduser"),
		pass: r.FormValue("fieldpass"),
	}
	log.Printf("login:%+v", det)
	confirmed := apexdb.Getconfstatus(det.user)
	if !confirmed {
		fmt.Fprintf(w, "not confirmed")
		return
	}
	err := auth(det)

	if err != nil {
		log.Println("auth err", err)
		fmt.Fprintf(w, "login failed")
		return
	}
	_, ips, _ := fromRequest(r)
	renewcookie(w, r, det.user, ips)

	teamassignment, err := assignteams()
	if err != nil {
		log.Println("error assigning teams", err)
	}
	apexdb.Assignteam(det.user, teamassignment)

	http.Redirect(w, r, "/home", http.StatusFound)

}
func reg(w http.ResponseWriter, r *http.Request) {
	log.Println("reg started")
	tmpl := template.Must(template.ParseFiles("static/html/reg.html"))

	//_, _, _ = fromRequest(r)

	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}
	//to log a game for that player
	var user apexdb.Creds
	user.Email = r.FormValue("email")
	user.Platform = r.FormValue("platform")
	user.Playerid = r.FormValue("playerid")
	user.Romanname = r.FormValue("romanname")
	user.Username = r.FormValue("username")
	user.Pass = r.FormValue("pass")

	switch user.Platform {
	case "PSN":
		user.Platform = "PS4"
	case "Xbox":
		user.Platform = "X1"
	}

	log.Println("form action received reg:", user.Email, user.Platform, user.Playerid, user.Romanname, user.Username, user.Pass)

	user.Confstr = createsessid()
	err := apexdb.Createuser(user)
	if err != nil {
		//Tourney.Errcode = err.Error()
		log.Println("reg error: ", err)
	}
	//renewcookie(w, r, username, ips)
	sendemail(user.Email, "http://apexlott.com/confirm?conf="+user.Confstr)
	log.Println("conf address: ", "http://apexlott.com/confirm?conf="+user.Confstr)
	http.Redirect(w, r, "/waitconf", http.StatusFound)
}
func waitconf(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/html/waitconf.html"))
	tmpl.Execute(w, nil)
}

func confirm(w http.ResponseWriter, r *http.Request) {
	log.Println("tourney started")
	tmpl := template.Must(template.ParseFiles("static/html/confirmed.html"))
	//ip prints
	//_, _, _ = fromRequest(r)
	//check confirmation code, update user db if code matches
	conf := r.URL.Query().Get("conf")
	err := apexdb.Updconfirmed(conf)

	if err != nil {
		http.Redirect(w, r, "/waitconf", http.StatusFound)
	}
	player := apexdb.Getplayeridfromuser(apexdb.Getuserfromconf(conf))
	platform := apexdb.Getplatfrompsn(player)

	err = api.Addapiuser(player, platform)
	if err != nil {
		log.Println("add api user err: ", err)
	}
	tmpl.Execute(w, nil)
}
func auth(l formlogin) error {

	c := apexdb.Seluser(l.user)
	log.Println("db user call", c)

	// if no user, do something else, then check for pass match
	//check if already logged in with "chkvalidsession"
	if l.user != c.Username || l.pass != c.Pass {
		return errors.New("login not successful")
	}

	return nil
}
func chkvalidsession(w http.ResponseWriter, r *http.Request) bool {

	cookie, err := getcookie(w, r, "apextoken")
	if err != nil {
		log.Printf("chkvalidsess - no cookie\n\n")
		http.Redirect(w, r, "/login", http.StatusFound)
		return false
	}

	sessid := cookie.Value

	_, ips, _ := fromRequest(r)
	//log.Printf("userfromsess: %s\n", apexdb.Getuserfromsess(sessid))

	apexdb.Logip(sessid, ips)
	log.Println("entered user ip: ", sessid, ips)
	////log.Println("before chkvalid db call")
	c := apexdb.Selsess(sessid)
	//log.Println("token expire time:", c.Exp)

	now := time.Now()

	if c.Exp.After(now) {
		apexdb.Updsessexp(sessid, newexpire())
		setnewapexcookie(w, r, sessid)
		//log.Println("team1 ", apexdb.Getteamassigns(1))
		//log.Println("team2 ", apexdb.Getteamassigns(2))
		return true
	}

	http.Redirect(w, r, "/login", http.StatusFound)
	return false

}

func chgname(w http.ResponseWriter, r *http.Request, newname string) {
	if len(newname) > 20 {
		log.Println("error: new name too long")
		//fmt.Fprintf(w, "error: name too long")
		return
	}
	log.Println("in chgname")
	cookie, err := getcookie(w, r, "apextoken")
	if err != nil {
		log.Println("chkvalidsess - no cookie")
		http.Redirect(w, r, "/login", http.StatusFound)

	}
	sessid := cookie.Value

	apexdb.Updpropername(sessid, newname)

}
func addplayertoteam(propername string) {
	num, err := assignteams()
	if err != nil {
		log.Println("error adding to team, finding team number")
	}
	user := apexdb.Getuser(propername)
	//log.Println("user,num, in addplayer to team", user, num)
	apexdb.Assignteam(user, num)
}
func assignteams() (int, error) {
	t1 := len(apexdb.Getteamassigns(1))
	t2 := len(apexdb.Getteamassigns(2))
	if t1+t2 >= 6 {
		return 0, errors.New("teams full")
	}
	if t1 < 3 {
		return 1, nil
	}
	return 2, nil
}
func renewcookie(w http.ResponseWriter, r *http.Request, user string, ips string) {
	sessid := createsessid()
	delcookie(w, r, sessid)
	setnewapexcookie(w, r, sessid)

	log.Println("sessid", sessid)
	apexdb.Updsessid(user, sessid)
	apexdb.Logip(sessid, ips)
	log.Println("entered user ip: ", sessid, ips)
	apexdb.Updsessexp(sessid, newexpire())

}
func getcookie(w http.ResponseWriter, r *http.Request, s string) (*http.Cookie, error) {
	//_, _, _ = fromRequest(r)

	cookie, err := r.Cookie(s)
	if err != nil {
		for _, c := range r.Cookies() {
			log.Println("range cookies:", c)
		}
		log.Println("r.header:", r.Header)
		body, err1 := ioutil.ReadAll(r.Body)
		if err1 != nil {
			log.Printf("Error reading body: %v", err)
		}
		log.Println("r.body:", string(body))
		log.Println("error retrieving cookie-getcookie", err)
		return cookie, err
	}

	return cookie, nil
}

func setnewapexcookie(w http.ResponseWriter, r *http.Request, sessid string) {

	newexp := newexpire() //create new expiration

	cookie := http.Cookie{Name: "apextoken", Value: sessid, Path: "/", HttpOnly: true, Secure: false, Expires: newexp, SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, &cookie)

}
func delcookie(w http.ResponseWriter, r *http.Request, sessid string) {

	//newexp := newexpire() //create new expiration
	log.Println("deleting cookie")
	cookie := http.Cookie{Name: "apextoken", Value: "", Path: "/", MaxAge: -1, SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, &cookie)

}

func newexpire() time.Time {
	now := time.Now()
	newexp := now.AddDate(0, 1, 0)
	//log.Println("generated new exp time", newexp)
	return newexp
}

func createsessid() string {

	b := make([]byte, 12)
	if _, err := rand.Reader.Read(b); err != nil {
		panic(err)
	}
	//fmt.Println("original", b)
	n := base64.URLEncoding.EncodeToString(b)

	/*
		fmt.Println(n)
		m, err := base64.URLEncoding.DecodeString(n)
		if err != nil {
			panic(err)
		}
		fmt.Println("decoded", m)
	*/
	return n
}
func logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("apextoken")
	if err != nil {
		log.Println("error retrieving cookie")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	sessid := cookie.Value
	apexdb.Delsess(sessid)
	http.Redirect(w, r, "/login", http.StatusFound)

}

//func for shopbot
func getemailkey() []string {
	f, err := os.Open("/var/lib/api/emailkey")
	if err != nil {
		log.Println("file open error:", err)
	}
	scanner := bufio.NewScanner(f)

	var key []string
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			fmt.Println("error: reading standard input:", err)
		}

		key = append(key, scanner.Text())
	}
	//log.Println("apikey:", string(r))
	return key
}
func sendemail(addr string, link string) {
	key := getemailkey()
	auth := smtp.PlainAuth("", key[0], key[1], "smtp.gmail.com")
	//sending without "To:" will make it bcc:
	msg := []byte("To:" + addr + "\r\n" + "Subject:Apexlott.com Confirmation Link\r\n" + "\r\n" + link)

	err := smtp.SendMail("smtp.gmail.com:587", auth, key[0], []string{addr}, msg)
	if err != nil {
		log.Println(err)
	}
	log.Println("smtp executed: ", addr)
}
