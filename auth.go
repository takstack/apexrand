package main

import (
	"apexrand/db"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type formlogin struct {
	user string
	pass string
}

func login(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("login.html"))
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}
	det := formlogin{
		user: r.FormValue("fielduser"),
		pass: r.FormValue("fieldpass"),
	}
	log.Printf("login:%+v", det)

	err := auth(det)

	if err != nil {
		log.Println("auth err", err)
		fmt.Fprintf(w, "login failed")
		return
	}

	sessid := createsessid()
	delcookie(w, r, sessid)
	setnewapexcookie(w, r, sessid)

	log.Println("sessid", sessid)
	apexdb.Updsessid(det.user, sessid)
	apexdb.Updsessexp(sessid, newexpire())

	teamassignment, err := assignteams()
	if err != nil {
		log.Println("error assigning teams", err)
	}
	apexdb.Assignteam(det.user, teamassignment)

	http.Redirect(w, r, "/current", 302)
	return
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
		http.Redirect(w, r, "/login", 302)
		return false
	}

	sessid := cookie.Value

	//log.Println("before chkvalid db call")
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

	http.Redirect(w, r, "/login", 302)
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
		http.Redirect(w, r, "/login", 302)
		return
	}
	sessid := cookie.Value

	apexdb.Updpropername(sessid, newname)
	return
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
		return 0, errors.New("Teams full")
	}
	if t1 < 3 {
		return 1, nil
	}
	return 2, nil
}
func getcookie(w http.ResponseWriter, r *http.Request, s string) (*http.Cookie, error) {
	ip, ips, err := fromRequest(r)
	_ = ips
	if err != nil {
		log.Println("Error - IP Parse: ", err)
	}
	log.Printf("request ip-getcookie: %v \n", ip)
	//log.Println("r.header:", r.Header)
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
		http.Redirect(w, r, "/login", 302)
		return
	}
	sessid := cookie.Value
	apexdb.Delsess(sessid)
	http.Redirect(w, r, "/login", 302)
	return
}
