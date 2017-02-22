package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode1 struct {
}

//---------------------------------------------------SHIM - INIT
func (t *SimpleChaincode1) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//check arguments length
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	return nil, nil
}

//--------------------------------------------------- SHIM - QUERY
func (t *SimpleChaincode1) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// Handle different functions
	if function == "read" { // read data by name from state
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)
	return nil, errors.New("Received unknown function query")
}

//---------------------------------------------SHIM - INVOKE
func (t *SimpleChaincode1) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Invoke is running this function :" + function)
	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	}
	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode1) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	key := args[0]
	return stub.GetState(key)
}

func (t *SimpleChaincode1) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	key := args[0]
	val := []byte(args[1])
	return nil, stub.PutState(key, val)
}

func main() {
	err := shim.Start(new(SimpleChaincode1))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
