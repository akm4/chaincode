package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	//"strings"
	//"crypto/sha256"
	//"reflect"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var (
	// prefix for saving person data
	personPrfx = "Person:"
	// prefix for saving history of person
	personHistoryPrfx = "PersonHistory:"
)

const ACTION_INSERT = "create"
const ACTION_UPDATE = "update"
const ACTION_DELETE = "delete"
const ACTION_SEARCH = "search"

const STATUS_OK = "ok"
const STATUS_SUSP = "suspicious"
const STATUS_DELETED = "deleted"

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//type for audit record
type Action struct {
	InsuranceCompany string    `json:"insuranceCompany"`
	User             string    `json:"user"`
	Method           string    `json:"method"`
	Date             time.Time `json:"date"`
	NewStatus        string    `json:newStatus`
}

//type for person data
type Person struct {
	Hash       string    `json:hash`
	Status     string    `json:"status"`
	ModifyDate time.Time `json:"modifyDate"`
}

//type for result of search request
type SearchResult struct {
	Existence     bool     `json:existence`
	CurrentStatus string   `json:currentStatus`
	History       []Action `json:history`
}

//---------------------------------------------------- MAIN
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

//---------------------------------------------------SHIM - INIT
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//check arguments length
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	return nil, nil
}

//--------------------------------------------------- SHIM - QUERY
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// Handle different functions
	if function == "getPersonInfo" { //read person by hash
		return t.getPersonInfo(stub, args)
	} else if function == "read" { // read data by name from state
		return t.read(stub, args)
	} else if function == "getPersonHistory" { // read history of person from state
		return t.getPersonHistory(stub, args)
	}
	//	 else if function == "calculateHash" {
	//		return t.calculatePersonHash(stub, args)
	//	}

	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query")
}

//---------------------------------------------SHIM - INVOKE
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Invoke is running this function :" + function)
	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function == "insertPerson" { //create a new person
		return t.insertPerson(stub, args)
	} else if function == "deletePerson" { //delete a person
		return t.deletePerson(stub, args)
	} else if function == "updatePerson" { // update a person
		return t.updatePerson(stub, args)
	} else if function == "searchPerson" {
		return t.searchPerson(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

//--------------------calculatePersonHash. Params:
// firstName
// middleName
// lastName
// day
// month
// year
//func (t *SimpleChaincode) calculatePersonHash(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
//	if len(args) < 6 {
//		return nil, errors.New("Incorrect number of arguments. Expecting 6")
//	}
//	firstName := args[0]
//	middleName := args[1]
//	lastName := args[2]
//	day := args[3]
//	month := args[4]
//	year := args[5]
//	//concat params
//	return calcHashByParams(firstName, middleName, lastName, day, month, year), nil
//}

func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	key := args[0]
	return stub.GetState(key)
}

func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	key := args[0]
	val := []byte(args[1])
	return nil, stub.PutState(key, val)
}

//print person data by hash
func (t *SimpleChaincode) getPersonInfo(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 1
	if len(args) != 1 {
		return nil, errors.New("incorrect number of arguments. need 1")
	}
	//hash := args[0]
	argsMap, err := getUnmarshalledArgument(args)
	if err != nil {
		return nil, err
	}
	hash, err := getStringParamFromArgs("hash", argsMap)
	if err != nil {
		return nil, err
	}
	fmt.Println("get info fro person " + hash)
	//get person from state
	res, err := stub.GetState(personPrfx + hash)
	return res, err
}

//print person history by hash
func (t *SimpleChaincode) getPersonHistory(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 1
	if len(args) != 1 {
		return nil, errors.New("incorrect number of arguments. need 1")
	}
	hash := args[0]
	//get person from state
	res, err := stub.GetState(personHistoryPrfx + hash)
	return res, err
}

func (t *SimpleChaincode) insertPerson(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 4
	if len(args) < 4 {
		return nil, errors.New("incorrect number of arguments. need 4")
	}
	hash := args[0]
	user := args[1]
	insComp := args[2]
	status := args[3]
	//-----add person hash to state
	newPerson := &Person{}
	newPerson.ModifyDate = time.Now()
	newPerson.Status = status
	newPerson.Hash = hash
	err := createOrUpdatePerson(stub, hash, *newPerson)
	if err != nil {
		return nil, errors.New("error inserting person")
	}
	//------add record to person history
	err = addHistoryRecord(stub, hash, ACTION_INSERT, user, insComp, status)
	if err != nil {
		return nil, errors.New("Error putting new history record " + hash + " to state")
	}
	return nil, nil
}

func (t *SimpleChaincode) updatePerson(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 4
	if len(args) < 4 {
		return nil, errors.New("incorrect number of arguments. need 4")
	}
	hash := args[0]
	user := args[1]
	insComp := args[2]
	status := args[3]
	//-----add person hash to state
	newPerson := &Person{}
	newPerson.Hash = hash
	newPerson.ModifyDate = time.Now()
	newPerson.Status = status
	err := createOrUpdatePerson(stub, hash, *newPerson)
	if err != nil {
		return nil, errors.New("error updating person")
	}
	//------add record to person history
	err = addHistoryRecord(stub, hash, ACTION_UPDATE, user, insComp, status)
	if err != nil {
		return nil, errors.New("Error putting new history record " + hash + " to state")
	}

	return nil, nil
}

func (t *SimpleChaincode) deletePerson(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 3
	if len(args) < 3 {
		return nil, errors.New("incorrect number of arguments. need 3")
	}
	hash := args[0]
	user := args[1]
	insComp := args[2]
	//TODO maybe check existence ???
	//-----add person hash to state
	newPerson := &Person{}
	newPerson.Hash = hash
	newPerson.ModifyDate = time.Now()
	newPerson.Status = STATUS_DELETED
	err := createOrUpdatePerson(stub, hash, *newPerson)
	if err != nil {
		return nil, errors.New("error updating person")
	}
	//------add record to person history
	err = addHistoryRecord(stub, hash, ACTION_DELETE, user, insComp, STATUS_DELETED)
	if err != nil {
		return nil, errors.New("Error putting new history record " + hash + " to state")
	}
	return nil, nil
}

func (t *SimpleChaincode) searchPerson(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	//parse parameters  - need 3
	if len(args) < 3 {
		return nil, errors.New("incorrect number of arguments. need 3")
	}
	hash := args[0]
	user := args[1]
	insComp := args[2]

	res := &SearchResult{}
	//check existence
	var found bool
	//retrieve Person from state by hash
	personBytes, err := stub.GetState(personPrfx + hash)
	found = err == nil && len(personBytes) != 0
	if found { // if exists
		var oldperson Person
		err = json.Unmarshal(personBytes, &oldperson)
		if err != nil {
			return nil, errors.New("Error unmarshalling person " + hash + " from state")
		}
		//fill response record
		res.CurrentStatus = oldperson.Status
		res.Existence = true
	}
	//getHistory for result
	var history []Action
	historyBytes, err := stub.GetState(personHistoryPrfx + hash)
	if err != nil {
		return nil, errors.New("Error getting history for person " + hash)
	}
	if historyBytes != nil {
		err = json.Unmarshal(historyBytes, &history)
		if err != nil {
			return nil, errors.New("Error unmarshalling history for person " + hash)
		}
	}
	//assign to result
	res.History = history
	//add record to history
	err = addHistoryRecord(stub, hash, ACTION_SEARCH, user, insComp, strconv.FormatBool(found))
	if err != nil {
		return nil, err
	}
	resBytes, _ := json.Marshal(res)
	return resBytes, nil
}

/*****************
stub shim.ChaincodeStubInterface
hash string
action string
user string
insuranceCompany string
status string
******************/
func addHistoryRecord(stub shim.ChaincodeStubInterface, hash string, action string, user string, insuranceCompany string, status string) error {
	var history []Action
	historyBytes, err := stub.GetState(personHistoryPrfx + hash)
	//chaincodeLogger.Info("current list: " + string(historyBytes))
	fmt.Println("current list: " + string(historyBytes))
	if err != nil {
		return errors.New("Error getting history for person ")
	}
	if historyBytes != nil {
		err = json.Unmarshal(historyBytes, &history)
		if err != nil {
			return errors.New("Error unmarshalling history for person ")
		}
	}
	newAction := &Action{}
	newAction.NewStatus = status
	newAction.Method = action
	newAction.User = user
	newAction.Date = time.Now()
	newAction.InsuranceCompany = insuranceCompany
	//insert action to history
	history = append(history, *newAction)
	//put history to state
	newHistoryBytes, err := json.Marshal(&history)
	if err != nil {
		return errors.New("Error parsing history for person " + hash)
	}
	err = stub.PutState(personHistoryPrfx+hash, newHistoryBytes)
	if err != nil {
		return errors.New("Error parsing history for person" + hash)
	}
	return nil
}

func createOrUpdatePerson(stub shim.ChaincodeStubInterface, hash string, newPerson Person) error {
	var oldPerson Person
	//retrieve Person from state by hash
	personBytes, err := stub.GetState(personPrfx + hash)
	if err != nil || len(personBytes) == 0 {
		//data not found, create scenario
		oldPerson = newPerson
	} else {
		//update scenario
		err = json.Unmarshal(personBytes, &oldPerson)
		if err != nil {
			return errors.New("error unmarshalling person from state")
		}
		//TODO merge data, now only replace
		oldPerson = newPerson
		if err != nil {
			return errors.New("error mergin data of Person")
		}
	}
	//put Person in state
	err = putPersonInState(stub, hash, oldPerson)
	if err != nil {
		return err
	}
	return nil
}

//func mergePerson(oldPerson Person, newPerson Person) (Person, error) {
//	o := reflect.ValueOf(&oldPerson).Elem()
//	n := reflect.ValueOf(&newPerson).Elem()
//	for i := 0; i < o.NumField(); i++ {
//		oldOne := o.Field(i)
//		newOne := n.Field(i)
//		if !reflect.ValueOf(newOne.Interface()).IsNil() {
//			oldOne.Set(reflect.Value(newOne))
//		}
//	}
//	return oldPerson, nil
//}

func putPersonInState(stub shim.ChaincodeStubInterface, hash string, person Person) error {
	personAsBytes, err := json.Marshal(&person)
	if err != nil {
		return errors.New("Error marhalling new person")
	}
	err = stub.PutState(personPrfx+hash, personAsBytes)
	if err != nil {
		return errors.New("Error puttin new person")
	}
	fmt.Println("put record for " + hash)
	return nil
}

//func calcHashByParams(firstName string, middleName string, lastName string, day string, month string, year string) []byte {
//	return calcHash(firstName + middleName + lastName + day + month + year)
//}

//func calcHash(input string) []byte {
//	hash := sha256.New()
//	hash.Write([]byte(input))
//	md := hash.Sum(nil)
//	return md
//}
