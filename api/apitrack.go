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
		now := time.Now()
		sl := []string{"full_send_deez", "jeffteeezy", "turbles", "theohmazingone", "lildongmanisme", "kringo506", "hochilinh"}
		for _, p := range sl {
			s := fmt.Sprintf("file/matchlist-%s", p)
			f, err := os.Create(s)
			if err != nil {
				log.Println(err)
				continue
			}
			defer f.Close()
			err = getmatches(p, f, apikey)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		/*
			f, err := os.Create("file/matchlist")
			if err != nil {
				log.Println(err)
			}
			getmatches("full_send_deez", f)
		*/
		//readjson()

		//log.Println(Reqtopapimatches())
		//Reqtopapimatches()

		//log.Println("pulled data in: ", time.Since(now))

		t := apexdb.Sellatestimport()
		if now.Sub(t) < time.Minute*30 {
			log.Printf("pulled data in: %v, time diff: %v, sleeping: %d secs ", time.Since(now).Round(time.Second/10), now.Sub(t).Round(time.Minute), 10)
			time.Sleep(time.Second * 10)
		} else {
			log.Printf("pulled data in: %v, time diff: %v, sleeping: %d secs ", time.Since(now).Round(time.Second/10), now.Sub(t).Round(time.Minute), 30)
			time.Sleep(time.Second * 30)
		}
	}
}

func getmatches(p string, f *os.File, apikey string) error {
	s := fmt.Sprintf("https://api.mozambiquehe.re/bridge?player=%s&platform=PS4&auth=%s&history=1&action=get", p, apikey)
	//req, err := http.NewRequest("GET", "https://api.mozambiquehe.re/bridge?player=pow_chaser&platform=PS4&auth=8uoPgHih7oHp8D8HXjuZ&history=1&action=info", nil)
	req, err := http.NewRequest("GET", s, nil)
	//_ = s
	if err != nil {
		log.Println(err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	n2, err := f.Write(body)
	if err != nil {
		log.Println(err)
		return err
	}
	var a apexdb.Apimain
	a, err = unmarjson(body)
	if err != nil {
		log.Println(err)
		return err
	}
	sendapitodb(a)
	if err != nil {
		log.Println(err)
		return err
	}
	_ = n2
	return nil
	//log.Println(string(body))
	//log.Println("wrote num bytes:", n2, p)
}
func sendapitodb(a apexdb.Apimain) {

	for _, elem := range a.Apiseries {

		elem.Username = apexdb.Getuserfrompsn(elem.Player)
		elem.Importdate = time.Now()
		elem.Stampconv = time.Unix(int64(elem.Timestamp), 0)

		n, err := strconv.Atoi(elem.UID[len(elem.UID)-6:])
		if err != nil {
			log.Println("strconv err:", err)
			continue
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
				elem.Totdmg += tracker.Val * 200
			}
			err = apexdb.Logtracker(elem, tracker)
			if err != nil {
				log.Println("db ins err:", err)
				time.Sleep(time.Second * 1)
				err = apexdb.Logtracker(elem, tracker)
				if err != nil {
					break
				}
			}
		}
		elem.Handi = apexdb.Gethandifromuser(elem.Username)
		elem.Adjdmg = int(float64(elem.Totdmg) * ((10000 - float64(elem.Handi)) / 10000))
		if len(elem.Throwaway) == 0 {
			//log.Println("len(elem.Throwaway)", len(elem.Throwaway))
			err = apexdb.Logapigame(elem)
			if err != nil {
				log.Println("db ins err:", err)
				continue
			}
		} else {
			//log.Println("len(elem.Throwaway)too long", len(elem.Throwaway))
		}

	}
}

func unmarjson(body []byte) (apexdb.Apimain, error) {
	var a apexdb.Apimain
	err := json.Unmarshal(body, &a.Apiseries)
	if err != nil {
		log.Println("json err:", err)
		return a, err
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
						log.Println("json err:", err)
						return a, err
					}
				}
				a.Apiseries[i].Throwaway = c.A1
			case '[':
				//log.Println("in [")
				var b []apexdb.Apitracker
				if err := json.Unmarshal(li.Rawtracker, &b); err != nil {
					if err != nil {
						log.Println("json err:", err)
						return a, err
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
	return a, nil
}

func jparse(body []byte) {
	var a apexdb.Apimain
	err := json.Unmarshal(body, &a.Apiseries)
	if err != nil {
		log.Println("json err:", err)
	}
	log.Printf("%+v\n", a)
}
func readjson() {
	f, err := os.Open("file/matchlist-turbles")
	if err != nil {
		log.Println(err)
	}
	body, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
	}

	unmarjson(body)
}

//Reqtopapimatches exp
func Reqtopapimatches(username string) apexdb.Apimain {
	matchlist := apexdb.SeltopAPImatches(username)
	for i := range matchlist.Apiseries {
		var p = apexdb.Pulltracker{Val1: "0", Val2: "0", Val3: "0"}
		//log.Println("match to find trackers for:", match.Userid, match.Stampconv)
		tracked := apexdb.Seltrackers(matchlist.Apiseries[i].Userid, matchlist.Apiseries[i].Stampconv)
		for _, elem := range tracked {
			//log.Println("request elem.Key == apexdb.Cat.Cat1", elem.Key, apexdb.Cat.Cat1)
			if elem.Key == apexdb.Cat.Cat1 {
				//log.Println("request in cat1")
				p.Val1 = strconv.Itoa(elem.Val)
			}
			if elem.Key == apexdb.Cat.Cat2 {
				p.Val2 = strconv.Itoa(elem.Val)
			}
			if elem.Key == apexdb.Cat.Cat3 {
				p.Val3 = strconv.Itoa(elem.Val)
			}
		}
		//log.Println("request p:", p)
		matchlist.Apiseries[i].Seltrackers = p
		matchlist.Apiseries[i].Stampconv = apexdb.Convertutc(matchlist.Apiseries[i].Stampconv)
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
