package main

import (
	//"encoding/json"
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
	fmt.Errorf("TEST PRINT ERROR")
	fmt.Print("TEST PRINT")
	fmt.Printf("TEST PRINF")
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
	fmt.Printf("init storage len=%d", len(clientList))
	return nil, nil

}

//SHIM - INVOKE
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)
	fmt.Printf("ivoke storage len=%d", len(clientList))
	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "insert_client" { //create a new client
		return t.insert_client(stub, args)
	}
	//	else if function == "update_client" { //update a client
	//		return t.update_client(stub, args)
	//	} else if function == "delete_client" { //delete a client
	//		return t.delete_client(stub, args)
	//	}

	//fmt.Println("invoke did not find func: " + function) //error

	return nil, errors.New("Received unknown function invocation")
}

// SHIM - QUERY
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	} else if function == "print_all_clients" {
		return t.print_all_clients(stub, args)
	}
	//	else if function == "print_client" {
	//			return t.print_client(stub, args)
	//		}

	//fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query")
}

func (t *SimpleChaincode) print_all_clients(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("current storage len %d", len(clientList))
	s := " clients="
	for clientHash, _ := range clientList {
		s += clientHash
	}
	return []byte(s), nil
}

func (t *SimpleChaincode) insert_client(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//parse parameters  - need 4
	if len(args) < 4 {
		return nil, errors.New("incorrect number of arguments. need 4")
	}
	hash := args[0]
	status := args[1]
	user := args[2]
	insComp := args[3]
	//fmt.Println("current storage len before insert=" + len(clientList))
	//get client by hash
	if _, ok := clientList[hash]; ok {
		return nil, errors.New("client " + hash + "already exists")
	}
	newClient := &Client{}
	newClient.make_action("insert", status, user, insComp)
	clientList[hash] = newClient
	//fmt.Println("current storage len after insert=" + len(clientList))
	return nil, nil
}

func (client *Client) make_action(actionMethod string, status string, user string, insuranceCompany string) {
	//update status
	client.Status = status
	//update actual date
	client.ModifyDate = time.Now()
	//create action
	action := Action{insuranceCompany, user, actionMethod, time.Now()}
	//add to history
	client.History = append(client.History, action)
}

//func print_history_of_client(client Client) {
//	//	//historyList :=make([]string,0,len(client.History))
//	for historyNum, _ := range client.History {
//		action := client.History[historyNum]
//		//historyList = append(historyList,"InsuranceCompany:"+action.InsuranceCompany + "User:"+action.User + "Method:"+action.Method +"Date:"+Date )
//		//fmt.Println("InsuranceCompany:'" + action.InsuranceCompany + "' User:" + action.User + "Method:" + action.Method + " Date:" + action.Date.String())
//	}
//}

//read by name from all state
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}
	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}
	return valAsbytes, nil
}
