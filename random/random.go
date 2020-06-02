package random

import (
	"apexrand/data"
	"log"
	"math/bits"
	"math/rand"
	"sort"
	"strings"
	"time"
)

//Player exported
type Player struct {
	Loadouts1 [3]Loadout
	Loadouts2 [3]Loadout
	Zones1    [2]string
	Zones2    [2]string
	Updmin    int
	Updsec    int
	Tchal1    []string
	Tchal2    []string
	T1str     string
	T2str     string
}

//Loadout exported
type Loadout struct {
	Num  int
	Char string
	W1   string
	W2   string
	Chal []string
	Cstr string
	Chct int //challenge count
}
var rollcounter int = 0
const maxInt = 1<<(bits.UintSize-1) - 1 // 1<<31 - 1 or 1<<63 - 1
//Rollnewload exp
func Rollnewload(res Player, mode int) Player {
	log.Println("New roll requested, mode:", mode)
	res.Updmin = time.Now().UTC().Minute()
	res.Updsec = time.Now().UTC().Second()
	res = fillplayernums(res)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	switch mode {
	case 1:
		randSL1 := fillrand(data.Chars, r) //[][]string with random chars
		res = assignchars(randSL1, res, 1)
		randSL2 := fillrand(data.Weapons, r) //[][]string with random weapons
		res = assignweapons(randSL2, res, 1)
		randSL3 := fillrand(data.Zonesking, r) //[][]string with random zonesking
		res = assignzones1(randSL3, res, 1)
		randSL4 := fillrand(data.Zonesworlds, r) //[][]string with random zonesworlds
		res = assignzones2(randSL4, res, 1)
		res = clearwhammies(res, 1)
		res = assignwhammies(res, 1, r)
		res = convstrings(res, 1)
	case 2:
		randSL1 := fillrand(data.Chars, r) //[][]string with random chars
		res = assignchars(randSL1, res, 2)
		randSL2 := fillrand(data.Weapons, r) //[][]string with random weapons
		res = assignweapons(randSL2, res, 2)
		randSL3 := fillrand(data.Zonesking, r) //[][]string with random zonesking
		res = assignzones1(randSL3, res, 2)
		randSL4 := fillrand(data.Zonesworlds, r) //[][]string with random zonesworlds
		res = assignzones2(randSL4, res, 2)
		res = clearwhammies(res, 2)
		res = assignwhammies(res, 2, r)
		res = convstrings(res, 2)
	case 3:
		randSL1 := fillrand(data.Chars, r) //[][]string with random chars
		res = assignchars(randSL1, res, 1)
		randSL2 := fillrand(data.Weapons, r) //[][]string with random weapons
		res = assignweapons(randSL2, res, 1)
		randSL3 := fillrand(data.Zonesking, r) //[][]string with random zonesking
		res = assignzones1(randSL3, res, 1)
		randSL4 := fillrand(data.Zonesworlds, r) //[][]string with random zonesworlds
		res = assignzones2(randSL4, res, 1)
		randSL5 := fillrand(data.Chars, r) //[][]string with random chars
		res = assignchars(randSL5, res, 2)
		randSL6 := fillrand(data.Weapons, r) //[][]string with random weapons
		res = assignweapons(randSL6, res, 2)
		randSL7 := fillrand(data.Zonesking, r) //[][]string with random zonesking
		res = assignzones1(randSL7, res, 2)
		randSL8 := fillrand(data.Zonesworlds, r) //[][]string with random zonesworlds
		res = assignzones2(randSL8, res, 2)
		res = clearwhammies(res, 1)
		res = clearwhammies(res, 2)
		res = assignwhammies(res, 1, r)
		res = assignwhammies(res, 2, r)
		res = convstrings(res, 1)
		res = convstrings(res, 2)
	}
	rollcounter++
	log.Println("rollcounter: ", rollcounter)
	//log.Println("res after reroll", res)
	return res
}
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
func fillplayernums(res Player) Player {
	for i := 0; i < 3; i++ {
		res.Loadouts1[i].Num = i + 1
		res.Loadouts2[i].Num = i + 4
	}
	return res
}
func fillrand(sl []string, r *rand.Rand) [][]int {
	//log.Println("beg fillrand. sl is:", sl, "len sl is:", len(sl))

	resSL := make([][]int, len(sl)) //sl to hold rand nums
	for elem := range resSL {
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
	threshold := 100
	switch team {
	case 1:
		for i := 0; i < 3; i++ {
			if genrand(r) < threshold {
				res.Loadouts1[i].Chal = append(res.Loadouts1[i].Chal, "No Attachments")
			}
			if genrand(r) < threshold {
				res.Loadouts1[i].Chal = append(res.Loadouts1[i].Chal, "No Shields")
			}
			if genrand(r) < 10 {
				res.Loadouts1[i].Chal = append(res.Loadouts1[i].Chal, "No Backpack")
			}
			if genrand(r) < threshold {
				res.Loadouts1[i].Chal = append(res.Loadouts1[i].Chal, "A pirate's life (swap boxes)")
			}
			if genrand(r) < 10 {
				res.Loadouts1[i].Chal = append(res.Loadouts1[i].Chal, "Crouch only (entire game)")
			}
			if genrand(r) < threshold {
				res.Loadouts1[i].Chal = append(res.Loadouts1[i].Chal, "No Throwables")
			}
			if genrand(r) < threshold {
				res.Loadouts1[i].Chal = append(res.Loadouts1[i].Chal, "Can't open doors")
			}
		}
		if genrand(r) < 200 {
			res.Tchal1 = append(res.Tchal1, "Land Blind")
		}
		if genrand(r) < 10 {
			res.Tchal1 = append(res.Tchal1, "HEALS ONLY!!!!")
		}
		if genrand(r) < 30 {
			res.Tchal1 = append(res.Tchal1, "Four corners")
		}
		if genrand(r) < threshold {
			res.Tchal1 = append(res.Tchal1, "No jump balloons")
		}
	case 2:
		for i := 0; i < 3; i++ {
			if genrand(r) < threshold {
				res.Loadouts2[i].Chal = append(res.Loadouts2[i].Chal, "No Attachments")
			}
			if genrand(r) < threshold {
				res.Loadouts2[i].Chal = append(res.Loadouts2[i].Chal, "No Shields")
			}
			if genrand(r) < 10 {
				res.Loadouts2[i].Chal = append(res.Loadouts2[i].Chal, "No Backpack")
			}
			if genrand(r) < threshold {
				res.Loadouts2[i].Chal = append(res.Loadouts2[i].Chal, "A pirate's life (swap boxes)")
			}
			if genrand(r) < 10 {
				res.Loadouts2[i].Chal = append(res.Loadouts2[i].Chal, "Crouch only (entire game)")
			}
			if genrand(r) < threshold {
				res.Loadouts2[i].Chal = append(res.Loadouts2[i].Chal, "No Throwables")
			}
			if genrand(r) < threshold {
				res.Loadouts2[i].Chal = append(res.Loadouts2[i].Chal, "Can't open doors")
			}
		}
		if genrand(r) < 200 {
			res.Tchal2 = append(res.Tchal2, "Land Blind")
		}
		if genrand(r) < 10 {
			res.Tchal2 = append(res.Tchal2, "Heals Only")
		}
		if genrand(r) < 30 {
			res.Tchal2 = append(res.Tchal2, "Four corners")
		}
		if genrand(r) < threshold {
			res.Tchal2 = append(res.Tchal2, "No jump balloons")
		}
	}
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
