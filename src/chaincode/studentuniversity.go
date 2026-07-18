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

	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

const (
	statusSubmitted = "submitted"
	statusApproved  = "approved"
	statusRejected  = "rejected"
)

type StudentUniversityContract struct{}

type agreement struct {
	ContractID     string  `json:"Key"`
	StudentName    string  `json:"StudentName"`
	UniversityName string  `json:"UniversityName"`
	Date           string  `json:"Date"`
	Amount         float64 `json:"Amount"`
	Email          string  `json:"Email"`
	DocumentHash   string  `json:"DocumentHash"`
	Status         string  `json:"Status"`
	CreatedBy      string  `json:"CreatedBy"`
	ReviewedBy     string  `json:"ReviewedBy,omitempty"`
	UpdatedAt      string  `json:"UpdatedAt"`
}

type queryRecord struct {
	Key   string          `json:"Key"`
	Value json.RawMessage `json:"Value"`
}

type historyRecord struct {
	TxID      string          `json:"TxId"`
	Value     json.RawMessage `json:"Value"`
	Timestamp string          `json:"Timestamp"`
	IsDelete  bool            `json:"IsDelete"`
}

func (t *StudentUniversityContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 5 && len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 5 or 6")
	}
	return t.createAgreement(stub, args, true)
}

func (t *StudentUniversityContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()

	switch function {
	case "initStudentUniversity", "createAgreement":
		return t.createAgreement(stub, args, false)
	case "queryByStudentEmail":
		return t.queryByStudentEmail(stub, args)
	case "queryAllAgreements":
		return t.queryAllAgreements(stub, args)
	case "getHistoryForStudent":
		return t.getHistoryForStudent(stub, args)
	case "getHistoryForAgreement":
		return t.getHistoryForAgreement(stub, args)
	case "invokeFunctionStudentUniversity", "getAgreement":
		return t.getAgreement(stub, args)
	case "reviewAgreement":
		return t.reviewAgreement(stub, args)
	case "verifyDocument":
		return t.verifyDocument(stub, args)
	default:
		return shim.Error("Received unknown function invocation: " + function)
	}
}

func (t *StudentUniversityContract) createAgreement(stub shim.ChaincodeStubInterface, args []string, lifecycleInit bool) peer.Response {
	if len(args) != 5 && len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 5 or 6")
	}
	for i, arg := range args[:5] {
		if strings.TrimSpace(arg) == "" {
			return shim.Error(fmt.Sprintf("Argument %d must be a non-empty string", i+1))
		}
	}

	amount, err := strconv.ParseFloat(args[3], 64)
	if err != nil || amount <= 0 {
		return shim.Error("4th argument must be a positive numeric string")
	}

	studentName := strings.ToLower(strings.TrimSpace(args[0]))
	studentEmail := strings.ToLower(strings.TrimSpace(args[1]))
	currentDate := strings.TrimSpace(args[2])
	if _, err = time.Parse("2006-01-02", currentDate); err != nil {
		return shim.Error("3rd argument must be an ISO date in YYYY-MM-DD format")
	}
	universityName := strings.ToLower(strings.TrimSpace(args[4]))
	documentHash := ""
	if len(args) == 6 {
		documentHash = strings.ToLower(strings.TrimSpace(args[5]))
		if documentHash != "" && !isSHA256(documentHash) {
			return shim.Error("6th argument must be an empty value or a SHA-256 hash")
		}
	}

	creatorMSP, err := cid.GetMSPID(stub)
	if err != nil {
		return shim.Error("Unable to identify transaction creator: " + err.Error())
	}
	if !lifecycleInit && creatorMSP != "StudentMSP" {
		return shim.Error("Only a StudentMSP member may submit agreements")
	}
	updatedAt, err := transactionTime(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	contractID := agreementID(studentName, universityName)
	existing, err := stub.GetState(contractID)
	if err != nil {
		return shim.Error(err.Error())
	}
	if existing != nil {
		return shim.Error("An agreement already exists for this student and university")
	}

	record := agreement{
		ContractID:     contractID,
		StudentName:    studentName,
		UniversityName: universityName,
		Date:           currentDate,
		Amount:         amount,
		Email:          studentEmail,
		DocumentHash:   documentHash,
		Status:         statusSubmitted,
		CreatedBy:      creatorMSP,
		UpdatedAt:      updatedAt,
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return shim.Error(err.Error())
	}
	if err = stub.PutState(contractID, payload); err != nil {
		return shim.Error(err.Error())
	}
	if err = stub.SetEvent("AgreementSubmitted", payload); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(payload)
}

func (t *StudentUniversityContract) reviewAgreement(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting agreement ID and decision")
	}
	decision := strings.ToLower(strings.TrimSpace(args[1]))
	if decision != statusApproved && decision != statusRejected {
		return shim.Error("Decision must be approved or rejected")
	}

	reviewerMSP, err := cid.GetMSPID(stub)
	if err != nil {
		return shim.Error("Unable to identify reviewer: " + err.Error())
	}
	if reviewerMSP != "UniversityMSP" {
		return shim.Error("Only a UniversityMSP member may review agreements")
	}

	record, err := loadAgreement(stub, strings.TrimSpace(args[0]))
	if err != nil {
		return shim.Error(err.Error())
	}
	if record.Status != "" && record.Status != statusSubmitted {
		return shim.Error("Only submitted agreements may be reviewed")
	}
	record.Status = decision
	record.ReviewedBy = reviewerMSP
	record.UpdatedAt, err = transactionTime(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return shim.Error(err.Error())
	}
	if err = stub.PutState(record.ContractID, payload); err != nil {
		return shim.Error(err.Error())
	}
	if err = stub.SetEvent("AgreementReviewed", payload); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(payload)
}

func (t *StudentUniversityContract) verifyDocument(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting agreement ID and SHA-256 hash")
	}
	hash := strings.ToLower(strings.TrimSpace(args[1]))
	if !isSHA256(hash) {
		return shim.Error("Document hash must be a SHA-256 hash")
	}
	record, err := loadAgreement(stub, strings.TrimSpace(args[0]))
	if err != nil {
		return shim.Error(err.Error())
	}
	result := map[string]interface{}{
		"agreementId": record.ContractID,
		"verified":    record.DocumentHash != "" && record.DocumentHash == hash,
		"status":      record.Status,
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(payload)
}

func (t *StudentUniversityContract) queryByStudentEmail(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
		return shim.Error("Incorrect number of arguments. Expecting one email")
	}
	selector := map[string]interface{}{
		"selector": map[string]string{"Email": strings.ToLower(strings.TrimSpace(args[0]))},
	}
	queryBytes, err := json.Marshal(selector)
	if err != nil {
		return shim.Error(err.Error())
	}
	results, err := queryByString(stub, string(queryBytes))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(results)
}

func (t *StudentUniversityContract) queryAllAgreements(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 0 {
		return shim.Error("queryAllAgreements does not accept arguments")
	}
	iterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer iterator.Close()

	records := make([]queryRecord, 0)
	for iterator.HasNext() {
		result, err := iterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		records = append(records, queryRecord{Key: result.Key, Value: json.RawMessage(result.Value)})
	}
	payload, err := json.Marshal(records)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(payload)
}

func (t *StudentUniversityContract) getAgreement(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
		return shim.Error("Incorrect number of arguments. Expecting one agreement ID")
	}
	record, err := loadAgreement(stub, strings.TrimSpace(args[0]))
	if err != nil {
		return shim.Error(err.Error())
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(payload)
}

func (t *StudentUniversityContract) getHistoryForStudent(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting student and university names")
	}
	contractID := agreementID(
		strings.ToLower(strings.TrimSpace(args[0])),
		strings.ToLower(strings.TrimSpace(args[1])),
	)
	return agreementHistory(stub, contractID)
}

func (t *StudentUniversityContract) getHistoryForAgreement(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
		return shim.Error("Incorrect number of arguments. Expecting one agreement ID")
	}
	return agreementHistory(stub, strings.TrimSpace(args[0]))
}

func agreementHistory(stub shim.ChaincodeStubInterface, contractID string) peer.Response {
	iterator, err := stub.GetHistoryForKey(contractID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer iterator.Close()

	records := make([]historyRecord, 0)
	for iterator.HasNext() {
		result, err := iterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		value := json.RawMessage("null")
		if !result.IsDelete {
			value = json.RawMessage(result.Value)
		}
		records = append(records, historyRecord{
			TxID:      result.TxId,
			Value:     value,
			Timestamp: time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos)).UTC().Format(time.RFC3339),
			IsDelete:  result.IsDelete,
		})
	}
	payload, err := json.Marshal(records)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(payload)
}

func queryByString(stub shim.ChaincodeStubInterface, query string) ([]byte, error) {
	iterator, err := stub.GetQueryResult(query)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	records := make([]queryRecord, 0)
	for iterator.HasNext() {
		result, err := iterator.Next()
		if err != nil {
			return nil, err
		}
		records = append(records, queryRecord{Key: result.Key, Value: json.RawMessage(result.Value)})
	}
	return json.Marshal(records)
}

func loadAgreement(stub shim.ChaincodeStubInterface, contractID string) (*agreement, error) {
	payload, err := stub.GetState(contractID)
	if err != nil {
		return nil, err
	}
	if payload == nil {
		return nil, fmt.Errorf("agreement %s does not exist", contractID)
	}
	record := &agreement{}
	if err = json.Unmarshal(payload, record); err != nil {
		return nil, err
	}
	return record, nil
}

func agreementID(studentName, universityName string) string {
	hash := sha256.Sum256([]byte(studentName + universityName))
	return hex.EncodeToString(hash[:])
}

func isSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func transactionTime(stub shim.ChaincodeStubInterface) (string, error) {
	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return "", fmt.Errorf("unable to read transaction timestamp: %s", err)
	}
	var buffer bytes.Buffer
	buffer.WriteString(time.Unix(timestamp.Seconds, int64(timestamp.Nanos)).UTC().Format(time.RFC3339))
	return buffer.String(), nil
}

func main() {
	if err := shim.Start(new(StudentUniversityContract)); err != nil {
		fmt.Printf("Error starting Yakusoku chaincode: %s", err)
	}
}
