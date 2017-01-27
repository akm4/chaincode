package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	//"strconv"
	//"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"github.com/op/go-logging"
)

//logger - not in v0.6
//var chaincodeLogger = logging.MustGetLogger("insurance")

var (
	// prefix for saving client data
	clientPrfx = "Client:"
	// prefix for saving history of client
	clientHistoryPrfx = "ClientHistory:"
	//key for saving client List
	clientListKey = "ClientListKey"
)

const ACTION_INSERT = "insert"
const ACTION_UPDATE = "update"
const ACTION_DELETE = "delete"

const STATUS_DELETED = "deleted"

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Action struct {
	InsuranceCompany string    `json:"insuranceCompany"`
	User             string    `json:"user"`
	Method           string    `json:"method"`
	Date             time.Time `json:"date"`
	NewStatus        string    `json:newStatus`
}
type Client struct {
	Status     string    `json:"status"`
	ModifyDate time.Time `json:"modifyDate"`
	Hash       string    `json:hash`
	//History    []Action  `json:"history"`
}

// MAIN
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		//chaincodeLogger.Error("Error starting Simple chaincode: %s", err)
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

//SHIM - INIT
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//check arguments length
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	//TODO reset all list
	var blank []string
	blankBytes, _ := json.Marshal(&blank)
	err := stub.PutState(clientListKey, blankBytes)
	if err != nil {
		//chaincodeLogger.Error("Failed to initialize client list")
		fmt.Println("Failed to initialize client list")
	}
	return nil, nil

}

// SHIM - QUERY
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// Handle different functions
	if function == "printClient" { //read client by hash
		return t.printClient(stub, args)
	} else if function == "printClientList" { // read all clients hash from state
		return t.printClientList(stub, args)
	} else if function == "read" { // read data by name from state
		return t.read(stub, args)
	} else if function == "printClientHistory" { // read history of client from state
		return t.printClientHistory(stub, args)
	}

	//chaincodeLogger.Error("query did not find func: " + function) //error
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query")
}

//SHIM - INVOKE
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//chaincodeLogger.Info("Invoke is running this function :" + function)
	fmt.Println("Invoke is running this function :" + function)
	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function == "insertClient" { //create a new client
		return t.insertClient(stub, args)
	} else if function == "deleteClient" { //delete a client
		return t.deleteClient(stub, args)
	} else if function == "updateClient" { // update a client
		return t.updateClient(stub, args)
	}
	return nil, errors.New("Received unknown function invocation")
}

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

//print client data by hash
func (t *SimpleChaincode) printClient(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 1
	if len(args) != 1 {
		return nil, errors.New("incorrect number of arguments. need 1")
	}
	hash := args[0]
	//get client from state
	res, err := stub.GetState(clientPrfx + hash)
	return res, err
}

//print client history by hash
func (t *SimpleChaincode) printClientHistory(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 1
	if len(args) != 1 {
		return nil, errors.New("incorrect number of arguments. need 1")
	}
	hash := args[0]
	//get client from state
	res, err := stub.GetState(clientHistoryPrfx + hash)
	return res, err
}

// print client list
func (t *SimpleChaincode) printClientList(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	res, err := stub.GetState(clientListKey)
	if err != nil {
		return nil, err
	}
	return res, err
}

func (t *SimpleChaincode) insertClient(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 4
	if len(args) < 4 {
		return nil, errors.New("incorrect number of arguments. need 4")
	}
	hash := args[0]
	status := args[1]
	user := args[2]
	insComp := args[3]
	//------ check client in list
	found, clientIndex, err := checkClientInClientList(stub, hash)
	if err != nil {
		return nil, errors.New("Error checking existance of " + hash + " :" + err.Error())
	}
	if found {
		//TODO maybe replace instead of return error ???
		return nil, errors.New("client " + hash + " already exists")
	}
	//-----add client hash to state
	//create new client record
	newClient := &Client{}
	newClient.ModifyDate = time.Now()
	//TODO maybe need const = SUSP ???
	newClient.Status = status
	newClient.Hash = hash
	newClientAsBytes, err := json.Marshal(&newClient)
	if err != nil {
		return nil, errors.New("Error creating new client")
	}
	err = stub.PutState(clientPrfx+hash, newClientAsBytes)
	if err != nil {
		return nil, errors.New("Error creating new client")
	}
	//------add record to client history
	err = addHistoryRecord(stub, hash, ACTION_INSERT, user, insComp, status)
	if err != nil {
		return nil, errors.New("Error putting new history record " + hash + " to state")
	}
	//------add client to client list
	clientIndex = append(clientIndex, hash)
	clientListAsBytesToWrite, err := json.Marshal(&clientIndex)
	if err != nil {
		return nil, errors.New("Error marshalling the client list")
	}
	err = stub.PutState(clientListKey, clientListAsBytesToWrite)
	if err != nil {
		return nil, errors.New("Error saving the client list")
	}

	return nil, nil
}

func (t *SimpleChaincode) updateClient(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 4
	if len(args) < 4 {
		return nil, errors.New("incorrect number of arguments. need 4")
	}
	hash := args[0]
	status := args[1]
	user := args[2]
	insComp := args[3]
	//--- check existance
	found, _, err := checkClientInClientList(stub, hash)
	if err != nil {
		return nil, errors.New("Error checking existance of " + hash + " :" + err.Error())
	}
	if !found {
		return nil, errors.New("Client " + hash + " dosn't exist")
	}
	//-- update client record
	// get client from state
	clientAsBytes, err := stub.GetState(clientPrfx + hash)
	if err != nil {
		return nil, errors.New("Error getting client " + hash + " from state")
	}
	var oldClient Client
	err = json.Unmarshal(clientAsBytes, &oldClient)
	if err != nil {
		return nil, errors.New("Error unmarshalling client " + hash + " from state")
	}
	oldClient.ModifyDate = time.Now()
	//TODO need analyze correct status
	oldClient.Status = STATUS_DELETED
	//put client record to state
	clientAsBytes, err = json.Marshal(&oldClient)
	if err != nil {
		return nil, errors.New("Error marshalling updated client " + hash)
	}
	err = stub.PutState(clientPrfx+hash, clientAsBytes)
	if err != nil {
		return nil, errors.New("Error putting updated client " + hash + " to state")
	}
	//--add history record
	err = addHistoryRecord(stub, hash, ACTION_UPDATE, user, insComp, status)
	if err != nil {
		return nil, errors.New("Error putting new history record " + hash + " to state")
	}
	return nil, nil
}

func (t *SimpleChaincode) deleteClient(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 3
	if len(args) < 3 {
		return nil, errors.New("incorrect number of arguments. need 3")
	}
	hash := args[0]
	user := args[1]
	insComp := args[2]

	//-- check existance
	found, clientIndex, err := checkClientInClientList(stub, hash)
	if err != nil {
		return nil, errors.New("Error checking existance of " + hash + " :" + err.Error())
	}
	if !found {
		return nil, errors.New("Client " + hash + " doesn't exist")
	}
	//-- update client record
	// get client from state
	clientAsBytes, err := stub.GetState(clientPrfx + hash)
	if err != nil {
		return nil, errors.New("Error getting client " + hash + " from state")
	}
	var oldClient Client
	err = json.Unmarshal(clientAsBytes, &oldClient)
	if err != nil {
		return nil, errors.New("Error unmarshalling client " + hash + " from state")
	}
	oldClient.ModifyDate = time.Now()
	//TODO need analyze if already deleted ???
	oldClient.Status = STATUS_DELETED
	//put client record to state
	clientAsBytes, err = json.Marshal(&oldClient)
	if err != nil {
		return nil, errors.New("Error marshalling updated client " + hash)
	}
	err = stub.PutState(clientPrfx+hash, clientAsBytes)
	if err != nil {
		return nil, errors.New("Error putting updated client " + hash + " to state")
	}
	//--add history record
	err = addHistoryRecord(stub, hash, ACTION_DELETE, user, insComp, STATUS_DELETED)
	if err != nil {
		return nil, errors.New("Error putting new history record " + hash + " to state")
	}
	//delete from client list
	for i := range clientIndex {
		if clientIndex[i] == hash {
			clientIndex = append(clientIndex[:i], clientIndex[i+1:]...)
			clientIndexAsBytes, _ := json.Marshal(clientIndex)
			err = stub.PutState(clientListKey, clientIndexAsBytes)
			if err != nil {
				return nil, errors.New("Error deleting  record for " + hash + " from state")
			}
			break
		}
	}
	return nil, nil
}

func checkClientInClientList(stub shim.ChaincodeStubInterface, hash string) (bool, []string, error) {
	var clientIndex []string
	var found bool
	clientListAsBytes, err := stub.GetState(clientListKey)
	//chaincodeLogger.Info("current list: " + string(clientListAsBytes))
	fmt.Println("current list: " + string(clientListAsBytes))
	if err != nil {
		return found, nil, errors.New("failed get client list")
	}
	if clientListAsBytes != nil {
		err = json.Unmarshal(clientListAsBytes, &clientIndex)
		if err != nil {
			return found, nil, errors.New("failed unmarshalling client list")
		}
		for _, v := range clientIndex {
			if v == hash {
				found = true
				break
			}
		}
	}
	return found, clientIndex, nil
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
	historyBytes, err := stub.GetState(clientHistoryPrfx + hash)
	//chaincodeLogger.Info("current list: " + string(historyBytes))
	fmt.Println("current list: " + string(historyBytes))
	if err != nil {
		return errors.New("Error getting history for client")
	}
	if historyBytes != nil {
		err = json.Unmarshal(historyBytes, &history)
		if err != nil {
			return errors.New("Error unmarshalling history for client")
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
		return errors.New("Error parsing history for client " + hash)
	}
	err = stub.PutState(clientHistoryPrfx+hash, newHistoryBytes)
	if err != nil {
		return errors.New("Error parsing history for client" + hash)
	}
	return nil
}

//------------------------- TODO list----------------------
// ++ 1.  Add logger
// 2.  Add method for search (with history record )
// ++ 3.  Covert string value to const
// ++ 4.  add method for client "delete"
// ++ 5.  add method for replacing code for client existance to function
// ++ 6. make method for  client update
// 7. Delete State  istead of remove form array in Delete method - need check version of HL
// 8. Refactor Delete method - change to update+delete from list
// 9. Refactor insert and update methods to isertAndUpdate
//---------------------------------------------------------
