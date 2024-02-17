package town

import (
	"fmt"
	"math"
	"math/rand"
)

/*
Towns are basically anywhere that:
- has a population of NPCs that live there
- those NPCs have roles, homes they live in, and daily routines
- based on their role, NPCs will be moving during the day to complete tasks or be where they are supposed to be
- ? may be developed over time (population increase/decrease, buildings built or destroyed, etc)

This is a special type of room that is dynamic and will be much more "alive" than other rooms.
*/

type OccupationDist struct {
	// peasant jobs

	Unemployed int
	Slave      int
	Laborer    int
	Sailor     int
	Farmer     int
	Fisherman  int
	Hunter     int

	// crime - lower chance of these roles appearing

	Thief int

	// middle class jobs

	Merchant   int
	InnKeep    int
	TownGuard  int
	Soldier    int
	Blacksmith int
	TownCrier  int

	// upper class jobs

	WineMaker int
	Jeweler   int
	Doctor    int
	Academic  int
	Priest    int

	// elite class jobs

	Senator   int
	Noble     int
	Centurion int
}

// geological details of this town, which impacts the type of jobs the population can work
type Location struct {
	HasHarbor   bool
	HasFarmland bool
	HasForest   bool
}

type Town struct {
	BaseRoomID string
	BaseSize   int // the size (in tiles) of this town
	Population int // real population of NPCs we will generate to live in this town, not including major characters
	Location
	OccupationDistribution OccupationDist // distribution (percentages) of jobs held by the population
	WealthRating           int            // how wealthy the town is - influences the jobs distributed among the population
}

func CreateTown(population int, wealthRating int) Town {
	t := Town{
		Population:   population,
		WealthRating: wealthRating,
	}
	t.generateJobDist()
	fmt.Println(t.OccupationDistribution)
	return t
}

func (t *Town) PrintToConsole() {
	dist := t.OccupationDistribution
	peasantPop := dist.Unemployed + dist.Slave + dist.Laborer + dist.Sailor + dist.Farmer + dist.Fisherman + dist.Hunter + dist.Thief
	middlePop := dist.Merchant + dist.InnKeep + dist.TownGuard + dist.Soldier + dist.Blacksmith + dist.TownCrier
	upperPop := dist.WineMaker + dist.Jeweler + dist.Doctor + dist.Academic + dist.Priest
	elitePop := dist.Senator + dist.Noble + dist.Centurion
	fmt.Println("== Town info ==")
	fmt.Printf("Total population: %v\n", t.Population)
	fmt.Println(" * Peasant pop: ", peasantPop)
	fmt.Println("Unemployed: ", dist.Unemployed)
	fmt.Println("Slave: ", dist.Slave)
	fmt.Println("Laborer: ", dist.Laborer)
	fmt.Println("Sailor: ", dist.Sailor)
	fmt.Println("Farmer: ", dist.Farmer)
	fmt.Println("Fisherman: ", dist.Fisherman)
	fmt.Println("Hunter: ", dist.Hunter)
	fmt.Println("Thief: ", dist.Thief)
	fmt.Println(" * Middle class pop: ", middlePop)
	fmt.Println("Merchant: ", dist.Merchant)
	fmt.Println("Innkeep: ", dist.InnKeep)
	fmt.Println("Town Guard: ", dist.TownGuard)
	fmt.Println("Soldier: ", dist.Soldier)
	fmt.Println("Blacksmith: ", dist.Blacksmith)
	fmt.Println("Town Crier: ", dist.TownCrier)
	fmt.Println(" * Upper class pop: ", upperPop)
	fmt.Println("Wine Maker: ", dist.WineMaker)
	fmt.Println("Jeweler: ", dist.Jeweler)
	fmt.Println("Doctor: ", dist.Doctor)
	fmt.Println("Academic: ", dist.Academic)
	fmt.Println("Priest: ", dist.Priest)
	fmt.Println(" * Elite class pop: ", elitePop)
	fmt.Println("Senator: ", dist.Senator)
	fmt.Println("Noble: ", dist.Noble)
	fmt.Println("Centurion: ", dist.Centurion)
}

func (t *Town) generateJobDist() {
	popLimit := t.Population
	currentPop := 0
	wealthRating := t.WealthRating
	//location := t.Location

	dist := OccupationDist{}
	numElites := 0
	numUpper := 0
	numMiddle := 0

	// start by deciding how many elites, if any
	if wealthRating >= 4 {
		// if there is enough wealth, there will be at least 1% or 1 elite
		numElites = int(max(math.Ceil(float64(popLimit)/100), 1))
		excessWealth := wealthRating - 4 // for every level above 4, there will be 1% or 1 more elite
		if excessWealth > 0 {
			numElites += int(max(math.Ceil(float64(popLimit)/100), 1)) * excessWealth
		}
		for i := 0; i < numElites; i++ {
			switch rand.Intn(3) {
			case 0:
				dist.Senator++
			case 1:
				dist.Noble++
			case 2:
				dist.Centurion++
			}
			currentPop++
			if currentPop == popLimit {
				break
			}
		}
	}
	if currentPop == popLimit {
		t.OccupationDistribution = dist
		return
	}
	// upper class - should be slightly more than elites, but still not many
	if wealthRating >= 3 {
		// baseline of 5% upper class, or at least 1
		numUpper = int(max(math.Ceil(float64(popLimit)/20), 1))
		excessWealth := wealthRating - 3
		if excessWealth > 0 {
			numUpper += int(max(math.Ceil(float64(popLimit)/100), 1)) * excessWealth
		}
		for i := 0; i < numUpper; i++ {
			switch rand.Intn(5) {
			case 0:
				dist.WineMaker++
			case 1:
				dist.Jeweler++
			case 2:
				dist.Doctor++
			case 3:
				dist.Academic++
			case 4:
				dist.Priest++
			}
			currentPop++
			if currentPop == popLimit {
				break
			}
		}
	}
	if currentPop == popLimit {
		t.OccupationDistribution = dist
		return
	}
	// middle class - should be relatively larger than upper
	if wealthRating >= 2 {
		// baseline of 15% middle class, or at least 1
		numMiddle = int(max(math.Ceil(float64(popLimit)*0.15), 1))
		excessWealth := wealthRating - 2
		if excessWealth > 0 {
			numMiddle += int(max(math.Ceil(float64(popLimit)/100), 1)) * excessWealth
		}
		for i := 0; i < numMiddle; i++ {
			switch rand.Intn(5) {
			case 0:
				dist.Merchant++
			case 1:
				// these are roles that will have actual shops - we don't need 10 blacksmiths in a single town
				if rand.Intn(2) == 0 {
					dist.InnKeep++
				} else {
					dist.Blacksmith++
				}
			case 2:
				dist.TownGuard++
			case 3:
				dist.Soldier++
			}
			currentPop++
			if currentPop == popLimit {
				break
			}
		}
		dist.TownCrier++ // only one town crier per town
	}
	if currentPop == popLimit {
		t.OccupationDistribution = dist
		return
	}
	// peasant class - the rest of the population will be this
	// let's assume elites and upper class have some slaves
	slaveCount := (numElites * 2) + (numUpper / 2)
	if slaveCount < popLimit-currentPop {
		dist.Slave += slaveCount
		currentPop += slaveCount
	} else {
		dist.Slave += popLimit - currentPop
		currentPop = popLimit

		t.OccupationDistribution = dist
		return
	}
	for currentPop < popLimit {
		switch rand.Intn(6) {
		case 0:
			if rand.Intn(3) == 0 {
				dist.Thief++
			} else {
				dist.Unemployed++
			}
		case 1:
			dist.Laborer++
		case 2:
			dist.Sailor++
		case 3:
			dist.Farmer++
		case 4:
			dist.Fisherman++
		case 5:
			dist.Hunter++
		}
		currentPop++
		if currentPop == popLimit {
			break
		}
	}
	t.OccupationDistribution = dist
}

/*

Unemployed int
Slave      int
Laborer    int
Sailor     int
Farmer     int
Fisherman  int
Hunter     int

*/
