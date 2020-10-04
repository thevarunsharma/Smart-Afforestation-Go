package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"
)

// TreeIdx tree index
var TreeIdx map[string]uint8

// struct for TreeInfo
type info struct {
	PlantSpecies   string  `json:"plantSpecies"`
	CommonName     string  `json:"commonName"`
	LifeForm       string  `json:"lifeForm"`
	ZoneI          float64 `json:"zoneI"`
	ZoneII         float64 `json:"zoneII"`
	ZoneIII        float64 `json:"zoneIII"`
	ZoneIV         float64 `json:"zoneIV"`
	CanopyDiameter int     `json:"canopyDiameter"`
	Utility        float64 `json:"utility"`
	Cost           int     `json:"cost"`
	Area           int     `json:"area"`
}

// TreeInfo tree info
var TreeInfo []*info

// treeTypeCounts
var treeTypesCount int

// TreePlanter struct
type TreePlanter struct {
	numOfChromosomes int
	areaLimit        int
	costLimit        int
	population       int
	level            string
	zone             string
	score            []float64
	area             []int
	cost             []int
	sampleSet        []int
	bestFit          float64
	chromosomes      [][]bool
	bestChromosome   []bool
	totalFit         []float64
	lastFit          float64
	repCount         int
	minc             float64
	maxc             float64
}

func getAqiRange(AQI int) (string, string) {
	switch {
	case AQI <= 50:
		return "Good", "Zone IV"
	case 50 < AQI && AQI <= 100:
		return "Moderate", "Zone IV"
	case 100 < AQI && AQI <= 150:
		return "Unhealthy for sensitive groups", "Zone III"
	case 150 < AQI && AQI <= 200:
		return "Unhealthy", "Zone III"
	case 200 < AQI && AQI <= 300:
		return "Very Unhealthy", "Zone II"
	default:
		return "Hazardous", "Zone I"
	}
}

func (planter *TreePlanter) initScoreAreaCost(w1, w2 float64) {
	planter.score = make([]float64, treeTypesCount)
	planter.area = make([]int, treeTypesCount)
	planter.cost = make([]int, treeTypesCount)

	for i := 0; i < treeTypesCount; i++ {
		// f = w1*p1 + w2
		planter.area[i] = TreeInfo[i].Area
		planter.cost[i] = TreeInfo[i].Cost
		var zoneScore float64 = TreeInfo[i].ZoneII
		switch planter.zone {
		case "Zone I":
			zoneScore = TreeInfo[i].ZoneI
		case "Zone II":
			zoneScore = TreeInfo[i].ZoneII
		case "Zone III":
			zoneScore = TreeInfo[i].ZoneIII
		case "Zone IV":
			zoneScore = TreeInfo[i].ZoneIV
		}
		utilScore := TreeInfo[i].Utility
		planter.score[i] = (w1*zoneScore + w2*utilScore) * float64(planter.area[i])
	}
}

func (planter *TreePlanter) getSamplingSet() {
	planter.minc, planter.maxc = math.Inf(1), math.Inf(-1)
	for i := 0; i < treeTypesCount; i++ {
		count := math.Min(float64(planter.costLimit/planter.cost[i]),
			float64(planter.areaLimit/planter.area[i]))
		planter.minc = math.Min(count, planter.minc)
		planter.maxc = math.Max(count, planter.maxc)
		filler := make([]int, int(count))
		for j := range filler {
			filler[j] = i
		}
		planter.sampleSet = append(planter.sampleSet, filler...)
	}

}

func randInt(min int, max int) int {
	return rand.Intn(max-min+1) + min
}

func (planter *TreePlanter) initChromosomes() {
	planter.chromosomes = make([][]bool, planter.numOfChromosomes)
	for i := range planter.chromosomes {
		planter.chromosomes[i] = make([]bool, len(planter.sampleSet))
	}
	for i := range planter.chromosomes {
		choose := randInt(int(planter.minc), int(planter.maxc))
		idx := rand.Perm(len(planter.chromosomes[i]))[:choose]
		for _, j := range idx {
			planter.chromosomes[i][j] = true
		}
	}
}

func (planter *TreePlanter) init(
	AQI,
	areaLimit,
	costLimit,
	population int,
	numOfChromosomes ...int) {
	if len(numOfChromosomes) > 0 {
		planter.numOfChromosomes = numOfChromosomes[0]
	} else {
		planter.numOfChromosomes = 20
	}

	planter.areaLimit = areaLimit
	planter.costLimit = costLimit
	planter.population = population
	planter.level, planter.zone = getAqiRange(AQI)

	planter.initScoreAreaCost(20, 5)
	planter.getSamplingSet()
	planter.initChromosomes()

	planter.bestFit = math.Inf(-1)
	planter.totalFit = make([]float64, planter.numOfChromosomes)
	for i := range planter.totalFit {
		planter.totalFit[i] = math.Inf(-1)
	}

	planter.lastFit = math.Inf(-1)
}

func (planter *TreePlanter) getFitness(idx int) float64 {
	chromosome := planter.chromosomes[idx]
	N := len(chromosome)
	totalCost := 0
	totalArea := 0
	for i := 0; i < N; i++ {
		if chromosome[i] {
			totalCost += planter.cost[planter.sampleSet[i]]
			totalArea += planter.area[planter.sampleSet[i]]
		}
	}

	if totalCost > planter.costLimit || totalArea > planter.areaLimit {
		return math.Inf(-1)
	}
	// per capita
	totalScore := 0.0
	for i := 0; i < N; i++ {
		if chromosome[i] {
			totalScore += planter.score[planter.sampleSet[i]]
		}
	}
	totalScore /= float64(planter.population)
	return totalScore * 100.0

}

func zip(s1, s2 []interface{}) []interface{} {
	s := make([]interface{}, 0)
	for i := range s1 {
		s = append(s, [2]interface{}{s1[i], s2[i]})
	}
	return s
}

// implementing interface for sorting
type twoArr struct {
	arr1 [][]bool
	arr2 []float64
}

func (ta twoArr) Len() int {
	return len(ta.arr1)
}

func (ta twoArr) Less(i, j int) bool {
	return ta.arr2[i] < ta.arr2[j]
}

func (ta twoArr) Swap(i, j int) {
	ta.arr1[i], ta.arr1[j] = ta.arr1[j], ta.arr1[i]
	ta.arr2[i], ta.arr2[j] = ta.arr2[j], ta.arr2[i]
}

func (planter *TreePlanter) crossover() {
	M := len(planter.sampleSet)
	X := planter.numOfChromosomes
	// for sorting one array according to other
	ta := twoArr{planter.chromosomes, planter.totalFit}
	sort.Sort(ta)

	newChrom := make([][]bool, X)

	for i := 0; i < X/2; i++ {
		piv := randInt(1, M-2)
		// array modified in place
		newChrom[i] = append(planter.chromosomes[i][:piv], planter.chromosomes[X/2-i-1][piv:]...)
		newChrom[X-i-1] = append(planter.chromosomes[X/2-i-1][:piv], planter.chromosomes[i][piv:]...)
		newChrom[X/2-i-1] = planter.chromosomes[i]
		newChrom[X/2+i] = planter.chromosomes[X/2-i-1]
	}

	planter.chromosomes = newChrom
}

func (planter *TreePlanter) runSearch(runtime float64, maxRep int, verbose bool) {
	tEnd := float64(time.Now().Unix()) + runtime
	t := 0
	for float64(time.Now().Unix()) <= tEnd {
		for i := 0; i < planter.numOfChromosomes; i++ {
			planter.totalFit[i] = planter.getFitness(i)
		}

		currBestFit := math.Inf(-1)
		var currBestCh []bool = nil
		for i, v := range planter.totalFit {
			if v > currBestFit {
				currBestFit = v
				currBestCh = planter.chromosomes[i]
			}
		}

		if currBestFit > planter.bestFit {
			planter.bestFit = currBestFit
			planter.bestChromosome = currBestCh
		}

		if t%1000 == 0 && verbose {
			fmt.Printf("Current Best Score at %d: %.2f\n", t, planter.bestFit)
		}

		relErr := math.Abs(currBestFit-planter.lastFit) / currBestFit
		if relErr <= 0.5 || math.IsNaN(relErr) {
			planter.repCount++
		}

		planter.lastFit = currBestFit
		t++
		if planter.repCount >= maxRep {
			planter.repCount = 0
			planter.initChromosomes()
			continue
		}
		planter.crossover()
	}
}

func (planter *TreePlanter) getResults() {
	trees := make(map[string]int, treeTypesCount)
	for i, s := range planter.sampleSet {
		if planter.bestChromosome[i] {
			trees[TreeInfo[s].CommonName]++
		}
	}

	totalScore, usedArea, usedCost := 0.0, 0, 0
	for k, v := range trees {
		totalScore += planter.score[TreeIdx[k]] * float64(v)
		usedArea += planter.area[TreeIdx[k]] * v
		usedCost += planter.cost[TreeIdx[k]] * v
	}

	totalScore /= float64(planter.population)

	treeMap, _ := json.MarshalIndent(trees, "", "\t")
	fmt.Printf("Trees : %s\nScore : %f\nArea : %d\nCost : %d\n",
		treeMap, totalScore, usedArea, usedCost)
}

func readTreeIdx(fname string) {
	// read tree indices
	data, _ := ioutil.ReadFile(fname)
	json.Unmarshal(data, &TreeIdx)
}

func readTreeInfo(fname string) {
	// read tree info
	data, _ := ioutil.ReadFile(fname)
	json.Unmarshal(data, &TreeInfo)

	treeTypesCount = len(TreeInfo)
}

func readCommandLineArgs(AQI, areaLimit, costLimit, population *int, runTime *float64) {
	// AQI, areaLimit, costLimit, population, runTime
	*AQI, _ = strconv.Atoi(os.Args[1])
	*areaLimit, _ = strconv.Atoi(os.Args[2])
	*costLimit, _ = strconv.Atoi(os.Args[3])
	*population, _ = strconv.Atoi(os.Args[4])
	*runTime, _ = strconv.ParseFloat(os.Args[5], 64)
}

func main() {
	// set the seed by clock
	rand.Seed(time.Now().Unix())

	readTreeIdx("/home/beast/tree_idx.json")
	readTreeInfo("/home/beast/tree_info.json")

	// read arguments from CLI
	var AQI, areaLimit, costLimit, population int
	var runTime float64
	readCommandLineArgs(&AQI, &areaLimit, &costLimit, &population, &runTime)

	// initialize tree planter
	planter := TreePlanter{}
	planter.init(AQI, areaLimit, costLimit, population)
	// run GA search
	planter.runSearch(runTime, 30, false)
	// display computed results
	planter.getResults()
}
