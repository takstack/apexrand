package random

import (
	"apexrand/data"
	"log"
	//"math/bits"
	"math/rand"
	"sort"
	"strings"
	"time"
)

//Player exported
type Player struct {
	Loadouts1 [3]Loadout //3 players
	Loadouts2 [3]Loadout
	Zones1    [2]string //2 zones
	Zones2    [2]string
	Updhr     int
	Updmin    int
	Updsec    int
	Tchal1    []string //holds team chals
	Tchal2    []string
	T1str     string //holds joined team chals
	T2str     string
}

//Loadout exported
type Loadout struct {
	Num  int
	Char string
	W1   string
	W2   string
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

//const maxInt = 1<<(bits.UintSize-1) - 1 // 1<<31 - 1 or 1<<63 - 1

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
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch mode {
	case 1:
		randlist1 := randomizelists(r)
		res = assignslots(randlist1, res, 1)
		res = handlewhammies(res, 1, r)
		res = convstrings(res, 1)
	case 2:
		randlist2 := randomizelists(r)
		res = assignslots(randlist2, res, 2)
		res = handlewhammies(res, 2, r)
		res = convstrings(res, 2)
	case 3:
		randlist1 := randomizelists(r)
		randlist2 := randomizelists(r)
		res = assignslots(randlist1, res, 1)
		res = assignslots(randlist2, res, 2)
		res = handlewhammies(res, 1, r)
		res = handlewhammies(res, 2, r)
		res = convstrings(res, 1)
		res = convstrings(res, 2)
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
		res.T1str = strings.Join(res.Tchal1, ", ")
		for elem := range res.Loadouts1 {
			//log.Println("pchal1:", res.Loadouts1[elem].Chal)
			res.Loadouts1[elem].Chct = res.Loadouts1[elem].Chct + len(res.Loadouts1[elem].Chal)
			res.Loadouts1[elem].Cstr = strings.Join(res.Loadouts1[elem].Chal, ", ")
			log.Println("Team 1 player curse count", res.Loadouts1[elem].Num, res.Loadouts1[elem].Chct, "+", len(res.Loadouts1[elem].Chal))
		}
	case 2:
		res.T2str = strings.Join(res.Tchal2, ", ")
		for elem2 := range res.Loadouts2 {
			//log.Println("pchal2:", res.Loadouts2[elem2].Chal)
			res.Loadouts2[elem2].Chct = res.Loadouts2[elem2].Chct + len(res.Loadouts2[elem2].Chal)
			res.Loadouts2[elem2].Cstr = strings.Join(res.Loadouts2[elem2].Chal, ", ")
			log.Println("Team 2 player curse count", res.Loadouts2[elem2].Num, res.Loadouts2[elem2].Chct, "+", len(res.Loadouts2[elem2].Chal))
		}
	}
	return res
}

//adds player nums to empty player sl
func fillplayernums(res Player) Player {
	for i := 0; i < 3; i++ {
		res.Loadouts1[i].Num = i + 1
		res.Loadouts2[i].Num = i + 4
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
			res.Loadouts1[i].Char = data.Chars[randSL[i][0]]
		}
	case 2:
		for i := 0; i < 3; i++ {
			res.Loadouts2[i].Char = data.Chars[randSL[i][0]]
		}
	}
	return res
}
func assignweapons(randSL [][]int, res Player, team int) Player {

	switch team {
	case 1:
		//log.Println("beg ass weapons, case 1", len(randSL))
		for i := 0; i < 3; i++ {
			res.Loadouts1[i].W1 = data.Weapons[randSL[i][0]]
			res.Loadouts1[i].W2 = data.Weapons[randSL[i+3][0]]
		}
	case 2:
		for i := 0; i < 3; i++ {
			res.Loadouts2[i].W1 = data.Weapons[randSL[i][0]]
			res.Loadouts2[i].W2 = data.Weapons[randSL[i+3][0]]
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
	threshold := 75
	var ichal []string
	var tchal []string

	for i := 0; i < 3; i++ {

		if genrand(r) < threshold {
			ichal = append(ichal, "No Attachments")
		}
		if genrand(r) < threshold {
			ichal = append(ichal, "No Shields")
		}
		if genrand(r) < 10 {
			ichal = append(ichal, "No Backpack")
		}
		if genrand(r) < threshold {
			ichal = append(ichal, "A pirate's life (swap victim's box)")
		}
		if genrand(r) < 10 {
			ichal = append(ichal, "Crouch only (entire game)")
		}
		if genrand(r) < threshold {
			ichal = append(ichal, "No Throwables")
		}
		if genrand(r) < threshold {
			ichal = append(ichal, "Can't open doors")
		}

		switch team {
		case 1:
			//log.Println("ichal3:", ichal)
			res.Loadouts1[i].Chal = ichal

		case 2:
			res.Loadouts2[i].Chal = ichal

		}
		ichal = nil
	}
	if genrand(r) < 100 {
		tchal = append(tchal, "Land Blind (put trashcan on head and land the squad)")
	}
	if genrand(r) < 10 {
		tchal = append(tchal, "Heals Only (no guns/throwables)")
	}
	if genrand(r) < 30 {
		tchal = append(tchal, "Four corners (land in different corners)")
	}
	if genrand(r) < threshold {
		tchal = append(tchal, "No jump balloons")
	}
	if genrand(r) < threshold {
		tchal = append(tchal, "Musical boxes (rotate boxes with squad)")
	}
	if genrand(r) < threshold {
		tchal = append(tchal, "Your L1 buttons broke!")
	}
	switch team {
	case 1:
		res.Tchal1 = tchal
	case 2:
		res.Tchal2 = tchal
	}
	//log.Println(res)
	return res
}
func clearwhammies(res Player, team int) Player {
	switch team {
	case 1:
		res.Tchal1 = nil
		for i := 0; i < 3; i++ {
			res.Loadouts1[i].Chal = nil
		}
	case 2:
		res.Tchal2 = nil
		for i := 0; i < 3; i++ {
			res.Loadouts2[i].Chal = nil
		}
	}
	return res
}
func genrand(r *rand.Rand) int {
	num := r.Intn(1000)
	//log.Println(num)
	return num
}
