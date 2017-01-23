package main

import (
	"encoding/json"
	"errors"
	"fmt"
	//"strconv"
	//"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var (
	// prefix for saving client data
	clientPrfx = "Client:"
	// prefix for saving history of client
	clientHistoryPrfx = "ClientHistory:"
	//key for saving client List
	clientListKey = "ClientListKey"
)

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

//type HistoryList struct {
//	History []Action `json:"history"`
//}

//type ClientList struct {
//	Storage []string `json:"storage"`
//}

// MAIN
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
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
		fmt.Println("Failed to initialize client list")
	}

	return nil, nil

}

// SHIM - QUERY
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// Handle different functions
	if function == "printClient" { //read client by hash
		return t.printClient(stub, args)
	} else if function == "printClientList" { // read by name from stste
		return t.printClientList(stub, args)
	} else if function == "read" { // read by name from stste
		return t.read(stub, args)
	} else if function == "printClientHistory" { // read by name from stste
		return t.printClientHistory(stub, args)
	}

	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query")
}

//SHIM - INVOKE
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)
	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function == "insertClient" { //create a new client
		return t.insertClient(stub, args)
	}
	//	} else if function == "updateClient" { //update a client
	//		return t.updateClient(stub, args)
	//	} else if function == "deleteClient" { //delete a client
	//		return t.deleteClient(stub, args)

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
	//------check client in list
	var clientIndex []string
	var found bool
	clientListAsBytes, err := stub.GetState(clientListKey)
	fmt.Println("current list: " + string(clientListAsBytes))
	if err != nil {
		return nil, errors.New("failed get client list")
	}
	if clientListAsBytes != nil {
		err = json.Unmarshal(clientListAsBytes, &clientIndex)
		if err != nil {
			return nil, errors.New("failed unmarshalling client list")
		}
		for _, v := range clientIndex {
			if v == hash {
				found = true
				break
			}
		}
	}
	if found {
		//TODO maybe update instead of error ???
		return nil, errors.New("client " + hash + "  already exists")
	}
	//-----add client hash to state
	//create new client record
	newClient := &Client{}
	newClient.ModifyDate = time.Now()
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
	//get history from state
	var history []Action
	historyBytes, err := stub.GetState(clientHistoryPrfx + hash)
	fmt.Println("current list: " + string(historyBytes))
	if err != nil {
		return nil, errors.New("Error getting history for client")
	}
	if historyBytes != nil {
		err = json.Unmarshal(historyBytes, &history)
		if err != nil {
			return nil, errors.New("Error unmarshalling history for client")
		}
	}
	newAction := &Action{}
	newAction.NewStatus = status
	newAction.Method = "insert"
	newAction.User = user
	newAction.Date = time.Now()
	newAction.InsuranceCompany = insComp
	//insert action to history
	history = append(history, *newAction)
	//put history to state
	newHistoryBytes, err := json.Marshal(&history)
	if err != nil {
		return nil, errors.New("Error parsing history for client")
	}
	err = stub.PutState(clientHistoryPrfx+hash, newHistoryBytes)
	if err != nil {
		return nil, errors.New("Error parsing history for client")
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

//func (t *SimpleChaincode) updateClient(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
//	//parse parameters  - need 4
//	if len(args) < 4 {
//		return nil, errors.New("incorrect number of arguments. need 4")
//	}
//	hash := args[0]
//	status := args[1]
//	user := args[2]
//	insComp := args[3]

//	findClient, ok := clientList[hash]
//	//get client by hash
//	if !ok {
//		return nil, errors.New("client " + hash + "not exists")
//	}
//	//TODO check if next state is correct (ie delete to ok)
//	findClient.Status = status
//	findClient.ModifyDate = time.Now()
//	findClient.makeAction("update", status, user, insComp)
//	return nil, nil
//}

//func (t *SimpleChaincode) deleteClient(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
//	//parse parameters  - need 3
//	if len(args) < 3 {
//		return nil, errors.New("incorrect number of arguments. need 3")
//	}
//	hash := args[0]
//	user := args[1]
//	insComp := args[2]

//	findClient, ok := clientList[hash]
//	//get client by hash
//	if !ok {
//		return nil, errors.New("client " + hash + "not exists")
//	}
//	findClient.Status = "deleted"
//	findClient.ModifyDate = time.Now()
//	findClient.makeAction("delete", "deleted", user, insComp)
//	return nil, nil
//}

////create action record for client request history
//func (client *Client) makeAction(actionMethod string, status string, user string, insuranceCompany string) {
//	//update status
//	client.Status = status
//	//update actual date
//	client.ModifyDate = time.Now()
//	//create action
//	action := Action{insuranceCompany, user, actionMethod, time.Now(), status}
//	//add to history
//	client.History = append(client.History, action)
//}
