package main

import (
	//"logger"
	"apexrand/db"
	"apexrand/random"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"
)

//Res exp
var Res random.Player

var viewcounter int = 0

func main() {
	srv := &http.Server{
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       10 * time.Second,

		Addr: ":9999",
	}
	apexdb.Insvarfromfile()
	apexdb.Inscursefromfile()
	http.HandleFunc("/current", handler1)
	//http.HandleFunc("/testroll", testroll)
	http.HandleFunc("/reroll1", reroll1)
	http.HandleFunc("/reroll2", reroll2)
	http.HandleFunc("/reroll3", reroll3) //roll both
	http.HandleFunc("/stats", getstats)
	http.HandleFunc("/wipestats", wipestats)
	http.HandleFunc("/rollauto", rollauto)
	http.HandleFunc("/apex", handler1)
	http.HandleFunc("/login", login)
	http.HandleFunc("/", helloServer)

	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))
	srv.SetKeepAlivesEnabled(false)
	log.Fatalln(srv.ListenAndServe())

}

//not working
func testroll(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < 2; i++ {
		reqbody, err := json.Marshal(map[string]string{
			"name": "howdy",
		})
		if err != nil {
			log.Fatal(err)
		}
		req, err := http.NewRequest("POST", "/reroll3", bytes.NewBuffer(reqbody))
		if err != nil {
			log.Fatal(err)
		}
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(body))
	}
}
func getstats(w http.ResponseWriter, r *http.Request) {
	log.Println("stats started")

	var st apexdb.Stats
	st.Cursecount = apexdb.Selcursestats()
	st.Playercount = apexdb.Selplayerstats()
	st.Totalrolls = random.Rollcounter

	tmpl := template.Must(template.ParseFiles("stats.html"))
	tmpl.Execute(w, st)

	return
}
func rollauto(w http.ResponseWriter, r *http.Request) {
	log.Println("autorolling")

	Res = random.Autoroller(Res)

	http.Redirect(w, r, "/stats", 302)
	return
}
func wipestats(w http.ResponseWriter, r *http.Request) {
	log.Println("stats wiped")

	apexdb.Wipestats()
	random.Rollcounter = 0
	http.Redirect(w, r, "/stats", 302)
	return
}
func reroll1(w http.ResponseWriter, r *http.Request) {
	log.Println("reroll1 started")
	Res = random.Rollnewload(Res, 1)
	_ = Res
	//log.Println("reroll res:", Res)
	//log.Println("reroll1 res:", Res)
	http.Redirect(w, r, "/current", 302)
	return
}
func reroll2(w http.ResponseWriter, r *http.Request) {
	log.Println("reroll2 started")
	Res = random.Rollnewload(Res, 2)
	_ = Res
	//log.Println("reroll res:", Res)
	http.Redirect(w, r, "/current", 302)
	return
}
func reroll3(w http.ResponseWriter, r *http.Request) {
	log.Println("reroll3 started")
	Res = random.Rollnewload(Res, 3)
	_ = Res
	//log.Printf("reroll res:%+v", Res.Tchals)
	http.Redirect(w, r, "/current", 302)
	return
}
func login(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("login.html"))
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}
	det := Credentials{
		Username: r.FormValue("fielduser"),
		Password: r.FormValue("fieldpass"),
	}
	log.Printf("login:%+v", det)
	http.Redirect(w, r, "/current", 302)
	return
}
func handler1(w http.ResponseWriter, r *http.Request) {
	//r.Host = "sheldonconn.ddns.net" //attempt to eliminate hanging issue
	viewcounter++
	ip, ips, err := fromRequest(r)
	if err != nil {
		log.Println("Error - IP Parse: ", err)
	}
	log.Println(r.Header)
	//log.Println("Read cookie:", r.Header.Get("Cookie"))
	log.Printf("%v, viewcounter:%d \n", ip, viewcounter)
	log.Printf("Request executed \n\n")

	expiration := time.Now().Add(1 * time.Hour)
	cookie := http.Cookie{Name: "CSRFtoken", Value: ips, Expires: expiration, SameSite: http.SameSiteStrictMode}
	http.SetCookie(w, &cookie)

	tmpl := template.Must(template.ParseFiles("forms.html"))
	tmpl.Execute(w, Res)
	return
}
func helloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "The server you were connecting to was disconnected or no longer in use.  Please try your request again or leave a message below")
	return
}

func fromRequest(req *http.Request) (net.IP, string, error) {
	//log.Println("parsing new req ip")
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, "", fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, "", fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}

	//create bogus csrf val based on ip
	IPsplit := strings.Split(ip, ".") //split ip into slice
	IPcalc := 0
	for _, elem := range IPsplit {
		//var err error
		ipint, err := strconv.Atoi(elem)
		IPcalc = IPcalc + ipint
		if err != nil {
			return nil, "", fmt.Errorf("userip: %q did not convert to int", req.RemoteAddr)
		}
	}
	userIPs := strconv.Itoa(IPcalc)
	if err != nil {
		return nil, "", fmt.Errorf("userip: %q did not convert to str", req.RemoteAddr)
	}
	//log.Println("ip parsed")
	//log.Println(userIP, userIPs)
	return userIP, userIPs, nil
}
