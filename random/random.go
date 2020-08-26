package random

import (
	"apexrand/data"
	"log"
	//"math/bits"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

//Player exported
type Player struct {
	AllLoads  [3]Loadouts //3 players
	Zones1    [2]string   //2 zones
	Zones2    [2]string
	Z1str     string //holds joined zones
	Z2str     string
	Updhr     int //update time
	Updmin    int
	Updsec    int
	T1tmpchal []string
	T2tmpchal []string
	Tchals    []Teamchals
	T1str     string //holds joined team chals
	T2str     string
}

//Loadouts holds pair of loadouts for html parsing
type Loadouts struct {
	L1 Loadout
	L2 Loadout
}

//Teamchals holds both sets of team challenges for html parsing
type Teamchals struct {
	Tchal1 string
	Tchal2 string
}

//Loadout exported
type Loadout struct {
	Num  int      //player number
	Char string   //player name
	W1   string   //weapon 1
	W2   string   //weapon 2
	Chal []string //holds player chals
	Cstr string   //holds joined player chals
	Chct int      //challenge count
}

//Randints holds int order for randomized lists
type Randints struct {
	R1 [][]int
	R2 [][]int
	R3 [][]int
	R4 [][]int
}

var rollcounter int = 0

//Rollnewload handler for rolling team 1 or 2
func Rollnewload(res Player, mode int) Player {
	log.Println("New roll requested, mode:", mode)
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Println(err)
	}
	res.Updhr = time.Now().UTC().In(loc).Hour()
	res.Updmin = time.Now().UTC().Minute()
	res.Updsec = time.Now().UTC().Second()

	res = fillplayernums(res)
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //set new rand seed, only once for each web call

	switch mode {
	case 1:
		randlist1 := randomizelists(r)
		res = assignslots(randlist1, res, 1)
		res = handlewhammies(res, 1, r)
		res = numtchal(res, 1)
		res = convstrings(res, 1)
		res = equalizeslice(res)
	case 2:
		randlist2 := randomizelists(r)
		res = assignslots(randlist2, res, 2)
		res = handlewhammies(res, 2, r)
		res = numtchal(res, 2)
		res = convstrings(res, 2)
		res = equalizeslice(res)
	case 3:
		randlist1 := randomizelists(r)
		randlist2 := randomizelists(r)
		res = assignslots(randlist1, res, 1)
		res = assignslots(randlist2, res, 2)
		res = handlewhammies(res, 1, r)
		res = handlewhammies(res, 2, r)
		res = numtchal(res, 1)
		res = numtchal(res, 2)
		res = convstrings(res, 1)
		res = convstrings(res, 2)
		res = equalizeslice(res)
	}
	rollcounter++
	log.Println("rollcounter: ", rollcounter)
	//log.Println("res after reroll", res)
	return res
}

//joins strings for display as single element in html
func convstrings(res Player, team int) Player {
	//log.Println("tchal1:", res.Tchal1)
	//log.Println("tchal2:", res.Tchal2)
	switch team {
	case 1:
		res.T1str = strings.Join(res.T1tmpchal, ", ")
		res.Z1str = strings.Join([]string{res.Zones1[0], res.Zones1[1]}, ", ")
		for elem := range res.AllLoads {
			//log.Println("pchal1:", res.AllLoads[elem].L1.Chal)
			res.AllLoads[elem].L1.Chct = res.AllLoads[elem].L1.Chct + len(res.AllLoads[elem].L1.Chal) //track num challenges assigned
			res.AllLoads[elem].L1.Cstr = strings.Join(res.AllLoads[elem].L1.Chal, ", ")
			//log.Println("Team 1 player curse count", res.AllLoads[elem].L1.Num, res.AllLoads[elem].L1.Chct, "+", len(res.AllLoads[elem].L1.Chal))
		}
	case 2:
		res.T2str = strings.Join(res.T2tmpchal, ", ")
		res.Z2str = strings.Join([]string{res.Zones2[0], res.Zones2[1]}, ", ")
		for elem2 := range res.AllLoads {
			//log.Println("pchal2:", res.AllLoads[elem2].L2.Chal)
			res.AllLoads[elem2].L2.Chct = res.AllLoads[elem2].L2.Chct + len(res.AllLoads[elem2].L2.Chal) //track num challenges assigned
			res.AllLoads[elem2].L2.Cstr = strings.Join(res.AllLoads[elem2].L2.Chal, ", ")
			//log.Println("Team 2 player curse count", res.AllLoads[elem2].L2.Num, res.AllLoads[elem2].L2.Chct, "+", len(res.AllLoads[elem2].L2.Chal))
		}
	}
	return res
}

//adds player nums to empty player sl
func fillplayernums(res Player) Player {
	for i := 0; i < 3; i++ {
		res.AllLoads[i].L1.Num = i + 1
		res.AllLoads[i].L2.Num = i + 4
	}
	return res
}

//add randomized ints to lists for each roll
func randomizelists(r *rand.Rand) Randints {
	var ri Randints

	ri.R1 = fillrand(data.Chars, r)       //[][]string with random chars
	ri.R2 = fillrand(data.Weapons, r)     //[][]string with random weapons
	ri.R3 = fillrand(data.Zonesking, r)   //[][]string with random zonesking
	ri.R4 = fillrand(data.Zonesworlds, r) //[][]string with random zonesworlds
	return ri
}

//assigns the random vars to players
func assignslots(ri Randints, res Player, team int) Player {
	res = assignchars(ri.R1, res, team)
	res = assignweapons(ri.R2, res, team)
	res = assignzones1(ri.R3, res, team)
	res = assignzones2(ri.R4, res, team)
	return res
}

//clears and assigns player and team whammies
func handlewhammies(res Player, team int, r *rand.Rand) Player {
	res = clearwhammies(res, team)
	res = assignwhammies(res, team, r)
	return res
}

//creates random sl for assigning later
func fillrand(sl []string, r *rand.Rand) [][]int {
	//log.Println("beg fillrand. sl is:", sl, "len sl is:", len(sl))

	resSL := make([][]int, len(sl)) //sl to hold rand nums
	for elem := range resSL {
		//makes 2d sl for each elem and assigns num and rand
		resSL[elem] = make([]int, 2)
		resSL[elem][0] = elem
		resSL[elem][1] = genrand(r)
	}
	sort.SliceStable(resSL, func(i, j int) bool {
		return resSL[i][1] > resSL[j][1]
	})
	//log.Println("end fillrand. ressl is:", resSL)
	return resSL
}
func assignchars(randSL [][]int, res Player, team int) Player {
	switch team {
	case 1:
		for i := 0; i < 3; i++ {
			res.AllLoads[i].L1.Char = data.Chars[randSL[i][0]]
		}
	case 2:
		for i := 0; i < 3; i++ {
			res.AllLoads[i].L2.Char = data.Chars[randSL[i][0]]
		}
	}
	return res
}
func assignweapons(randSL [][]int, res Player, team int) Player {

	switch team {
	case 1:
		//log.Println("beg ass weapons, case 1", len(randSL))
		for i := 0; i < 3; i++ {
			res.AllLoads[i].L1.W1 = data.Weapons[randSL[i][0]]
			res.AllLoads[i].L1.W2 = data.Weapons[randSL[i+3][0]]
		}
	case 2:
		for i := 0; i < 3; i++ {
			res.AllLoads[i].L2.W1 = data.Weapons[randSL[i][0]]
			res.AllLoads[i].L2.W2 = data.Weapons[randSL[i+3][0]]
		}
	}
	return res
}
func assignzones1(randSL [][]int, res Player, team int) Player {
	switch team {
	case 1:
		res.Zones1[0] = data.Zonesking[randSL[0][0]]
	case 2:
		res.Zones2[0] = data.Zonesking[randSL[0][0]]
	}
	return res
}
func assignzones2(randSL [][]int, res Player, team int) Player {
	switch team {
	case 1:
		res.Zones1[1] = data.Zonesworlds[randSL[0][0]]
	case 2:
		res.Zones2[1] = data.Zonesworlds[randSL[0][0]]
	}
	return res
}
func assignwhammies(res Player, team int, r *rand.Rand) Player {
	threshold := 30 // team starting threshold
	thresh := 10    //team starting threshold
	var ichal []string
	var tchal []string

	//assign indv challenges here
	for i := 0; i < 3; i++ {

		if genrand(r) < thresh {
			ichal = append(ichal, "Move between towns backward")
		}
		if genrand(r) < thresh {
			ichal = append(ichal, "No Attachments")
		}
		if genrand(r) < thresh/2 {
			ichal = append(ichal, "No Shields")
		}
		if genrand(r) < thresh/2 {
			ichal = append(ichal, "No Backpack")
		}
		if genrand(r) < thresh/2 {
			ichal = append(ichal, "Crouch only (entire game)")
		}
		switch team {
		case 1:
			//log.Println("ichal3:", ichal)
			res.AllLoads[i].L1.Chal = ichal

		case 2:
			res.AllLoads[i].L2.Chal = ichal

		}
		ichal = nil
	}
	//assign team challenges here
	for {
		if genrand(r) < threshold {
			tchal = append(tchal, "No Attachments")
		}
		if genrand(r) < threshold {
			tchal = append(tchal, "No Shields")
		}
		if genrand(r) < threshold/10 {
			tchal = append(tchal, "No Backpack")
		}
		if genrand(r) < threshold {
			tchal = append(tchal, "A pirate's life (swap victim's box)")
		}
		if genrand(r) < threshold/10 {
			tchal = append(tchal, "Crouch only (entire game)")
		}
		if genrand(r) < threshold {
			tchal = append(tchal, "Can't open doors")
		}
		if genrand(r) < threshold+(threshold/4) {
			tchal = append(tchal, "Land Blind (close eyes and ping map)")
		}
		if genrand(r) < threshold/10 {
			tchal = append(tchal, "Heals Only (no guns/throwables)")
		}
		if genrand(r) < threshold/2 {
			tchal = append(tchal, "Four corners (land in different corners)")
		}
		if genrand(r) < threshold {
			tchal = append(tchal, "No jump balloons")
		}
		if genrand(r) < threshold {
			tchal = append(tchal, "Musical boxes (rotate all inventory with squad)")
		}
		if genrand(r) < threshold {
			tchal = append(tchal, "Your L1 buttons broke! (ult ok)")
		}
		if genrand(r) < threshold {
			tchal = append(tchal, "Your R1 buttons broke!")
		}
		if len(tchal) > 0 {
			break
		}
	}

	switch team {
	case 1:
		res.T1tmpchal = tchal
	case 2:
		res.T2tmpchal = tchal
	}
	//log.Println(res)
	return res
}

func clearwhammies(res Player, team int) Player {
	switch team {
	case 1:
		res.T1tmpchal = nil
		//clear placeholder character for non-rolled team
		for i := range res.T2tmpchal {
			if res.T2tmpchal[i] == "-" {
				res.T2tmpchal = res.T2tmpchal[0:i]
				break
			}
		}
		for i := 0; i < 3; i++ {
			res.AllLoads[i].L1.Chal = nil
		}
	case 2:
		res.T2tmpchal = nil
		//clear placeholder character for non-rolled team
		for i := range res.T1tmpchal {
			if res.T1tmpchal[i] == "-" {
				res.T1tmpchal = res.T1tmpchal[0:i]
				break
			}
		}
		for i := 0; i < 3; i++ {
			res.AllLoads[i].L2.Chal = nil
		}
	}
	res.Tchals = nil //slice of challenges gets rebuilt before return to handler
	return res
}
func genrand(r *rand.Rand) int {
	num := r.Intn(1000)
	//log.Println(num)
	return num
}

//makes slices same length, splits into Teamchals struct and assigns Tchals in Player
func equalizeslice(res Player) Player {
	t1 := res.T1tmpchal
	t2 := res.T2tmpchal
	l := len(t1) - len(t2)
	switch {
	case l < 0:
		for i := 0; i < int(math.Round(math.Abs(float64(l)))); i++ {
			t1 = append(t1, "-")
		}
	case l > 0:
		for i := 0; i < int(math.Round(math.Abs(float64(l)))); i++ {
			t2 = append(t2, "-")
		}
	default:
	}
	tc := res.Tchals
	for i := 0; i < len(t1); i++ {
		var c Teamchals
		c.Tchal1 = t1[i]
		c.Tchal2 = t2[i]
		tc = append(tc, c)
	}
	res.T1tmpchal = t1
	res.T2tmpchal = t2
	res.Tchals = tc
	return res
}

//just numbers team chals for team passed
func numtchal(res Player, team int) Player {
	t1 := res.T1tmpchal
	t2 := res.T2tmpchal

	if team == 1 {
		for i := range t1 {
			//log.Println("for 1 i:", i, t1[i])
			s := strconv.Itoa(i + 1)
			t1[i] = s + ". " + t1[i]
		}
	} else if team == 2 {
		for i := range t2 {
			//log.Println("for 2 i:", i, t2[i])
			s := strconv.Itoa(i + 1)
			t2[i] = s + ". " + t2[i]
		}
	} else {
		log.Fatalln("error - team not set for numtchal ")
	}
	res.T1tmpchal = t1
	res.T2tmpchal = t2
	return res
}
