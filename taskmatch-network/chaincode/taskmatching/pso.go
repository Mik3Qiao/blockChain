package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

// Vector : This is the position inside a velocity or a position struct
// type Vector []float64

// Position : this contains the currrent position and the point's fitness value
type Position struct {
	position []float64
	cost     float64
}

// Problem : defines the structure of a problem, including the number of tasks and
// resources
type Problem struct {
	nVar   int // number of tasks
	varmin int
	varMax int
}

// Particle : this is the particle struct
type Particle struct {
	position []float64
	velocity []float64
	pBest    []float64
	cost     float64
	bestCost float64
}

type indexValuePair struct {
	row   int
	col   int
	value float64
}

func multiplyNumAndArr(factor float64, arrIn []float64) []float64 {
	result := make([]float64, len(arrIn))
	for i := 0; i < len(arrIn); i++ {
		result[i] = arrIn[i] * factor
	}
	return result
}

func trimPosition(inputVec []float64, lower int, upper int) {
	for i := 0; i < len(inputVec); i++ {
		inputVec[i] = math.Max(inputVec[i], float64(lower))
		inputVec[i] = math.Min(inputVec[i], float64(upper-1))
	}
}

func addArrs(arrs ...[]float64) []float64 {
	result := make([]float64, len(arrs[0]))
	for _, arr := range arrs {
		for i := 0; i < len(result); i++ {
			result[i] += arr[i]
		}
	}
	return result
}

func multiplyArrs(arr1 []float64, arr2 []float64) []float64 {
	result := make([]float64, len(arr1))
	for i := 0; i < len(arr1); i++ {
		result[i] = arr1[i] * arr2[i]
	}
	return result
}

func subtractArrs(arr1 []float64, arr2 []float64) []float64 {
	result := make([]float64, len(arr1))
	for i := 0; i < len(arr1); i++ {
		result[i] = arr1[i] - arr2[i]
	}
	return result
}

func generateRandomArr(lower float64, upper float64, size int) []float64 {
	result := make([]float64, size)
	for i := 0; i < size; i++ {
		rand.Seed(time.Now().UnixNano())
		result[i] = rand.Float64()*(upper-lower) + lower
	}
	return result
}

func fetchRunTime(inputMatrix [][]float64, task int, resource int) float64 {
	return inputMatrix[task][resource]
}

func evaluate(inputMatrix [][]float64, inputSol []int) float64 {
	makespan := make([]float64, len(inputMatrix[0]))
	resources := len(inputMatrix[0])
	for i := 0; i < len(inputSol); i++ { // length of input solution = # of tasks
		temp := inputSol[i] // temp => corresponding resource assigned to each task
		if temp > resources || temp < 0 {
			temp = temp % (resources - 1)
		}
		result := fetchRunTime(inputMatrix, i, temp)
		makespan[temp] = makespan[temp] + result
	}
	maxCompletion := float64(-1)
	for j := 0; j < len(makespan); j++ {
		if makespan[j] > maxCompletion {
			maxCompletion = makespan[j]
		}
	}
	return maxCompletion
}

// ETCgenerator : generate an ETC matrix based on # tasks, resources, heterogenety of task and resource
func ETCgenerator(task int, resource int, taskHetero string, resourceHetero string) [][]float64 {
	result := make([][]float64, task)
	for i := range result {
		result[i] = make([]float64, resource)
	}
	var taskBound float64
	var resourceBound float64

	if taskHetero == "hi" {
		taskBound = 3000
	} else {
		taskBound = 100
	}

	if resourceHetero == "hi" {
		resourceBound = 1000
	} else {
		resourceBound = 10
	}

	for i := range result {
		result[i][0] = rand.Float64()*(taskBound-1.0) + 1.0
	}

	start := 1

	for i := 0; i < task; i++ {
		start = 1
		for j := 0; j < resource; j++ {
			if j == (resource - 1) {
				start = resource - 1
			}
			result[i][start] = result[i][0] * (rand.Float64()*(resourceBound-1.0) + 1.0)
			start++
		}
	}

	for i := 0; i < task; i++ {
		result[i][0] = result[i][0] * (rand.Float64()*(resourceBound-1.0) + 1.0)
	}

	return result
}

func pso(inputProblem Problem, inputMatrix [][]float64, maxIter int, popSize int, c1 float64, c2 float64, w float64, wdamp float64) (Position, []Particle) {
	// Initialize an empty object of type "Particle"
	var emptyParticle Particle

	// Extract problem information
	varMin := inputProblem.varmin
	varMax := inputProblem.varMax
	nVar := inputProblem.nVar

	gBest := Position{nil, math.Inf(1)}

	pop := []Particle{}

	// This loop is for initialization
	for i := 0; i < popSize; i++ {
		pop = append(pop, emptyParticle)
		pop[i].position = generateRandomArr(float64(varMin), float64(varMax), nVar)
		pop[i].velocity = generateRandomArr(float64(-varMax), float64(varMax), nVar)
		x := make([]int, len(pop[i].position))
		for j := 0; j < len(x); j++ {
			x[j] = int(pop[i].position[j])
		}
		pop[i].cost = evaluate(inputMatrix, x)
		// copy(pop[i].pBest, pop[i].position)
		pop[i].pBest = pop[i].position
		pop[i].bestCost = pop[i].cost

		if pop[i].bestCost < gBest.cost {
			// copy(gBest.position, pop[i].pBest)
			gBest.position = pop[i].pBest
			gBest.cost = pop[i].bestCost
		}
		// fmt.Println(pop[i].velocity)
	}
	//PSO loop
	for iter := 0; iter < maxIter; iter++ {
		for i := 0; i < popSize; i++ {
			pop[i].velocity = addArrs(multiplyNumAndArr(w, pop[i].velocity),
				multiplyArrs(multiplyNumAndArr(c1, generateRandomArr(0, 1, nVar)), subtractArrs(pop[i].pBest, pop[i].position)),
				multiplyArrs(multiplyNumAndArr(c2, generateRandomArr(0, 1, nVar)), subtractArrs(gBest.position, pop[i].position)))

			pop[i].position = addArrs(pop[i].position, pop[i].velocity)
			trimPosition(pop[i].position, varMin, varMax)

			x := make([]int, len(pop[i].position))
			for j := 0; j < len(x); j++ {
				x[j] = int(pop[i].position[j])
			}

			pop[i].cost = evaluate(inputMatrix, x)
			if pop[i].cost < pop[i].bestCost {
				// copy(pop[i].pBest, pop[i].position)
				pop[i].pBest = pop[i].position
				pop[i].bestCost = pop[i].cost
				if pop[i].bestCost < gBest.cost {
					// copy(gBest.position, pop[i].pBest)
					gBest.position = pop[i].pBest
					gBest.cost = pop[i].bestCost
				}
			}
			// fmt.Printf("%s%f\n", "The current position is:", pop[i].position)
		}
		w *= wdamp
		// fmt.Printf("%s%d%s%f%s%f\n", "Iteration: ", iter, " Best Cost: ", gBest.cost, " ,the position chosen is:", gBest.position)

	}
	return gBest, pop
}

func getminIndicesByRow(inputMatrix [][]float64) []indexValuePair {
	result := make([]indexValuePair, len(inputMatrix))
	for i := 0; i < len(inputMatrix); i++ {
		var minofRow float64 = math.MaxInt16
		for j := 0; j < len(inputMatrix[i]); j++ {
			if inputMatrix[i][j] < minofRow {
				minofRow = inputMatrix[i][j]
				result[i].row = i
				result[i].col = j
				result[i].value = minofRow
			}
		}
	}
	return result
}

func getmaxIndicesByRow(inputMatrix [][]float64) []indexValuePair {
	result := make([]indexValuePair, len(inputMatrix))
	for i := 0; i < len(inputMatrix); i++ {
		var maxofRow float64 = -1
		for j := 0; j < len(inputMatrix[i]); j++ {
			if inputMatrix[i][j] > maxofRow {
				maxofRow = inputMatrix[i][j]
				result[i].row = i
				result[i].col = j
				result[i].value = maxofRow
			}
		}
	}
	return result
}
func getmaxIndicesByCol(inputMatrix [][]float64) []indexValuePair {
	result := make([]indexValuePair, len(inputMatrix[0]))
	for j := 0; j < len(inputMatrix[0]); j++ {
		var maxofCol float64 = -1
		for i := 0; i < len(inputMatrix); i++ {
			if inputMatrix[i][j] > maxofCol {
				maxofCol = inputMatrix[i][j]
				result[j].row = i
				result[j].col = j
				result[j].value = maxofCol
			}
		}
	}
	return result
}

func getminIndicesByCol(inputMatrix [][]float64) []indexValuePair {

	result := make([]indexValuePair, len(inputMatrix[0]))
	for j := 0; j < len(inputMatrix[0]); j++ {
		var minofCol float64 = math.MaxInt16
		for i := 0; i < len(inputMatrix); i++ {
			if inputMatrix[i][j] < minofCol {
				minofCol = inputMatrix[i][j]
				result[j].row = i
				result[j].col = j
				result[j].value = minofCol
			}
		}
	}
	return result
}

func getMinIndexValuePair(inputMatrix []indexValuePair) indexValuePair {
	var min float64 = math.MaxInt16
	var minPair indexValuePair
	for i := 0; i < len(inputMatrix); i++ {
		if inputMatrix[i].value < min {
			min = inputMatrix[i].value
			minPair = inputMatrix[i]
		}
	}
	return minPair
}

func getMaxIndexValuePair(inputMatrix []indexValuePair) indexValuePair {
	var max float64 = -1
	var maxPair indexValuePair
	for i := 0; i < len(inputMatrix); i++ {
		if inputMatrix[i].value > max {
			max = inputMatrix[i].value
			maxPair = inputMatrix[i]
		}
	}
	return maxPair
}

func shrinkMatrixRow(inputMatrix [][]float64, rowRemoved int) [][]float64 {
	result := make([][]float64, len(inputMatrix)-1)
	for c := range result {
		result[c] = make([]float64, len(inputMatrix[c]))
	}
	if len(inputMatrix) == 1 {
		return inputMatrix
	}

	newRow := 0
	for OriRow := 0; OriRow < len(inputMatrix); OriRow++ {
		if OriRow != rowRemoved {
			result[newRow] = inputMatrix[OriRow]
			newRow++
		}
	}
	return result
}

func helper(inputMatrix [][]float64, result []indexValuePair, timespent []float64, input string) ([]indexValuePair, []float64) {
	if len(inputMatrix) == 1 {
		var incides []indexValuePair
		inputArr := strings.Split(input, "-")

		if inputArr[0] == "MIN" && inputArr[2] == "TASK" {
			incides = getminIndicesByRow(inputMatrix)
		} else if inputArr[0] == "MAX" && inputArr[2] == "TASK" {
			incides = getmaxIndicesByRow(inputMatrix)
		} else if inputArr[0] == "MIN" && inputArr[2] == "RESOURCE" {
			incides = getminIndicesByCol(inputMatrix)
		} else if inputArr[0] == "MAX" && inputArr[2] == "RESOURCE" {
			incides = getmaxIndicesByCol(inputMatrix)
		}

		var valuePair indexValuePair
		if inputArr[1] == "MIN" {
			valuePair = getMinIndexValuePair(incides)
		} else {
			valuePair = getMaxIndexValuePair(incides)
		}

		result = append(result, valuePair)
		timespent = append(timespent, valuePair.value)
		return result, timespent
	}

	var incides []indexValuePair
	inputArr := strings.Split(input, "-")

	if inputArr[0] == "MIN" && inputArr[2] == "TASK" {
		incides = getminIndicesByRow(inputMatrix)
	} else if inputArr[0] == "MAX" && inputArr[2] == "TASK" {
		incides = getmaxIndicesByRow(inputMatrix)
	} else if inputArr[0] == "MIN" && inputArr[2] == "RESOURCE" {
		incides = getminIndicesByCol(inputMatrix)
	} else if inputArr[0] == "MAX" && inputArr[2] == "RESOURCE" {
		incides = getmaxIndicesByCol(inputMatrix)
	}

	var valuePair indexValuePair
	if inputArr[1] == "MIN" {
		valuePair = getMinIndexValuePair(incides)
	} else {
		valuePair = getMaxIndexValuePair(incides)
	}

	tempMatrix := shrinkMatrixRow(inputMatrix, valuePair.row)
	for i := 0; i < len(tempMatrix); i++ {
		tempMatrix[i][valuePair.col] += valuePair.value
	}
	result = append(result, valuePair)
	timespent = append(timespent, valuePair.value)
	return helper(tempMatrix, result, timespent, input)
}

func deepcopy(inputMatrix [][]float64) [][]float64 {
	result := make([][]float64, len(inputMatrix))
	for i := range result {
		result[i] = make([]float64, len(inputMatrix[i]))
	}
	for i := 0; i < len(inputMatrix); i++ {
		for j := 0; j < len(inputMatrix[i]); j++ {
			result[i][j] = inputMatrix[i][j]
		}
	}
	return result
}

func main() {
	var newproblem = Problem{100, 0, 10} // 512 tasks and 16 resources
	// ETC := [][]float64{
	// 	{1.2, 1.3, 1.4},
	// 	{7.2, 6.9, 10.2},
	// 	{16.2, 3.9, 4.7},
	// 	{1.2, 5.8, 12.0}}

	// for i := 0; i < len(ETC); i++ {
	// 	fmt.Printf("%.2f\n", ETC[i])
	// }
	// tasks := len(ETC)
	// resources := len(ETC[0])
	ETC := ETCgenerator(100, 10, "low", "low")
	ETC1 := deepcopy(ETC)
	// ETC2 := deepcopy(ETC)
	// for i := 0; i < len(ETC); i++ {
	// 	fmt.Printf("%.2f\n", ETC[i])
	// }

	// startminmax := time.Now()
	// var emptyArr []indexValuePair
	// var emptyArr1 []float64
	// _, timespent := helper(ETC2, emptyArr, emptyArr1, "MIN-MIN-TASK")
	// timeCost := float64(-1)
	// for i := 0; i < len(timespent); i++ {
	// 	if timespent[i] > timeCost {
	// 		timeCost = timespent[i]
	// 	}
	// }
	// elapsedminmax := time.Since(startminmax)
	// fmt.Printf("%s%.2f\n", "The time cost by minmin task driven approach is: ", timeCost)
	// fmt.Printf("time took by by minmin: %s\n", elapsedminmax)

	// ------------------------------------------------------------------------------------------------------------
	// ------------------------------------------------------------------------------------------------------------
	// ------------------------------------------------------------------------------------------------------------
	// ------------------------------------------------------------------------------------------------------------
	// ------------------------------------------------------------------------------------------------------------

	startpso := time.Now()
	gbest, _ := pso(newproblem, ETC1, 500, 50, 1.796180, 1.796180, 0.729844, 0.995)
	fmt.Printf("%s%.2f\n", "Cost token by pso is: ", gbest.cost)
	elapsedpso := time.Since(startpso)
	fmt.Printf("time took by pso: %s\n\n\n", elapsedpso)
	// sol := generateRandomArr(0, 3, 4)
	// fmt.Print(sol)

}
