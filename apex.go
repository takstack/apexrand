package main

import (
	//"logger"
	"apexrand/api"
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
	"net/url"
	//"runtime"
	"bufio"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

//Res exp
var Res random.Player

var viewcounter int = 0

func main() {

	srv := &http.Server{
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       15 * time.Second,

		Addr: ":80",
	}
	//apexdb.Opendb()
	apexdb.Insvarfromfile()
	apexdb.Inscursefromfile()
	apexdb.Insuserfromfile()

	//apexdb.Sethandicap() //set new handicaps based on closed tourney
	//apexdb.Writetourngamescsv2()
	//apexdb.Delallsess() leave all sessions open for now
	go api.Apipull()

	log.Println("reminder: set tournament time in loggame if in tournament")

	http.HandleFunc("/shopbot", shopbot)
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
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/teams", teams)
	http.HandleFunc("/tournament", tourneyapi)
	http.HandleFunc("/wipetourn", wipetourn)
	http.HandleFunc("/redirtourn", redirtourn)
	http.HandleFunc("/apires", apires)
	http.HandleFunc("/", helloServer)

	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))
	srv.SetKeepAlivesEnabled(false)
	log.Fatalln(srv.ListenAndServe())

}

//func for shopbot
func shopbot(w http.ResponseWriter, r *http.Request) {
	log.Println("shopbot started")
	tmpl := template.Must(template.ParseFiles("shopbot.html"))
	ip, ips, err := fromRequest(r)
	_ = ips
	if err != nil {
		log.Println("Error - IP Parse: ", err)
	}
	log.Printf("request ip: %v \n\n", ip)

	focus := r.URL.Query().Get("key")
	if focus == "shop2020getit" {
		tmpl.Execute(w, "")
		sendtxts()
	}
	http.Redirect(w, r, "/", 302)
	return

}

//func for shopbot
func sendtxts() {
	key := getemailkey()
	auth := smtp.PlainAuth("", key[0], key[1], "smtp.gmail.com")
	//log.Println("key", key)

	tonums := getphonelist()
	for _, addr := range tonums {
		go sendtext(key, auth, addr)
	}
}
func sendtext(key []string, auth smtp.Auth, addr string) {
	msg := []byte("To:" + addr + "\r\n" + "Subject:Sam Sam Bo Fam\r\n")

	err := smtp.SendMail("smtp.gmail.com:587", auth, key[0], []string{addr}, msg)
	if err != nil {
		log.Println(err)
	}
}

//func for shopbot
func getphonelist() []string {
	f, err := os.Open("/var/lib/api/phonelist")
	if err != nil {
		log.Println("file open error:", err)
	}
	scanner := bufio.NewScanner(f)

	var pl []string
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			fmt.Println("error: reading standard input:", err)
		}

		pl = append(pl, scanner.Text())
	}
	//log.Println("apikey:", string(r))
	return pl
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
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("stats started")
		ip, ips, err := fromRequest(r)
		_ = ips
		if err != nil {
			log.Println("Error - IP Parse: ", err)
		}
		log.Printf("request ip: %v \n\n", ip)

		var st apexdb.Stats
		st.Cursecount = apexdb.Selcursestats()
		st.Playercount = apexdb.Selplayerstats()
		st.Totalrolls = apexdb.Getnumrolls()
		st.Thresh = random.Thresh
		tmpl := template.Must(template.ParseFiles("stats.html"))
		tmpl.Execute(w, st)
	}
	return
}
func rollauto(w http.ResponseWriter, r *http.Request) {
	log.Println("autorolling")

	Res = random.Autoroller(Res)

	http.Redirect(w, r, "/stats", 302)
	return
}
func wipestats(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("stats wiped")

		apexdb.Wipestats()
		random.Rollcounter = 0
		http.Redirect(w, r, "/stats", 302)
	}
	return

}
func wipetourn(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("games wiped")
		cookie, err := r.Cookie("apextoken")
		if err != nil {
			log.Println("error retrieving cookie")
			http.Redirect(w, r, "/login", 302)
			return
		}
		sessid := cookie.Value
		apexdb.Wipetourn(sessid)

		http.Redirect(w, r, "/tournament", 302)
	}
	return

}
func tourneyapi(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("tourney started")
		tmpl := template.Must(template.ParseFiles("tourneyapi.html"))
		ip, ips, err := fromRequest(r)
		_ = ips
		if err != nil {
			log.Println("Error - IP Parse: ", err)
		}
		log.Printf("request ip: %v \n\n", ip)

		var Tourney apexdb.Tourney
		Tourney.T = apexdb.Seltourngames()
		Tourney.Activeusers = apexdb.Getactiveusers()

		if r.Method != http.MethodPost {
			//Tourney.G = Tourney.T[0].Games
			//Tourney.P = Tourney.T[0].Player

			focus := r.URL.Query().Get("focus")
			//r.URL.Query().Del("focus")

			log.Printf("after redir web param: %s \n\n", focus)
			Tourney.P = focus
			for _, elem := range Tourney.T {
				if elem.Player == focus {
					//log.Println(elem.Games)
					Tourney.G = elem.Games
				}
			}
			tmpl.Execute(w, Tourney)
			return
		}
		showdata := r.FormValue("showdata")

		var focusname string
		focusname = showdata
		Tourney.P = focusname

		if len(showdata) > 0 {
			log.Println("showdata in else if:", showdata)
			/*
				log.Println("r.URL.Path", r.URL.Path)
				u, err := url.Parse(r.URL.Path)
				if err != nil {
					log.Println("URL Parse failed: ", err)
				}
				q := u.Query()
				q.Del("focus")
				u.RawQuery = q.Encode()
			*/
			http.Redirect(w, r, "/redirtourn?focus="+showdata, 302) //redirect instead of executing template directly
			return
		}
		//log.Println("after redir showdata:", showdata, "focusname:", focusname)

		/*
			not executed if methodpost chk exists above
			focus := r.URL.Query().Get("focus")
			r.URL.Query().Del("focus")

			log.Println("after redir web param:", focus)

			for _, elem := range Tourney.T {
				if elem.Player == focus {
					log.Println(elem.Games)
					Tourney.G = elem.Games
				}
			}
			log.Println("before execute")
			tmpl.Execute(w, Tourney)
		*/
	}
	return
}
func tourney(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("tourney started")
		tmpl := template.Must(template.ParseFiles("tourney.html"))
		ip, ips, err := fromRequest(r)
		_ = ips
		if err != nil {
			log.Println("Error - IP Parse: ", err)
		}
		log.Printf("request ip: %v \n\n", ip)

		var Tourney apexdb.Tourney
		Tourney.T = apexdb.Seltourngames()
		Tourney.Activeusers = apexdb.Getactiveusers()

		if r.Method != http.MethodPost {
			//Tourney.G = Tourney.T[0].Games
			//Tourney.P = Tourney.T[0].Player

			focus := r.URL.Query().Get("focus")
			//r.URL.Query().Del("focus")

			log.Printf("after redir web param: %s \n\n", focus)
			Tourney.P = focus
			for _, elem := range Tourney.T {
				if elem.Player == focus {
					//log.Println(elem.Games)
					Tourney.G = elem.Games
				}
			}
			tmpl.Execute(w, Tourney)
			return
		}
		showdata := r.FormValue("showdata")
		player := r.FormValue("player")
		dmg := r.FormValue("dmg")
		place := r.FormValue("place")
		log.Println("form action received tourney:", showdata, player)
		log.Println("dmg,place:", dmg, place)

		var focusname string
		focusname = showdata
		Tourney.P = focusname

		if len(player) > 0 {

			log.Println("userform player >0")
			focusname = player
			err := apexdb.Loggame(player, dmg, place)
			if err != nil {
				Tourney.Errcode = err.Error()
			} else {
				Tourney.Errcode = ""
				Tourney.T = apexdb.Seltourngames()
				//http.Redirect(w, r, "/redir", 302) //redirect instead of executing template directly

			}
			focus := r.URL.Query().Get("focus")
			http.Redirect(w, r, "/redirtourn?focus="+focus, 302) //redirect instead of executing template directly
			return
		} else if len(showdata) > 0 {
			log.Println("showdata in else if:", showdata)
			/*
				log.Println("r.URL.Path", r.URL.Path)
				u, err := url.Parse(r.URL.Path)
				if err != nil {
					log.Println("URL Parse failed: ", err)
				}
				q := u.Query()
				q.Del("focus")
				u.RawQuery = q.Encode()
			*/
			http.Redirect(w, r, "/redirtourn?focus="+showdata, 302) //redirect instead of executing template directly
			return
		}
		//log.Println("after redir showdata:", showdata, "focusname:", focusname)

		/*
			not executed if methodpost chk exists above
			focus := r.URL.Query().Get("focus")
			r.URL.Query().Del("focus")

			log.Println("after redir web param:", focus)

			for _, elem := range Tourney.T {
				if elem.Player == focus {
					log.Println(elem.Games)
					Tourney.G = elem.Games
				}
			}
			log.Println("before execute")
			tmpl.Execute(w, Tourney)
		*/
	}
	return
}
func teams(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("teams started")
		ip, ips, err := fromRequest(r)
		_ = ips
		if err != nil {
			log.Println("Error - IP Parse: ", err)
		}
		log.Printf("request ip: %v \n\n", ip)

		cookie, err := r.Cookie("apextoken")
		if err != nil {
			log.Println("error retrieving cookie")
			http.Redirect(w, r, "/login", 302)
			return
		}
		sessid := cookie.Value
		tmpl := template.Must(template.ParseFiles("teams.html"))

		var user apexdb.User
		user.Teams = apexdb.Getbothteams()
		user.Activeusers = apexdb.Getactiveusers()

		if r.Method != http.MethodPost {
			tmpl.Execute(w, user)
			return
		}

		sw := r.FormValue("switch")
		rem := r.FormValue("remove")
		addu := r.FormValue("adduser")
		chgnm := r.FormValue("chgname")
		username := r.FormValue("username")
		pass := r.FormValue("pass")

		log.Println("form action received sw/rem/addu/chgnm:", sw, rem, addu, chgnm)

		if len(sw) > 1 {
			err := apexdb.Switchteams(sw)
			if err != nil {
				log.Println("switch teams error:", err)
			}
		} else if len(rem) > 0 {
			apexdb.Removeplayer(rem)
		} else if len(addu) > 0 {
			addplayertoteam(addu)
		} else if len(chgnm) > 0 {
			chgname(w, r, chgnm)
		} else if len(username) > 0 {
			apexdb.Updlogin(username, pass, sessid)
		}

		user.Teams = apexdb.Getbothteams()
		user.Activeusers = apexdb.Getactiveusers()

		//log.Println("teams:", user)
		//tmpl := template.Must(template.ParseFiles("teams.html"))
		tmpl.Execute(w, user)
	}
	return

}
func reroll1(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("reroll1 started")
		Res = random.Rollnewload(Res, 1)
		_ = Res
		//log.Println("reroll res:", Res)
		//log.Println("reroll1 res:", Res)
		http.Redirect(w, r, "/current", 302)
	}
	return
}
func reroll2(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("reroll2 started")
		Res = random.Rollnewload(Res, 2)
		_ = Res
		//log.Println("reroll res:", Res)
		http.Redirect(w, r, "/current", 302)
	}
	return
}
func reroll3(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("reroll3 started")
		Res = random.Rollnewload(Res, 3)
		_ = Res
		//log.Printf("reroll res:%+v", Res.Tchals)
		http.Redirect(w, r, "/current", 302)
	}
	return
}
func redirtourn(w http.ResponseWriter, r *http.Request) {
	focus := r.URL.Query().Get("focus")

	u, err := url.Parse(r.Header.Get("Referer"))
	if err != nil {
		log.Println("URL Parse failed: ", err)
	}
	q := u.Query()
	q.Del("focus")
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String()+"?focus="+focus, 302)
	return
}
func apires(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {

		ip, ips, err := fromRequest(r)
		_ = ips
		if err != nil {
			log.Println("Error - IP Parse: ", err)
		}
		cookie, err := r.Cookie("apextoken")
		if err != nil {
			log.Println("error retrieving cookie")
			http.Redirect(w, r, "/login", 302)
			return
		}
		sessid := cookie.Value
		username := apexdb.Getuserfromsess(sessid)
		//log.Println(r.Header)
		//log.Println("Read cookie:", r.Header.Get("Cookie"))

		Apimain := api.Reqtopapimatches(username)
		log.Printf("%v, viewcounter:%d \n", ip, viewcounter)
		log.Printf("Request executed \n\n")
		//log.Println("runtime heap allocation: ", runtime.ReadMemStats())

		tmpl := template.Must(template.ParseFiles("apires.html"))
		tmpl.Execute(w, Apimain)
	}
	return
}
func handler1(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		viewcounter++
		ip, ips, err := fromRequest(r)
		_ = ips
		if err != nil {
			log.Println("Error - IP Parse: ", err)
		}
		//log.Println(r.Header)
		//log.Println("Read cookie:", r.Header.Get("Cookie"))
		log.Printf("%v, viewcounter:%d \n", ip, viewcounter)
		log.Printf("Request executed \n\n")

		tmpl := template.Must(template.ParseFiles("forms.html"))
		tmpl.Execute(w, Res)
	}
	return
}

func helloServer(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "The server you were connecting to was disconnected or no longer in use.  Please try your request again or leave a message below")
	http.Redirect(w, r, "/apex", 302)
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

	//create bogus csrf val based on ip--this val not used anymore
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
	_ = userIPs
	if err != nil {
		return nil, "", fmt.Errorf("userip: %q did not convert to str", req.RemoteAddr)
	}
	//log.Println("ip parsed")
	//log.Println(userIP, userIPs)
	return userIP, ip, nil
}
