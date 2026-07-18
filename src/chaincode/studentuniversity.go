package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type StudentUniversityContract struct {
}

type agreement struct {
	ContractID     string  `json:"Key"`
	StudentName    string  `json:"StudentName"`
	UniversityName string  `json:"UniversityName"`
	Date           string  `json:"Date"`
	Amount         float64 `json:"Amount"`
	Email          string  `json:"Email"`
}

func (t *StudentUniversityContract) Init(stubInterface shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Student University Contract Initiated")

	_, args := stubInterface.GetFunctionAndParameters()

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments.")
	}
	return t.initStudentUniversity(stubInterface, args)
}

func (t *StudentUniversityContract) Invoke(stubInterface shim.ChaincodeStubInterface) peer.Response {
	function, args := stubInterface.GetFunctionAndParameters()
	fmt.Println("Invoke has started running " + function)

	// Handle different functions
	if function == "initStudentUniversity" {
		return t.initStudentUniversity(stubInterface, args)
	} else if function == "queryByStudentEmail" {
		return t.queryByStudentEmail(stubInterface, args)
	} else if function == "getHistoryForStudent" {
		return t.getHistoryForStudent(stubInterface, args)
	} else if function == "invokeFunctionStudentUniversity" {
		return t.invokeFunctionStudentUniversity(stubInterface, args)
	}
	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")

}

func (t *StudentUniversityContract) initStudentUniversity(stubInterface shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	fmt.Println("- start init Student University contract")

	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}

	studentName := strings.ToLower(args[0])
	studentEmail := strings.ToLower(args[1])
	currentDate := strings.ToLower(args[2])
	amount, err := strconv.ParseFloat(args[3], 64)
	universityName := strings.ToLower(args[4])
	if err != nil {
		return shim.Error("4th argument must be a numeric string")
	}

	hash := sha256.New()
	hash.Write([]byte(studentName + universityName))
	contractID := hex.EncodeToString(hash.Sum(nil))

	studentContract := &agreement{contractID, studentName, universityName, currentDate, amount, studentEmail}
	studentContractJSONasBytes, err := json.Marshal(studentContract)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stubInterface.PutState(contractID, studentContractJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end init studentContract" + "Transaction ID: " + stubInterface.GetTxID())
	return shim.Success(nil)
}

//Function
// =========================================================================================
func (t *StudentUniversityContract) queryByStudentEmail(stubInterface shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	studentEmail := strings.ToLower(args[0])
	selector := map[string]interface{}{
		"selector": map[string]string{"Email": studentEmail},
	}
	queryBytes, err := json.Marshal(selector)
	if err != nil {
		return shim.Error(err.Error())
	}

	queryResults, err := getQueryResultForQueryString(stubInterface, string(queryBytes))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *StudentUniversityContract) getHistoryForStudent(stubInterface shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	studentName := strings.ToLower(args[0])
	universityName := strings.ToLower(args[1])

	hash := sha256.New()
	hash.Write([]byte(studentName + universityName))
	contractID := hex.EncodeToString(hash.Sum(nil))

	fmt.Printf("- start getHistoryForStudent: %s\n", contractID)

	resultsIterator, err := stubInterface.GetHistoryForKey(contractID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		fmt.Printf("- getHistoryForStudent response: %v\n", response)
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForStudent returning:\n%s\n", buffer.String())

	return shim.Success([]byte(buffer.String()))
}

func (t *StudentUniversityContract) invokeFunctionStudentUniversity(stubInterface shim.ChaincodeStubInterface, args []string) peer.Response {
	fmt.Println("Start invokeFunctionStudentUniversity")
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	var mykey, jsonResp string
	var err error
	mykey = args[0]
	//newGasAmount := args[1]
	valAsbytes, err := stubInterface.GetState(string(mykey))
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + mykey + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"invokeFunctionStudentUniversity does not exist: " + mykey + "\"}"
		return shim.Error(jsonResp)
	}
	fmt.Println(valAsbytes)
	MYagreement := agreement{}
	//umarshal the data to a new ballot struct
	if err = json.Unmarshal(valAsbytes, &MYagreement); err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println(MYagreement)
	return shim.Success(valAsbytes)

}

// =========================================================================================
func getQueryResultForQueryString(stubInterface shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Println("Inside getQueryResultForQueryString")

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stubInterface.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(StudentUniversityContract))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
