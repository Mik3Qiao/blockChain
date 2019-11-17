package main

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type indexValuePair struct {
	index int
	value int
}

type TaskMatching struct {
	identifier string `json:"id"`       //docType is used to distinguish the various types of objects in state database
	Runtimes   string `json:"runtimes"` //the fieldtags are needed to keep case from bouncing around
}

type Peer struct {
	identifier string `json:"id"`
	Status     string `json:"status"`
	Solution   []int  `json:"sol"`
	Runtime    int    `json:"runtime"`
	Name       string `json:"name"`
}

type TaskMatchingSol struct {
	identifier string `json:"id"`
	Runtime    int    `json:"runtime"`
	Solution   []int  `json:"sol"`
	Owner      string `json:"owner"`
	Algorithm  string `json:"alg"`
	Runtimes   string `json:"runtimes"`
}

type Count struct {
	identifier string `json:"id"`
	Counter    int    `json:"count"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "createTaskMatching" { //create a new taskmatching
		return t.createTaskMatching(stub, args)
	} else if function == "readTaskMatching" { //reads a taskmatching
		return t.readTaskMatching(stub, args)
	} else if function == "Initialize" { //initialize the network
		return t.Initialize(stub)
	} else if function == "calculateTaskMatching" { //calculate a taskmatching
		t.calculateTaskMatching(stub, args)

		if t.allPeersDone(stub) {
			return t.setBestSol(stub)
		} else {
			return shim.Success(nil)
		}
	}
	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

func (t *SimpleChaincode) Initialize(stub shim.ChaincodeStubInterface) pb.Response {
	var err error

	p1 := &Peer{"p1", "waiting", make([]int, 0), -1, "Peer 1"}
	p1JSONasBytes, _ := json.Marshal(p1)

	err = stub.PutState("p1", p1JSONasBytes) //write the peer
	if err != nil {
		return shim.Error(err.Error())
	}

	p2 := &Peer{"p2", "waiting", make([]int, 0), -1, "Peer 2"}
	p2JSONasBytes, _ := json.Marshal(p2)

	err = stub.PutState("p2", p2JSONasBytes) //write the peer
	if err != nil {
		return shim.Error(err.Error())
	}

	p3 := &Peer{"p3", "waiting", make([]int, 0), -1, "Peer 3"}
	p3JSONasBytes, _ := json.Marshal(p3)

	err = stub.PutState("p3", p3JSONasBytes) //write the peer
	if err != nil {
		return shim.Error(err.Error())
	}

	count := &Count{"count", 0}
	countAsBytes, _ := json.Marshal(count)

	stub.PutState("count", countAsBytes)

	return shim.Success(nil)
}

func (t *SimpleChaincode) calculateTaskMatching(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//Get the Task Math matrix
	TaskMatchAsBytes, _ := stub.GetState("work")
	tmpTM := TaskMatching{}

	json.Unmarshal(TaskMatchAsBytes, &tmpTM)

	//Convert matrix string to float matrix
	var matrix [][]int = strToMatrix(tmpTM.Runtimes)

	//pass matrix to solution calculator
	var sol []int
	var runtime int

	sol, runtime = Assign(matrix, args[0])

	// var runtime float64 = calcRuntime(matrix, sol)
	if args[0] == "p3" {
		runtime = calcRuntime(matrix, sol)
	}

	//change Peer info
	PeerasBytes, _ := stub.GetState(args[0])
	tmpPeer := Peer{}

	json.Unmarshal(PeerasBytes, &tmpPeer)

	tmpPeer.Status = "done"
	tmpPeer.Solution = sol
	tmpPeer.Runtime = runtime

	PeerAsJSONbytes, _ := json.Marshal(tmpPeer)

	stub.PutState(args[0], PeerAsJSONbytes)

	return shim.Success(nil)
	//
}

func strToMatrix(input string) [][]int {
	var parsed [][]int
	json.Unmarshal([]byte(input), &parsed)
	return parsed
}

func Assign(matrix [][]int, peer string) ([]int, int) {
	var sol []int
	rand.Seed(time.Now().UnixNano())
	timeCost := -1
	if peer == "p1" {
		sol, timeCost = minmin(matrix)
	} else if peer == "p2" {
		solIndexValuePair, time := minmax(matrix)
		for i := 0; i < len(solIndexValuePair); i++ {
			sol = append(sol, solIndexValuePair[i].index)
			sol = append(sol, solIndexValuePair[i].value)
		}
		timeCost = time
	} else if peer == "p3" {
		sol = simulatedAnnealing(matrix)
		timeCost = -1
	}

	return sol, timeCost
}

func calcRuntime(mat [][]int, indices []int) int {
	var runtimes = make([]int, len(mat[0]))

	//add runtimes
	for i := 0; i < len(mat); i++ {
		runtimes[indices[i]] += mat[i][indices[i]]
	}

	//calculate max
	var max int
	max = -1

	for i := 0; i < len(runtimes); i++ {
		if runtimes[i] > max {
			max = runtimes[i]
		}
	}

	fmt.Println(runtimes)
	return max
}

func (t *SimpleChaincode) allPeersDone(stub shim.ChaincodeStubInterface) bool {
	peerArray := [3]string{"p1", "p2", "p3"}
	tmpPeer := Peer{}

	//loop over all of the peers
	for i := 0; i < len(peerArray); i++ {
		//check to see if any of the peers haven't finished

		//query chaincode to get the result
		PeerasBytes, _ := stub.GetState(peerArray[i])
		json.Unmarshal(PeerasBytes, &tmpPeer)

		if tmpPeer.Status != "done" {
			return false
		}
	}

	return true
}

//Method to set the best solution
func (t *SimpleChaincode) setBestSol(stub shim.ChaincodeStubInterface) pb.Response {
	peerArray := [3]string{"p1", "p2", "p3"}
	tmpPeer := Peer{}
	solPeer := Peer{}
	// var min float64 = math.MaxFloat64
	var min int = math.MaxInt8

	//find which peer found the best solution and save their information
	for i := 0; i < len(peerArray); i++ {
		PeerasBytes, _ := stub.GetState(peerArray[i])
		json.Unmarshal(PeerasBytes, &tmpPeer)

		if tmpPeer.Runtime < min {
			min = tmpPeer.Runtime
			solPeer = tmpPeer
		}
	}

	//get the current matrix we were working on from the ledger
	taskMatchingAsBytes, _ := stub.GetState("work")
	tmpTM := TaskMatching{}

	json.Unmarshal(taskMatchingAsBytes, &tmpTM)

	//get the current count for how many solutions have been created.
	countAsBytes, _ := stub.GetState("count")
	tmpCount := Count{}

	json.Unmarshal(countAsBytes, &tmpCount)

	tmpCount.Counter += 1
	solNum := strconv.Itoa(tmpCount.Counter)

	var algName string

	//find which algorithm was used to calculate the solution.
	if solPeer.Name == "Peer 1" {
		algName = "min-min"
	} else if solPeer.Name == "Peer 2" {
		algName = "max-min"
	} else {
		algName = "Simulated Annealing"
	}

	TMSol := TaskMatchingSol{solNum, solPeer.Runtime, solPeer.Solution, solPeer.Name, algName, tmpTM.Runtimes}

	//update count and add TM sol
	countAsJSON, _ := json.Marshal(tmpCount)
	stub.PutState("count", countAsJSON)

	TMSolAsJSON, _ := json.Marshal(TMSol)
	stub.PutState(solNum, TMSolAsJSON)

	return shim.Success(nil)
}

// ============================================================
// createTaskMatching - create a taskmatching
// ============================================================
func (t *SimpleChaincode) createTaskMatching(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// 0       1
	//id   runtimes
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	fmt.Println("- creating TaskMatching")

	identifier := args[0]
	runtimes := strings.ToLower(args[1])

	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	// ==== Check if TaskMatching already exists ====
	TaskMatchingAsBytes, err := stub.GetState(identifier)
	if err != nil {
		return shim.Error("Failed to get TaskMatching: " + err.Error())
	} else if TaskMatchingAsBytes != nil {
		fmt.Println("This TaskMatching already exists: " + identifier)
		return shim.Error("This TaskMatching already exists: " + identifier)
	}

	// ==== Create TaskMatching object and marshal to JSON ====
	TaskMatching := &TaskMatching{identifier, runtimes}
	TaskMatchingJSONasBytes, err := json.Marshal(TaskMatching)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save taskmatching to state ===
	err = stub.PutState(identifier, TaskMatchingJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== taskmathing saved. Return success ====
	fmt.Println("- end init TaskMatching")
	return shim.Success(nil)
}

// ================================================================================================================
// readTaskMatching: This method can actually read anything that is saved onto the ledger not just taskmatchings
// ================================================================================================================
func (t *SimpleChaincode) readTaskMatching(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var identifier, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the TaskMatching to query")
	}

	identifier = args[0]
	TaskMatchingAsbytes, err := stub.GetState(identifier) //get the TaskMatching from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + identifier + "\"}"
		return shim.Error(jsonResp)
	} else if TaskMatchingAsbytes == nil {
		jsonResp = "{\"Error\":\"TaskMatching does not exist: " + identifier + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(TaskMatchingAsbytes)
}

// Newly added code
// -----------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------
func minmin(inputMatrix [][]int) ([]int, int) {
	var emptyArr []int
	tempMatrix := inputMatrix
	choices := minminhelper(tempMatrix, emptyArr)
	timeCost := -1
	for i := 0; i < len(tempMatrix); i++ {
		if tempMatrix[i][choices[i]] > timeCost {
			timeCost = tempMatrix[i][choices[i]]
		}
	}
	return choices, timeCost
}

func minminhelper(inputMatrix [][]int, result []int) []int {
	if len(inputMatrix) == 1 {
		minIncides := getminIndices(inputMatrix)
		result = append(result, minIncides[0])
		return result
	}
	minIncides := getminIndices(inputMatrix)
	result = append(result, minIncides[0])
	tempMatrix := shrinkMatrixRow(inputMatrix, 0)
	for i := 0; i < len(tempMatrix); i++ {
		tempMatrix[i][minIncides[0]] += inputMatrix[0][minIncides[0]]
	}
	return minminhelper(tempMatrix, result)
}

func getminIndices(inputMatrix [][]int) []int {
	result := make([]int, len(inputMatrix))
	for i := 0; i < len(inputMatrix); i++ {
		var minofRow int = math.MaxInt16
		for j := 0; j < len(inputMatrix[i]); j++ {
			if inputMatrix[i][j] < minofRow {
				minofRow = inputMatrix[i][j]
				result[i] = j
			}
		}
	}
	return result
}

func shrinkMatrixRow(inputMatrix [][]int, rowRemoved int) [][]int {
	result := make([][]int, len(inputMatrix)-1)
	for c := range result {
		result[c] = make([]int, len(inputMatrix[c]))
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

func minmax(inputMatrix [][]int) ([]indexValuePair, int) {
	var emptyArr []indexValuePair
	var emptyArr1 []int
	tempMatrix := inputMatrix
	choices, timespent := minmaxHelper(tempMatrix, emptyArr, emptyArr1) // choices contains the row and col number of the original matrix
	timeCost := -1
	fmt.Println(timespent)
	for i := 0; i < len(timespent); i++ {
		if timespent[i] > timeCost {
			timeCost = timespent[i]
		}
	}
	return choices, timeCost
}

func getMaxIndexValuePair(inputMatrix []indexValuePair) *indexValuePair {
	var max int = -1
	var maxPair *indexValuePair
	for i := 0; i < len(inputMatrix); i++ {
		if inputMatrix[i].value > max {
			max = inputMatrix[i].value
			maxPair = &inputMatrix[i]
		}
	}
	return maxPair
}

func minmaxHelper(inputMatrix [][]int, result []indexValuePair, timespent []int) ([]indexValuePair, []int) {
	if len(inputMatrix) == 1 {
		minIncides := getminIndices(inputMatrix)
		result = append(result, indexValuePair{0, minIncides[0]})
		timespent = append(timespent, inputMatrix[0][minIncides[0]])
		return result, timespent
	}
	minIncides := getminIndices(inputMatrix)
	var indeciesExtracted []indexValuePair
	var minValues []indexValuePair
	for i := 0; i < len(inputMatrix); i++ {
		minValues = append(minValues, indexValuePair{index: i, value: inputMatrix[i][minIncides[i]]})
		indeciesExtracted = append(indeciesExtracted, indexValuePair{index: i, value: minIncides[i]})

	}
	maxValuePair := getMaxIndexValuePair(minValues)         // row# and maxValue
	indexExtracted := indeciesExtracted[maxValuePair.index] // row# and index of maxValue
	// tempMatrix := copyMatrix(inputMatrix)
	// for i := 0; i < len(tempMatrix) && i != maxValuePair.index; i++ {
	// 	tempMatrix[i][indexExtracted.value] += maxValuePair.value
	// }
	tempMatrix := shrinkMatrixRow(inputMatrix, maxValuePair.index)
	for i := 0; i < len(tempMatrix); i++ {
		tempMatrix[i][indexExtracted.value] += maxValuePair.value
	}
	result = append(result, indexExtracted)
	row := indexExtracted.index
	col := indexExtracted.value
	timespent = append(timespent, inputMatrix[row][col])
	return minmaxHelper(tempMatrix, result, timespent)
}

// Newly added code
// -----------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------

/*
 * Min-Max calculates the minimum of every row and then takes the **MAXIMUM** of those minimums
 * and adds it to our solution. This process is repeated until a solution is created.
 */
func minmax_rec(label_matrix [][][]int, min_indices []int, sol []int) []int {
	if len(label_matrix) == 1 {
		matrix := getMatrix(label_matrix)
		var row_ind int = maxOfMins(matrix, min_indices)
		var col_ind int = min_indices[row_ind]

		var orig_row int = int(label_matrix[row_ind][col_ind][1])
		var orig_col int = int(label_matrix[row_ind][col_ind][2])

		sol[orig_row] = orig_col

		return sol
	} else {

		//fmt.Println(len(label_matrix))
		matrix := getMatrix(label_matrix)
		var row_ind int = maxOfMins(matrix, min_indices)
		var col_ind int = min_indices[row_ind]

		var orig_row int = int(label_matrix[row_ind][col_ind][1])
		var orig_col int = int(label_matrix[row_ind][col_ind][2])

		sol[orig_row] = orig_col

		label_matrix = decreaseSize(label_matrix, row_ind, col_ind)
		matrix = getMatrix(label_matrix)

		min_indices = fix_min_indices(matrix, min_indices, row_ind, col_ind)

		return minmax_rec(label_matrix, min_indices, sol)
	}
}

/*
 * Decreases the size of a matrix by removing a row & column
 */
func decreaseSize(matrix [][][]int, row_ind int, col_ind int) [][][]int {
	var new_mat [][][]int

	new_mat = make([][][]int, len(matrix)-1)

	for i := range new_mat {
		new_mat[i] = make([][]int, len(matrix[0])-1)
		for j := range new_mat[i] {
			new_mat[i][j] = make([]int, 3)
		}
	}

	var row_num int = 0
	var col_num int = 0

	for i := 0; i < len(matrix); i++ {
		if i != row_ind {
			col_num = 0

			for j := 0; j < len(matrix[0]); j++ {
				if j != col_ind {
					new_mat[row_num][col_num][0] = matrix[i][j][0]
					new_mat[row_num][col_num][1] = matrix[i][j][1]
					new_mat[row_num][col_num][2] = matrix[i][j][2]
					col_num = col_num + 1
				}
			}

			row_num = row_num + 1
		}
	}

	return new_mat
}

/*
 * Min indices can get broken after the matrix size is decreased so this method fixes them
 */
func fix_min_indices(matrix [][]int, min_indices []int, rem_row int, rem_col int) []int {
	var new_min_ind []int

	new_min_ind = make([]int, len(min_indices)-1)
	var count int = 0

	for i := 0; i < len(min_indices); i++ {
		if i != rem_row {
			if min_indices[i] == rem_col {
				new_min_ind[count] = min_index(matrix[count])
			} else if min_indices[i] > rem_col {
				new_min_ind[count] = min_indices[i] - 1
			} else {
				new_min_ind[count] = min_indices[i]
			}
			count = count + 1
		}
	}

	return new_min_ind
}

/*
 * Initialize the minimum indices for the matrix
 */
func init_mins(matrix [][]int) []int {
	var min int = math.MaxInt8
	var ind int = -1
	var sol []int

	sol = make([]int, len(matrix))

	for i := 0; i < len(matrix); i++ {
		min = math.MaxInt8
		ind = -1

		for j := 0; j < len(matrix[0]); j++ {
			if matrix[i][j] < min {
				min = matrix[i][j]
				ind = j
			}
		}

		sol[i] = ind
	}

	return sol
}

/*
 * Method to intialize a matrix with an additional dimension for the x and y values.
 */
func initMatrix(matrix [][]int) [][][]int {
	var new_matrix [][][]int

	new_matrix = make([][][]int, len(matrix))

	for i := range new_matrix {
		new_matrix[i] = make([][]int, len(matrix[0]))
		for j := range new_matrix[i] {
			new_matrix[i][j] = make([]int, 3)
		}
	}

	for i := 0; i < len(matrix); i++ {
		for j := 0; j < len(matrix[0]); j++ {
			new_matrix[i][j][0] = matrix[i][j]
			// new_matrix[i][j][1] = float64(i)
			new_matrix[i][j][1] = (i)
			// new_matrix[i][j][2] = float64(j)
			new_matrix[i][j][2] = (j)
		}
	}

	return new_matrix
}

/*
 * Finds the minimum of the minimums
 */
// func minOfMins(matrix [][]int, min_ind []int) int {
// 	// var min float64 = math.MaxFloat64
// 	var min int = math.MaxInt8
// 	var ind int = -1

// 	for i := 0; i < len(matrix); i++ {
// 		if matrix[i][min_ind[i]] < min {
// 			min = matrix[i][min_ind[i]]
// 			ind = i
// 		}
// 	}

// 	return ind
// }

/*
 * Finds the maximum of the minimums
 */
func maxOfMins(matrix [][]int, min_ind []int) int {
	var max int = -1
	var ind int = -1

	for i := 0; i < len(matrix); i++ {
		if matrix[i][min_ind[i]] > max {
			max = matrix[i][min_ind[i]]
			ind = i
		}
	}

	return ind
}

/*
 * Gets just the matrix values from the augmented matrix that also contains column and row data for each value.
 */
func getMatrix(matrix [][][]int) [][]int {

	var new_matrix [][]int

	new_matrix = make([][]int, len(matrix))

	for i := range new_matrix {
		new_matrix[i] = make([]int, len(matrix[0]))
	}

	for i := 0; i < len(matrix); i++ {
		for j := 0; j < len(matrix[0]); j++ {
			new_matrix[i][j] = matrix[i][j][0]
		}
	}

	return new_matrix
}

/*
 * Find the minimum index of a row
 */
func min_index(row []int) int {
	var min int = math.MaxInt8
	var ind int = -1

	for i := 0; i < len(row); i++ {
		if row[i] < min {
			min = row[i]
			ind = i
		}
	}

	return ind
}

func iToFMatrix(inputMatrix [][]int) [][]float64 {
	newMatrix := make([][]float64, len(inputMatrix))
	for i := range newMatrix {
		newMatrix[i] = make([]float64, len(inputMatrix[i]))
	}
	for j := 0; j < len(inputMatrix); j++ {
		for k := 0; k < len(inputMatrix[j]); k++ {
			newMatrix[j][k] = float64(inputMatrix[j][k])
		}
	}
	return newMatrix
}

/**************************************************
 **          Simulated Annealing Code           **
**************************************************/

func simulatedAnnealing(matrix [][]int) []int {
	var temp float64 = 10000
	var coolingRate float64 = 0.003
	var currentEnergy float64
	var newEnergy float64

	var best_sol []int = init_sol(len(matrix))
	var bestEnergy = float64(calcRuntime(matrix, best_sol))

	var curr_sol []int = copyIntArr(best_sol)
	currentEnergy = bestEnergy

	var new_sol []int

	for i := 0; temp > 1; i++ {
		new_sol = copyIntArr(curr_sol)
		new_sol = SA_swap(new_sol)
		newEnergy = float64(calcRuntime(matrix, new_sol))

		if acceptanceProbability(currentEnergy, newEnergy, temp) > rand.Float64() {
			curr_sol = new_sol
			currentEnergy = newEnergy
		}

		if newEnergy < bestEnergy {
			bestEnergy = newEnergy
			best_sol = new_sol
		}

		temp = temp * (1 - coolingRate)
	}

	return best_sol
}

func SA_swap(sol []int) []int {
	rand.Seed(time.Now().UnixNano())

	var i_1 int = rand.Intn(len(sol))
	var i_2 int = rand.Intn(len(sol))

	for i := 0; i_1 == i_2; i++ {
		i_2 = rand.Intn(len(sol))
	}

	tmpVal := sol[i_1]

	sol[i_1] = sol[i_2]
	sol[i_2] = tmpVal

	return sol

}

func copyIntArr(arr []int) []int {
	var copyArr []int = make([]int, len(arr))

	for i := 0; i < len(arr); i++ {
		copyArr[i] = arr[i]
	}

	return copyArr
}

func init_sol(len int) []int {
	var sol = make([]int, len)

	for i := 0; i < len; i++ {
		sol[i] = i
	}

	return sol
}

func acceptanceProbability(energy float64, newEnergy float64, temperature float64) float64 {
	if newEnergy < energy {
		return 1.0
	}

	return math.Exp((energy - newEnergy) / temperature)
}
