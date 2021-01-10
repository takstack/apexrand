package api

import (
	apexdb "apexrand/db"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

//used for api status
type jsonmap struct {
	Area Dets `json:"EU-West"`
}

//Dets used for api status
type Dets struct {
	Status   string `json:"Status"`
	Code     int    `json:"HTTPCode"`
	Resptime int    `json:"ResponseTime"`
	Tstamp   int    `json:"QueryTimestamp"`
}

//APIerr tracks last status of the api so if connection is failing, users can manually log games
var APIerr string
var lastpull time.Time

//Apipull main process to pull down api data
func Apipull() {
	apikey := getapikey()
	lastpull = time.Now()
	sleeptime := int(5)
	statuscounter := 0
	for {
		<-time.After(time.Second * time.Duration(sleeptime))

		//check api status json and continue if down
		status, err := decjsonmap(apikey)
		if status != "UP" || err != nil {
			if err != nil {
				log.Println("from decjsonmap err:", err)
			}
			if statuscounter%10 == 0 {
				log.Println("API Servers are: ", status, time.Since(lastpull).Round(time.Second/10),
					"since last pull, sleep: 30, statuscounter:", statuscounter)
			}
			statuscounter++
			APIerr = "CONNECTION FAILED... Manually log games at bottom of this page"
			sleeptime = 30
			continue
		}

		now := time.Now()
		sl := []string{"full_send_deez", "jeffteeezy", "turbles", "theohmazingone",
			"lildongmanisme", "kringo506", "hochilinh", "linh4tw"}

		//sl := []string{"lildongmanisme"}
		for _, p := range sl {
			s := fmt.Sprintf("file/matchlist-%s", p)
			f, err := os.Create(s)
			if err != nil {
				log.Println(err)
				continue
			}
			/*
				var platform string
				if p == "linh4tw" {
					platform = "X1"
				} else {
					platform = "PS4"
				}
			*/
			platform := apexdb.Getplatfrompsn(p)
			err = getmatches(p, platform, f, apikey)

			f.Close()
			if err != nil {
				APIerr = "CONNECTION FAILED... Manually log games at bottom of this page"
				log.Println(err, p)
				continue
			}
			APIerr = "Connection successful"
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
		lastpull = time.Now()
		statuscounter = 0
		t := apexdb.Sellatestimport()
		//log.Println("sellatestimport, now: ", t,now)
		if now.Sub(t) < time.Minute*30 {
			log.Printf("pulled data in: %v, time diff: %v, sleeping: %d secs ", time.Since(now).Round(time.Second/10), lastpull.Sub(t).Round(time.Second), 5)
			//time.Sleep(time.Second * 5)
			sleeptime = 5
		} else {
			log.Printf("pulled data in: %v, time diff: %v, sleeping: %d secs ", time.Since(now).Round(time.Second/10), lastpull.Sub(t).Round(time.Second), 30)
			//time.Sleep(time.Second * 30)
			sleeptime = 30
		}
	}
}
func decjsonmap(apikey string) (string, error) {
	s := fmt.Sprintf("https://api.mozambiquehe.re/servers?auth=%s", apikey)
	//req, err := http.NewRequest("GET", "https://api.mozambiquehe.re/bridge?player=pow_chaser&platform=PS4&auth=8uoPgHih7oHp8D8HXjuZ&history=1&action=info", nil)
	req, err := http.NewRequest("GET", s, nil)
	//_ = s
	if err != nil {
		fmt.Println("err decjsonmap newreq:", err)
		return "", errors.New("decjsonmap err newreq")
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != 200 {
		log.Println("statuscode: ", resp.StatusCode)
		return "", errors.New("Non-200 http response")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("err decjsonmap readall:", err)
		return "", errors.New("decjsonmap err readall")
	}
	fmt.Println("body:", string(body))
	var j map[string]jsonmap
	err = json.Unmarshal(body, &j)
	if err != nil {
		fmt.Println("err decjsonmap unmar:", err)
		return "", errors.New("decjsonmap err unmarshal")
	}
	return j["Mozambiquehere_StatsAPI"].Area.Status, nil
	//fmt.Println("body:", string(body))
	//fmt.Println("j unmarshaled", j["Mozambiquehere_StatsAPI"].Area.Status)
}

func getmatches(p string, platform string, f *os.File, apikey string) error {
	now := time.Now()
	s := fmt.Sprintf("https://api.mozambiquehe.re/bridge?player=%s&platform=%s&auth=%s&history=1&action=get", p, platform, apikey)
	//req, err := http.NewRequest("GET", "https://api.mozambiquehe.re/bridge?player=pow_chaser&platform=PS4&auth=8uoPgHih7oHp8D8HXjuZ&history=1&action=info", nil)
	req, err := http.NewRequest("GET", s, nil)
	//_ = s
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)

	if os.IsTimeout(err) {
		return errors.New("client.Timeout exceeded while awaiting headers")
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("statuscode: ", resp.StatusCode)
		return errors.New("Non-200 http response")
	}
	_ = now
	//log.Printf("API access: %v, %s", time.Since(now), p)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	/*
		n2, err := f.Write(body)
		if err != nil {
			log.Println(err)
			return err
		}
	*/
	var a apexdb.Apimain
	a, err = unmarjson(body)
	if err != nil {
		return err
	}
	//log.Println("getmatches a:", a)
	sendapitodb(a)
	if err != nil {
		return err
	}
	//_ = n2
	return nil
	//log.Println(string(body))
	//log.Println("wrote num bytes:", n2, p)
}
func sendapitodb(a apexdb.Apimain) {

	//getting list of all unix timestamps for user
	var uid int
	if len(a.Apiseries) > 0 {
		var err error
		uid, err = strconv.Atoi(a.Apiseries[0].UID[len(a.Apiseries[0].UID)-6:])
		if err != nil {
			log.Println("strconv err:", err)
		}

	}
	stamplist := apexdb.Selustamps(uid)

	//looping through each item in api results
	for _, elem := range a.Apiseries {

		//if the timestamp exists already or this api item is not a tracker, skip logging
		if !notindb(stamplist, elem.Timestamp) || len(elem.Throwaway) != 0 {
			continue
		}
		elem.Username = apexdb.Getuserfrompsn(elem.Player)
		//log.Println("in sendapitodb, notindb succeeded", elem.Username, elem.Timestamp)
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
				elem.Totdmg += tracker.Val * 250
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
		//check if char is allowed in tourn
		//elem.Inctourn = checkchar(elem.Legend)
		elem.Inctourn = true
		//log.Println("len(elem.Throwaway)", len(elem.Throwaway))
		//log.Println("elem.Importdate", elem.Importdate)
		err = apexdb.Logapigame(elem)
		if err != nil {
			log.Println("db ins err:", err)
			continue
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
						a.Apiseries[i].Throwaway = "abc" //c.A1
						return a, err
					}
				}
				//log.Println("rawtracker", string(li.Rawtracker))
				a.Apiseries[i].Throwaway = "abc" //c.A1
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
func notindb(stamplist []int, ts int) bool {
	if len(stamplist) <= 0 {
		//log.Println("notindb len(stamplist) is zero")
		return true
	}
	for _, elem := range stamplist {
		if elem == ts {
			//log.Println("notindb match found, returning false")
			return false
		}
	}
	return true
}
func checkchar(c string) bool {

	for _, elem := range apexdb.Char {
		if elem == c {
			//log.Println("character match ", c)
			return true
		}
	}
	return false
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
	now := time.Now()
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
	matchlist.Timesincepull = time.Since(lastpull).Round(time.Second / 100)
	matchlist.Timeselect = time.Since(now).Round(time.Millisecond / 100)
	log.Println("req top matches in: ", time.Since(now))
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
