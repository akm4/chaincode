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
	History    []Action  `json:"history"`
}

var (
	//list of all clients
	clientList map[string]*Client
)

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
	//reset list
	clientList = make(map[string]*Client)
	return nil, nil

}

//SHIM - INVOKE
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)
	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "insertClient" { //create a new client
		return t.insertClient(stub, args)
	} else if function == "updateClient" { //update a client
		return t.updateClient(stub, args)
	} else if function == "deleteClient" { //delete a client
		return t.deleteClient(stub, args)
	} else if function == "makeMultiplePutState" {
		return t.makeMultiplePutState(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

// SHIM - QUERY
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// Handle different functions
	if function == "printClient" { //read client by hash
		return t.printClient(stub, args)
	} else if function == "printAllClients" { // read all clients
		return t.printAllClients(stub, args)
	} else if function == "readValueFromState" { // read by name from stste
		return t.readValueFromState(stub, args)
	}
	fmt.Println("query did not find func: " + function) //error
	return nil, errors.New("Received unknown function query")
}

// TODO delete this function
func (t *SimpleChaincode) makeMultiplePutState(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	for index, value := range args {
		key := "index" + string(index)
		val := []byte(value)
		fmt.Println("put to state index = " + key + " value = " + string(val))
		stub.PutState(key, val)
	}
	return nil, nil
}

// TODO delete this function
func (t *SimpleChaincode) readValueFromState(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	fmt.Println("try to get value by key = " + name)
	valAsbytes, err := stub.GetState(name) //get the var from chaincode state
	if err != nil {
		fmt.Println("ERROR")
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}
	fmt.Println("get VALUE = " + string(valAsbytes))
	return valAsbytes, nil
}

func (t *SimpleChaincode) printAllClients(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	jsonAsBytes, err := json.Marshal(clientList)
	return jsonAsBytes, err
}
func (t *SimpleChaincode) printClient(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 4
	if len(args) < 1 {
		return nil, errors.New("incorrect number of arguments. need 1")
	}
	hash := args[0]
	res, ok := clientList[hash]
	//check client by hash
	if !ok {
		return nil, errors.New("not found")
	}
	jsonAsBytes, err := json.Marshal(res)
	return jsonAsBytes, err
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
	//check client by hash
	_, ok := clientList[hash]
	if ok {
		//TODO check if client delete - maybe recreate ???
		return nil, errors.New("client " + hash + " already exists")
	}
	newClient := &Client{}
	newClient.makeAction("insert", status, user, insComp)
	clientList[hash] = newClient
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

	findClient, ok := clientList[hash]
	//get client by hash
	if !ok {
		return nil, errors.New("client " + hash + "not exists")
	}
	//TODO check if next state is correct (ie delete to ok)
	findClient.Status = status
	findClient.ModifyDate = time.Now()
	findClient.makeAction("update", status, user, insComp)
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

	findClient, ok := clientList[hash]
	//get client by hash
	if !ok {
		return nil, errors.New("client " + hash + "not exists")
	}
	findClient.Status = "deleted"
	findClient.ModifyDate = time.Now()
	findClient.makeAction("delete", "deleted", user, insComp)
	return nil, nil
}

//create action record for client request history
func (client *Client) makeAction(actionMethod string, status string, user string, insuranceCompany string) {
	//update status
	client.Status = status
	//update actual date
	client.ModifyDate = time.Now()
	//create action
	action := Action{insuranceCompany, user, actionMethod, time.Now(), status}
	//add to history
	client.History = append(client.History, action)
}
