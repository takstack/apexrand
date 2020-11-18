package api

import (
	"apexrand/db"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

//Apipull main process to pull down api data
func Apipull() {
	apikey := getapikey()
	for {
		sl := []string{"full_send_deez", "jeffteeezy", "turbles", "theohmazingone", "lildongmanisme", "kringo506", "hochilinh"}
		for _, p := range sl {
			s := fmt.Sprintf("file/matchlist-%s", p)
			f, err := os.Create(s)
			if err != nil {
				log.Fatalln(err)
			}
			defer f.Close()
			getmatches(p, f, apikey)
		}

		/*
			f, err := os.Create("file/matchlist")
			if err != nil {
				log.Fatalln(err)
			}
			getmatches("full_send_deez", f)
		*/
		//readjson()

		time.Sleep(time.Second * 30)
		log.Println(Reqtopapimatches())
	}
}

func getmatches(p string, f *os.File, apikey string) {
	s := fmt.Sprintf("https://api.mozambiquehe.re/bridge?player=%s&platform=PS4&auth=%s&history=1&action=get", p, apikey)
	//req, err := http.NewRequest("GET", "https://api.mozambiquehe.re/bridge?player=pow_chaser&platform=PS4&auth=8uoPgHih7oHp8D8HXjuZ&history=1&action=info", nil)
	req, err := http.NewRequest("GET", s, nil)
	//_ = s
	if err != nil {
		log.Fatalln(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	n2, err := f.Write(body)
	sendapitodb(unmarjson(body))

	//log.Println(string(body))
	log.Println("wrote num bytes:", n2, p)
}
func sendapitodb(a apexdb.Apimain) {

	for _, elem := range a.Apiseries {

		elem.Username = apexdb.Getuserfrompsn(elem.Player)
		elem.Importdate = time.Now()
		elem.Stampconv = time.Unix(int64(elem.Timestamp), 0)

		n, err := strconv.Atoi(elem.UID[len(elem.UID)-6:])
		if err != nil {
			log.Fatalln("strconv err:", err)
		}
		elem.Userid = n

		for _, tracker := range elem.Tracker {
			if tracker.Key == apexdb.Cat.Cat2 {
				elem.Totdmg += tracker.Val
			}
			if tracker.Key == apexdb.Cat.Cat1 {
				elem.Totdmg += tracker.Val * 100
			}
			if tracker.Key == apexdb.Cat.Cat3 {
				elem.Totdmg += tracker.Val * 10
			}
			err = apexdb.Logtracker(elem, tracker)
			if err != nil {
				log.Fatalln("json err:", err)
			}
		}
		elem.Handi = apexdb.Gethandifromuser(elem.Username)
		elem.Adjdmg = int(float64(elem.Totdmg) * ((10000 - float64(elem.Handi)) / 10000))
		if len(elem.Throwaway) == 0 {
			//log.Println("len(elem.Throwaway)", len(elem.Throwaway))
			err = apexdb.Logapigame(elem)
		} else {
			//log.Println("len(elem.Throwaway)too long", len(elem.Throwaway))
		}
		if err != nil {
			log.Println("json err:", err)
		}
	}
}

func unmarjson(body []byte) apexdb.Apimain {
	var a apexdb.Apimain
	err := json.Unmarshal(body, &a.Apiseries)
	if err != nil {
		log.Fatalln("json err:", err)
	}

	for i := range a.Apiseries {
		li := &a.Apiseries[i]
		if len(li.Rawtracker) > 0 {
			//log.Println("in if")
			//log.Println("li.Rawtracker[0]:", string(li.Rawtracker[0]))
			switch li.Rawtracker[0] {
			case '{':
				//log.Println("in {")
				var c apexdb.Apievent
				if err := json.Unmarshal(li.Rawtracker, &c); err != nil {
					if err != nil {
						log.Fatalln("json err:", err)
					}
				}
				a.Apiseries[i].Throwaway = c.A1
			case '[':
				//log.Println("in [")
				var b []apexdb.Apitracker
				if err := json.Unmarshal(li.Rawtracker, &b); err != nil {
					if err != nil {
						log.Fatalln("json err:", err)
					}
				}
				//log.Println("t,n:", a.Apiseries[i].Player, t, n)
				a.Apiseries[i].Tracker = b

			default:
				log.Println("no case satisfied")
			}
		}
	}
	//log.Printf("%+v\n", a)
	return a
}

func jparse(body []byte) {
	var a apexdb.Apimain
	err := json.Unmarshal(body, &a.Apiseries)
	if err != nil {
		log.Fatalln("json err:", err)
	}
	log.Printf("%+v\n", a)
}
func readjson() {
	f, err := os.Open("file/matchlist-turbles")
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}

	unmarjson(body)
}

//Reqtopapimatches exp
func Reqtopapimatches() apexdb.Apimain {
	matchlist := apexdb.SeltopAPImatches()
	for _, match := range matchlist.Apiseries {
		var p = apexdb.Pulltracker{Val1: "0", Val2: "0", Val3: "0"}
		tracked := apexdb.Seltrackers(match.Userid, match.Stampconv)
		for _, elem := range tracked {
			if elem.Key == apexdb.Cat.Cat1 {
				p.Val1 = strconv.Itoa(elem.Val)
			}
			if elem.Key == apexdb.Cat.Cat2 {
				p.Val2 = strconv.Itoa(elem.Val)
			}
			if elem.Key == apexdb.Cat.Cat3 {
				p.Val3 = strconv.Itoa(elem.Val)
			}
		}
		match.Seltrackers = p
	}
	return matchlist
}
func getapikey() string {
	f, err := os.Open("/var/lib/api/apikey")
	if err != nil {
		log.Println("file open error:", err)
	}
	r, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("file open error:", err)
	}
	//log.Println("apikey:", string(r))
	return string(r)
}
