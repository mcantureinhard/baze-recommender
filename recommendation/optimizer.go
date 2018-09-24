package recommendation

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

type NutrientVectorCompare struct {
	comp int
	dist float64
}

type NutrientVector struct {
	values []float64
}

type PillNutrientVectors struct {
	m map[string]NutrientVector
}

type SimplePillList struct {
	names []string
}

type PillListNutrientVector struct {
	nutrientVector *NutrientVector
	pillList       *SimplePillList
}

type PillListNutrientVectors []PillListNutrientVector

type NutrientVectorMap struct {
	m map[string]int
}

type PillListMicronutrientsAccumulator struct {
	list *SimplePillList
	m    map[string]Micronutrient
}

type PillListMicronutrientsAccumulators []PillListMicronutrientsAccumulator

type PillNameMap struct {
	m map[string]*Pill
}

type Memory struct {
	sync.RWMutex
	m map[string]int
}

/*
	Returns:
	        comp:
			 -1 if any element is less than the target
			 0 if all elements are within target
			 1 if any element is greater than the target

			 dist:
			 euclidian distance if comp is -1 or 0
*/
func compareNutrientVectors(respond chan<- *NutrientVectorCompare, target *NutrientVector, check *NutrientVector, rangeLow float64, rangeHigh float64) {
	response := NutrientVectorCompare{comp: 0, dist: 0}
	var sqsum float64
	sqsum = 0
	var res int
	res = 0
	for i := range target.values {
		low := target.values[i] * (1 - rangeLow)
		high := target.values[i] * (1 + rangeHigh)
		if check.values[i] > high {
			response.comp = 1
			res = 1
			break
		}
		if check.values[i] < low {
			res = -1
		}
		sqsum += math.Pow(target.values[i]-check.values[i], 2)
	}
	response.comp = res
	response.dist = math.Sqrt(sqsum)
	respond <- &response
}

func optimizeLeastNumberOfPills(respond chan<- *PillListNutrientVectors, wg *sync.WaitGroup, deficitsVector *NutrientVector, pills *PillNutrientVectors, pillListNutrientVector PillListNutrientVector, mem *Memory) {
	var results PillListNutrientVectors
	var reiterate PillListNutrientVectors
	for k, v := range pills.m {
		tempNutrientVector := NutrientVector{values: make([]float64, len(pillListNutrientVector.nutrientVector.values))}
		for i, v := range pillListNutrientVector.nutrientVector.values {
			tempNutrientVector.values[i] = v
		}
		tempPillList := SimplePillList{}
		for _, v := range pillListNutrientVector.pillList.names {
			tempPillList.names = append(tempPillList.names, v)
		}

		tempPillListNutrientVector := PillListNutrientVector{pillList: &tempPillList, nutrientVector: &tempNutrientVector}
		tempPillListNutrientVector.pillList.names = append(tempPillListNutrientVector.pillList.names, k)
		if len(tempPillListNutrientVector.pillList.names) > 0 {
			sort.Sort(sort.StringSlice(tempPillListNutrientVector.pillList.names))
		} else {
			fmt.Println("uhhmmm got an empty list")
		}
		mem.Lock()
		shouldContinue := false
		_, ok := mem.m[strings.Join(tempPillListNutrientVector.pillList.names, "")]
		if !ok {
			mem.m[strings.Join(tempPillListNutrientVector.pillList.names, "")] = 1
		} else {
			shouldContinue = true
		}
		mem.Unlock()
		if shouldContinue {
			continue
		}
		for i, value := range v.values {
			tempPillListNutrientVector.nutrientVector.values[i] = tempPillListNutrientVector.nutrientVector.values[i] + value
		}
		compChan := make(chan *NutrientVectorCompare)
		go compareNutrientVectors(compChan, deficitsVector, tempPillListNutrientVector.nutrientVector, .3, .2)
		compResult := <-compChan
		close(compChan)
		if compResult.comp == 1 {
			//Exceeded range
			continue
		} else if compResult.comp == -1 {
			reiterate = append(reiterate, tempPillListNutrientVector)
		} else {
			results = append(results, tempPillListNutrientVector)
		}
	}
	if len(results) > 0 {
	} else if len(reiterate) > 0 {
		//if len(reiterate) > 0 {
		reiterateMaps := make(chan *PillListNutrientVectors, 4000)
		innerWG := new(sync.WaitGroup)
		for i := range reiterate {
			innerWG.Add(1)
			go optimizeLeastNumberOfPills(reiterateMaps, innerWG, deficitsVector, pills, reiterate[i], mem)
		}
		innerWG.Wait()
		close(reiterateMaps)
		//minPills := int((^uint(0)) >> 1)
		minPills := 9999999
		for c := range reiterateMaps {
			current := c
			if len(*current) > 0 {
				for _, el := range *current {
					if len(el.pillList.names) < minPills {
						var tempResult PillListNutrientVectors
						tempResult = append(tempResult, el)
						results = tempResult
						minPills = len(el.pillList.names)
					} else if len(el.pillList.names) == minPills {
						results = append(results, el)
					}
				}
			}
		}
	}
	respond <- &results
	wg.Done()
}

func optimize(respond chan<- *Pills, pillInventories *PillInventories, deficits *Deficits) {
	mem := Memory{m: make(map[string]int)}
	pillNameMap := PillNameMap{m: make(map[string]*Pill)}
	nutrientsVectorMap := NutrientVectorMap{m: make(map[string]int)}
	var deficitsVector NutrientVector
	//Build the vector that we want to attain
	for i, deficit := range *deficits {
		deficitsVector.values = append(deficitsVector.values, deficit.Quantity)
		nutrientsVectorMap.m[deficit.Name] = i
	}

	pillVectors := PillNutrientVectors{m: make(map[string]NutrientVector)}
	for _, pillInventory := range *pillInventories {
		pillName := pillInventory.PillData.Name
		pillNameMap.m[pillName] = pillInventory.PillData
		pillNutrientsVector := NutrientVector{values: make([]float64, len(deficitsVector.values))}
		for _, pillMicronutrient := range pillInventory.PillData.PillMicronutrients {
			value := float64(pillMicronutrient.Absortion) * float64(pillMicronutrient.MicroNutrient.Quantity) / float64(100)
			pillNutrientsVector.values[nutrientsVectorMap.m[pillMicronutrient.MicroNutrient.Name]] = value
		}
		pillVectors.m[pillInventory.PillData.Name] = pillNutrientsVector
	}
	optimizeChannel := make(chan *PillListNutrientVectors, 4000)
	baseNutrientVector := NutrientVector{values: make([]float64, len(deficitsVector.values))}
	seedList := SimplePillList{}
	basePillListNutrientVector := PillListNutrientVector{nutrientVector: &baseNutrientVector, pillList: &seedList}
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go optimizeLeastNumberOfPills(optimizeChannel, wg, &deficitsVector, &pillVectors, basePillListNutrientVector, &mem)
	wg.Wait()
	close(optimizeChannel)
	var results *PillListNutrientVectors
	for c := range optimizeChannel {
		results = c
	}
	var pills Pills
	if len(*results) > 0 {
		//I should write a compare fn that returns also the vector...
		minDist := 99999999.99 //I should set max float64
		var bestMatch PillListNutrientVector
		for _, result := range *results {
			compChannel := make(chan *NutrientVectorCompare)
			go compareNutrientVectors(compChannel, &deficitsVector, result.nutrientVector, .3, .2)
			comp := <-compChannel
			if comp.dist < minDist {
				bestMatch = result
				minDist = comp.dist
			}
		}

		for _, name := range bestMatch.pillList.names {
			pills = append(pills, *pillNameMap.m[name])
		}
	}
	respond <- &pills
}
