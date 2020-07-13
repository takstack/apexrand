package main

import (
	//"logger"
	"apexrand/random"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
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

	http.HandleFunc("/current", handler1)
	http.HandleFunc("/reroll1", reroll1)
	http.HandleFunc("/reroll2", reroll2)
	http.HandleFunc("/reroll3", reroll3) //roll both
	http.HandleFunc("/apex", handler1)
	http.HandleFunc("/", helloServer)
	log.Fatalln(srv.ListenAndServe())

}

func reroll1(w http.ResponseWriter, r *http.Request) {
	Res = random.Rollnewload(Res, 1)
	//log.Println("reroll1 res:", Res)
	http.Redirect(w, r, "/current", 302)
	return
}
func reroll2(w http.ResponseWriter, r *http.Request) {
	Res = random.Rollnewload(Res, 2)
	http.Redirect(w, r, "/current", 302)
	return
}
func reroll3(w http.ResponseWriter, r *http.Request) {
	Res = random.Rollnewload(Res, 3)
	http.Redirect(w, r, "/current", 302)
	return
}
func handler1(w http.ResponseWriter, r *http.Request) {
	viewcounter++
	ip, ips, _ := fromRequest(r)
	log.Println("Read cookie:", r.Header.Get("Cookie"))
	log.Println(ip, ", viewcounter: ", viewcounter)
	//log.Println("Page loaded")

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
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, "", fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}
	userIP := net.ParseIP(ip)

	if userIP == nil {
		return nil, "", fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}

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
	//log.Println(userIP, userIPs)
	return userIP, userIPs, nil
}
