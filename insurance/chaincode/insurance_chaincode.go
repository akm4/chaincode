package main

import (
    //"encoding/json"
    "errors"
    "fmt"
    "strconv"
    //"strings"
    //"time"

    "github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// MAIN
func main() {
    err := shim.Start(new(SimpleChaincode))
    if err != nil {
        fmt.Printf("Error starting Simple chaincode: %s", err)
    }
}

//SHIM - INIT
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	var Aval int 
 	var err error 

    if len(args) != 1 {
        return nil, errors.New("Incorrect number of arguments. Expecting 1")
    }
    Aval, err = strconv.Atoi(args[0])
    // Write the state to the ledger
    err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))
    if err != nil {
        return nil, err
    }
    return nil, nil
}

//SHIM - INVOKE
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
    fmt.Println("invoke is running " + function)

    // Handle different functions
    if function == "init" { //initialize the chaincode state, used as reset
        return t.Init(stub, "init", args)
    } else if function == "init_client" { //create a new client
        return nil,nil
		//t.init_client(stub, args)
    }
    fmt.Println("invoke did not find func: " + function) //error

    return nil, errors.New("Received unknown function invocation")
}

// SHIM - QUERY
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
    fmt.Println("query is running " + function)

    // Handle different functions
    if function == "read" { //read a variable
        return t.read(stub, args)
    }
    fmt.Println("query did not find func: " + function) //error

    return nil, errors.New("Received unknown function query")
}

//read by name
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

