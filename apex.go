package main

import (
	//"logger"
	"apexrand/random"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

//Res exp
var Res random.Player
var viewcounter int = 0
var rollcounter int = 0

func main() {
	http.HandleFunc("/current", handler1)
	http.HandleFunc("/reroll1", reroll1)
	http.HandleFunc("/reroll2", reroll2)
	http.HandleFunc("/reroll3", reroll3) //roll both
	http.HandleFunc("/apex", handler1)
	http.HandleFunc("/", helloServer)
	http.ListenAndServe(":9999", nil)

}
func reroll1(w http.ResponseWriter, r *http.Request) {
	rollcounter++
	log.Println("rollcounter: ", rollcounter)
	Res = random.Rollnewload(Res, 1)
	//log.Println("reroll1 res:", Res)
	http.Redirect(w, r, "/current", 302)
}
func reroll2(w http.ResponseWriter, r *http.Request) {
	rollcounter++
	log.Println("rollcounter: ", rollcounter)
	Res = random.Rollnewload(Res, 2)
	http.Redirect(w, r, "/current", 302)
}
func reroll3(w http.ResponseWriter, r *http.Request) {
	rollcounter++
	log.Println("rollcounter: ", rollcounter)
	Res = random.Rollnewload(Res, 3)
	http.Redirect(w, r, "/current", 302)
}
func handler1(w http.ResponseWriter, r *http.Request) {
	viewcounter++
	log.Println("viewcounter: ", viewcounter)
	log.Println("Page loaded")
	tmpl := template.Must(template.ParseFiles("forms.html"))
	tmpl.Execute(w, Res)
}
func helloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "The server you were connecting to was disconnected or no longer in use.  Please try your request again or leave a message below")
}
