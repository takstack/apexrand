package main

import (
	//"logger"
	"apexrand/api"
	apexdb "apexrand/db"
	"apexrand/random"

	//"bytes"
	//"encoding/json"
	"fmt"
	"html/template"

	//"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"

	//"runtime"
	//"bufio"
	//"os"
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
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,

		Addr: ":80",
	}
	//apexdb.Opendb()
	apexdb.Insvarfromfile()
	apexdb.Inscursefromfile()
	//apexdb.Insuserfromfile() //disabled to allow for email to be used for testing

	//apexdb.Sethandicap() //set new handicaps based on closed tourney
	//apexdb.Writetourngamescsv2()
	//apexdb.Delallsess() leave all sessions open for now
	go api.Apipull()

	log.Println("reminder: set tournament time in loggame, logmangames params if in tournament")
	http.HandleFunc("/reg", reg)
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
	http.HandleFunc("/roulette", roulette)
	http.HandleFunc("/login", login)
	http.HandleFunc("/waitconf", waitconf)
	http.HandleFunc("/trackers", trackersapi)
	http.HandleFunc("/confirm", confirm)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/teams", teams)
	http.HandleFunc("/friends", friends)
	http.HandleFunc("/manage", manage)
	http.HandleFunc("/user", user)
	http.HandleFunc("/tournament", tourneyapi)
	http.HandleFunc("/loggame", loggame)
	http.HandleFunc("/wipetourn", wipetourn)
	http.HandleFunc("/redirtourn", redirtourn)
	http.HandleFunc("/redirtrackers", redirtrackers)
	http.HandleFunc("/apires", apires)
	http.HandleFunc("/", helloServer)

	//http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	srv.SetKeepAlivesEnabled(false)
	log.Fatalln(srv.ListenAndServe())

}

//func for shopbot
func shopbot(w http.ResponseWriter, r *http.Request) {
	log.Println("shopbot started")
	tmpl := template.Must(template.ParseFiles("static/html/shopbot.html"))
	_, _, _ = fromRequest(r)

	focus := r.URL.Query().Get("key")
	if focus == "shop2020getit" {
		tmpl.Execute(w, "")
		//texting disabled due to no longer in use
		//sendtxts()
		log.Println("send texts should have been executed but disabled for now")
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func getstats(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if !validsess {
		return
	}
	log.Println("stats started")
	//_, _, _ = fromRequest(r)

	var st apexdb.Stats
	st.Cursecount = apexdb.Selcursestats()
	st.Playercount = apexdb.Selplayerstats()
	st.Totalrolls = apexdb.Getnumrolls()
	st.Thresh = random.Thresh
	tmpl := template.Must(template.ParseFiles("static/html/stats.html"))
	tmpl.Execute(w, st)
}

func rollauto(w http.ResponseWriter, r *http.Request) {
	log.Println("autorolling")

	Res = random.Autoroller(Res)

	http.Redirect(w, r, "/stats", http.StatusFound) //instead of 302

}
func wipestats(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("stats wiped")

		apexdb.Wipestats()
		random.Rollcounter = 0
		http.Redirect(w, r, "/stats", http.StatusFound)
	}

}
func wipetourn(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("games wiped")
		cookie, err := r.Cookie("apextoken")
		if err != nil {
			log.Println("error retrieving cookie")
			http.Redirect(w, r, "/login", http.StatusFound)

		}
		sessid := cookie.Value
		apexdb.Wipetourn(sessid)

		http.Redirect(w, r, "/tournament", http.StatusFound)
	}

}
func tourneyapi(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if !validsess {
		return
	}

	log.Println("tourney started")
	tmpl := template.Must(template.ParseFiles("static/html/tourneyapi.html"))
	//_, _, _ = fromRequest(r)

	var Tourney apexdb.Tourney
	Tourney.T = apexdb.Seltourngames()
	Tourney.Activeusers = apexdb.Getactiveusers()

	if r.Method != http.MethodPost {
		//Tourney.G = Tourney.T[0].Games
		//Tourney.P = Tourney.T[0].Player

		focus := r.URL.Query().Get("focus")
		Tourney.Errcode = r.URL.Query().Get("err")
		//r.URL.Query().Del("focus")

		log.Printf("after redir web param: %s \n\n", focus)
		Tourney.P = focus
		for _, elem := range Tourney.T {
			if elem.Player == focus {
				//log.Println(elem.Games)
				Tourney.G = elem.Games
			}
		}
		Tourney.APIerr = api.APIerr
		err := tmpl.Execute(w, Tourney)
		if err != nil {
			log.Println("roulette exec error")
		}
		Tourney.Errcode = ""
		return
	}
	showdata := r.FormValue("showdata") //selected to show players games

	var focusname string
	focusname = showdata
	Tourney.P = focusname

	//to log a game for that player
	player := r.FormValue("player")
	field1 := r.FormValue("field1")
	field2 := r.FormValue("field2")
	field3 := r.FormValue("field3")

	log.Println("form action received tourney:", showdata, player)
	log.Println("field1, field2, field3:", field1, field2, field3)

	if len(player) > 0 {

		log.Println("userform player >0")
		focusname = player
		err := apexdb.Logmanualgame(player, field1, field2, field3)
		if err != nil {
			Tourney.Errcode = err.Error()
			log.Println("logmanualgame error: ", Tourney.Errcode)

		} else {
			Tourney.Errcode = ""
			Tourney.T = apexdb.Seltourngames()
			//http.Redirect(w, r, "/redir", http.StatusFound) //redirect instead of executing template directly

		}
		focus := r.URL.Query().Get("focus")
		http.Redirect(w, r, "/redirtourn?focus="+focus+"&err="+Tourney.Errcode, http.StatusFound) //redirect instead of executing template directly
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
		http.Redirect(w, r, "/redirtourn?focus="+showdata+"&err="+Tourney.Errcode, http.StatusFound) //redirect instead of executing template directly
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
func loggame(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if !validsess {
		return
	}

	log.Println("loggame started")
	tmpl := template.Must(template.ParseFiles("static/html/loggame.html"))
	//_, _, _ = fromRequest(r)

	var Tourney apexdb.Tourney
	Tourney.Activeusers = apexdb.Getactiveusers()

	if r.Method != http.MethodPost {
		tmpl.Execute(w, Tourney)
		return
	}

	//to log a game for that player
	player := r.FormValue("player")
	field1 := r.FormValue("field1")
	field2 := r.FormValue("field2")
	field3 := r.FormValue("field3")

	log.Println("field1, field2, field3:", field1, field2, field3)

	if len(player) > 0 {
		log.Println("userform player >0")
		err := apexdb.Logmanualgame(player, field1, field2, field3)
		if err != nil {
			Tourney.Errcode = err.Error()
			log.Println("logmanualgame error: ", Tourney.Errcode)

		} else {
			Tourney.Errcode = ""
			Tourney.T = apexdb.Seltourngames()
			//http.Redirect(w, r, "/redir", http.StatusFound) //redirect instead of executing template directly

		}
		http.Redirect(w, r, "/loggame"+"&err="+Tourney.Errcode, http.StatusFound) //redirect instead of executing template directly
		return
	}

}
func trackersapi(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if !validsess {
		return
	}
	log.Println("trackers started")
	//_, _, _ = fromRequest(r)

	if r.Method != http.MethodPost {
		focus := r.URL.Query().Get("focus")
		users := apexdb.Getactiveusers()
		games := apexdb.Seltourntrackers(focus)
		//log.Println("active users", users)
		//log.Println("focus, games", focus, games)

		//log.Println("users sl", sl)

		Data := struct {
			G []apexdb.Game
			U []apexdb.Onlineuser
			T apexdb.Tournvar //tourninfo
		}{
			G: games,
			U: users,
			T: apexdb.Tvar,
		}
		//log.Println("data.U", Data.U)

		tmpl := template.Must(template.ParseFiles("static/html/trackersapi.html"))

		//set focus to get trackers-----------------------------------------------------------------------------------------------------

		//data.g = apexdb.Seltourntrackers(focus)
		//data.u = apexdb.Getactiveusers()
		//r.URL.Query().Del("focus")

		log.Printf("after redir web param: %s \n\n", focus)

		err := tmpl.Execute(w, Data)
		if err != nil {
			log.Println("roulette exec error")
		}
		return //check============================================================================================================================
	}
	showdata := r.FormValue("showdata") //selected to show players games
	http.Redirect(w, r, "/redirtrackers?focus="+showdata, http.StatusFound)
}
func teams(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("teams started")
		//_, _, _ = fromRequest(r)

		cookie, err := r.Cookie("apextoken")
		if err != nil {
			log.Println("error retrieving cookie")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		sessid := cookie.Value
		tmpl := template.Must(template.ParseFiles("static/html/teams.html"))

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
		//tmpl := template.Must(template.ParseFiles("static/html/teams.html"))
		tmpl.Execute(w, user)
	}

}
func reroll1(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("reroll1 started")
		Res = random.Rollnewload(Res, 1)
		_ = Res
		//log.Println("reroll res:", Res)
		//log.Println("reroll1 res:", Res)
		http.Redirect(w, r, "/roulette", http.StatusFound)
	}

}
func reroll2(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("reroll2 started")
		Res = random.Rollnewload(Res, 2)
		_ = Res
		//log.Println("reroll res:", Res)
		http.Redirect(w, r, "/roulette", http.StatusFound)
	}

}
func reroll3(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("reroll3 started")
		Res = random.Rollnewload(Res, 3)
		_ = Res
		//log.Printf("reroll res:%+v", Res.Tchals)
		http.Redirect(w, r, "/roulette", http.StatusFound)
	}

}
func redirtourn(w http.ResponseWriter, r *http.Request) {
	errcode := r.URL.Query().Get("err")
	if errcode != "" {
		focus := r.URL.Query().Get("focus")
		u1, err := url.Parse(r.Header.Get("Referer"))
		if err != nil {
			log.Println("URL Parse failed, err: ", err)
		}
		q1 := u1.Query()
		q1.Del("err")
		q1.Del("focus")
		u1.RawQuery = q1.Encode()
		log.Println("u1.string: ", u1.String())
		http.Redirect(w, r, u1.String()+"?focus="+focus+"&err="+errcode, http.StatusFound)
		return
	}

	focus := r.URL.Query().Get("focus")

	u, err := url.Parse(r.Header.Get("Referer"))
	if err != nil {
		log.Println("URL Parse failed: ", err)
	}
	q := u.Query()
	q.Del("err")
	q.Del("focus")
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String()+"?focus="+focus, http.StatusFound)

}
func redirtrackers(w http.ResponseWriter, r *http.Request) {
	focus := r.URL.Query().Get("focus")

	u, err := url.Parse(r.Header.Get("Referer"))
	if err != nil {
		log.Println("URL Parse failed: ", err)
	}
	q := u.Query()
	q.Del("focus")
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String()+"?focus="+focus, http.StatusFound)

}
func apires(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if !validsess {
		return
	}
	//_, _, _ = fromRequest(r)

	cookie, err := r.Cookie("apextoken")
	if err != nil {
		log.Println("error retrieving cookie")
		http.Redirect(w, r, "/login", http.StatusFound)

	}
	sessid := cookie.Value
	username := apexdb.Getuserfromsess(sessid)
	//log.Println(r.Header)
	//log.Println("Read cookie:", r.Header.Get("Cookie"))

	Apimain := api.Reqlatesttrackers(username)
	//log.Println("runtime heap allocation: ", runtime.ReadMemStats())

	tmpl := template.Must(template.ParseFiles("static/html/apires.html"))
	tmpl.Execute(w, Apimain)

}
func user(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if !validsess {
		return
	}
	//_, _, _ = fromRequest(r)

	tmpl := template.Must(template.ParseFiles("static/html/user.html"))
	tmpl.Execute(w, nil)

}
func handler1(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if !validsess {
		return
	}
	log.Println("handler1 started")
	viewcounter++
	//_, _, _ = fromRequest(r)
	//log.Println(r.Header)
	//log.Println("Read cookie:", r.Header.Get("Cookie"))

	cookie, err := r.Cookie("apextoken")
	if err != nil {
		log.Println("error retrieving cookie")
		http.Redirect(w, r, "/login", http.StatusFound)

	}
	sessid := cookie.Value
	username := apexdb.Getuserfromsess(sessid)
	api.H.Playername = apexdb.Getplayeridfromuser(username)
	api.H.Platform = apexdb.Getplatfromuser(username)

	log.Printf("viewcounter-handler1:%d \n", viewcounter)

	tmpl := template.Must(template.ParseFiles("static/html/home.html"))
	err = tmpl.Execute(w, api.H)
	if err != nil {
		log.Println("handler1 exec error")
	}
}
func roulette(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if !validsess {
		return
	}
	log.Println("roulette handler started")
	viewcounter++
	//_, _, _ = fromRequest(r)

	//log.Println(r.Header)
	//log.Println("Read cookie:", r.Header.Get("Cookie"))

	//log.Printf("viewcounter:%d \n", viewcounter)
	//log.Printf("Request executed \n\n")

	tmpl := template.Must(template.ParseFiles("static/html/roulette.html"))
	err := tmpl.Execute(w, Res)
	if err != nil {
		log.Println("roulette exec error")
	}

}
func friends(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/html/friends.html"))
	tmpl.Execute(w, nil)
}
func manage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/html/manage.html"))
	tmpl.Execute(w, nil)
}
func helloServer(w http.ResponseWriter, r *http.Request) {
	log.Println("helloserver started, redirect to /apex")
	//fmt.Fprintf(w, "The server you were connecting to was disconnected or no longer in use.  Please try your request again or leave a message below")
	http.Redirect(w, r, "/apex", http.StatusFound)
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
	log.Println("ip parsed: ", ip)
	//log.Println(userIP, userIPs)
	return userIP, ip, nil
}

/*
func tourney(w http.ResponseWriter, r *http.Request) {
	validsess := chkvalidsession(w, r)
	if validsess {
		log.Println("tourney started")
		tmpl := template.Must(template.ParseFiles("static/html/tourney.html"))
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
				//http.Redirect(w, r, "/redir", http.StatusFound) //redirect instead of executing template directly

			}
			focus := r.URL.Query().Get("focus")
			http.Redirect(w, r, "/redirtourn?focus="+focus, http.StatusFound) //redirect instead of executing template directly
			return
		} else if len(showdata) > 0 {
			log.Println("showdata in else if:", showdata)
*/
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
/*
		http.Redirect(w, r, "/redirtourn?focus="+showdata, http.StatusFound) //redirect instead of executing template directly
		return
	}
	//log.Println("after redir showdata:", showdata, "focusname:", focusname)
*/
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
/*
	}
	return
}
*/

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

/*
//func for shopbot
func sendtxts() {
	key := getemailkey()
	auth := smtp.PlainAuth("", key[0], key[1], "smtp.gmail.com")
	//log.Println("key", key)

	tonums := getphonelist()
	for _, addr := range tonums {
		//concurrent async to allow the http page to serve while smtp sends in background, avoiding timeouts
		go sendtext(key, auth, addr)
		go sendtext2(key, auth, addr)
	}
}

func sendtext2(key []string, auth smtp.Auth, addr string) {
	//sending without "To:" will make it bcc:
	msg := []byte("To:" + addr + "\r\n" + "Subject:Link\r\n" + "\r\n" + "https://www.bestbuy.com/cart?loc=PS5%20restock%20and%20other%20tech+gaming%20finds/deals&ref=198&cmp=RMX&acampID=0")

	err := smtp.SendMail("smtp.gmail.com:587", auth, key[0], []string{addr}, msg)
	if err != nil {
		log.Println(err)
	}
	log.Println("smtp executed: ", addr)
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


*/
/*
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
*/
