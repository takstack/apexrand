package main

import (
	//"logger"
	"apexrand/random"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"time"
)

//Res exp
var Res random.Player
var viewcounter int = 0

func main() {
	srv := &http.Server{
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,

		Addr: ":9999",
	}

	http.HandleFunc("/current", handler1)
	http.HandleFunc("/reroll1", reroll1)
	http.HandleFunc("/reroll2", reroll2)
	http.HandleFunc("/reroll3", reroll3) //roll both
	http.HandleFunc("/apex", handler1)
	http.HandleFunc("/", helloServer)
	log.Println(srv.ListenAndServe())

}
func reroll1(w http.ResponseWriter, r *http.Request) {
	Res = random.Rollnewload(Res, 1)
	//log.Println("reroll1 res:", Res)
	http.Redirect(w, r, "/current", 302)
}
func reroll2(w http.ResponseWriter, r *http.Request) {
	Res = random.Rollnewload(Res, 2)
	http.Redirect(w, r, "/current", 302)
}
func reroll3(w http.ResponseWriter, r *http.Request) {
	Res = random.Rollnewload(Res, 3)
	http.Redirect(w, r, "/current", 302)
}
func handler1(w http.ResponseWriter, r *http.Request) {
	viewcounter++
	ip, _ := fromRequest(r)
	log.Println(ip, ", viewcounter: ", viewcounter)
	//log.Println("Page loaded")
	tmpl := template.Must(template.ParseFiles("forms.html"))
	tmpl.Execute(w, Res)
}
func helloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "The server you were connecting to was disconnected or no longer in use.  Please try your request again or leave a message below")
}

func fromRequest(req *http.Request) (net.IP, error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}
	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}
	return userIP, nil
}
