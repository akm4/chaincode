package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/util"
)

//SimpleChaincode - default "class"
type SimpleChaincode struct {
}

//---------------------------------------------------- MAIN
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

//Init - shim method
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//check arguments length
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	return nil, nil
}

//Query - - shim method
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// Handle different functions
	if function == "read" { // read data by name from state
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)
	return nil, errors.New("Received unknown function query")
}

//Invoke - shim method
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Invoke is running this function :" + function)
	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	}
	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2 or more")
	}
	var value string
	var queryArgs [][]byte
	var response []byte
	var err error

	chaincodeURL := args[0]
	f := args[1]
	key := args[2]
	if f == "write" {
		value = args[3]
		if chaincodeURL == "local" {
			//RESTRICTED
			err = stub.PutState(key, []byte(value))
		} else {
			queryArgs = util.ToChaincodeArgs(f, key, value)
			response, err = stub.InvokeChaincode(chaincodeURL, queryArgs)
		}
	} else if f == "read" {
		if chaincodeURL == "local" {
			response, err = stub.GetState(key)
		} else {
			// RESTRICTED
			queryArgs = util.ToChaincodeArgs(f, key)
			response, err = stub.QueryChaincode(chaincodeURL, queryArgs)
		}
	}
	if err != nil {
		errStr := fmt.Sprintf("SHOP:Failed to query chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}
	return response, nil
}

func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2 or more")
	}
	var value string
	var queryArgs [][]byte
	var response []byte
	var err error

	chaincodeURL := args[0]
	f := args[1]
	key := args[2]
	if f == "write" {
		value = args[3]
		if chaincodeURL == "local" {
			err = stub.PutState(key, []byte(value))
		} else {
			queryArgs = util.ToChaincodeArgs(f, key, value)
			response, err = stub.InvokeChaincode(chaincodeURL, queryArgs)
		}
	} else if f == "read" {
		if chaincodeURL == "local" {
			response, err = stub.GetState(key)
		} else {
			queryArgs = util.ToChaincodeArgs(f, key)
			response, err = stub.QueryChaincode(chaincodeURL, queryArgs)
		}
	}
	if err != nil {
		errStr := fmt.Sprintf("SHOP:Failed to invoke chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}
	fmt.Println("SHOP:response = " + string(response))
	return response, err
}

func (t *SimpleChaincode) invokeChainCode(stub shim.ChaincodeStubInterface, chaincodeURL string, function string, key string, value string) ([]byte, error) {
	var queryArgs [][]byte
	var response []byte
	var err error
	if function == "write" {
		if chaincodeURL == "local" {
			err = stub.PutState(key, []byte(value))
		} else {
			queryArgs = util.ToChaincodeArgs(function, key, value)
			response, err = stub.InvokeChaincode(chaincodeURL, queryArgs)
		}
	} else if function == "read" {
		if chaincodeURL == "local" {
			response, err = stub.GetState(key)
		} else {
			queryArgs = util.ToChaincodeArgs(function, key)
			response, err = stub.QueryChaincode(chaincodeURL, queryArgs)
		}
	}
	return response, err
}
