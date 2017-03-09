package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	//"strconv"
	//"strings"

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
		return t.doActions(stub, args[0])
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
		return t.doActions(stub, args[0])
	}
	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) doActions(stub shim.ChaincodeStubInterface, args string) ([]byte, error) {
	var buffer bytes.Buffer
	var err error
	var response []byte
	//parse input json
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(args), &parsed)
	if err != nil {
		return nil, err
	}
	actions := parsed["actions"].([]interface{})
	for _, act := range actions {
		vv := act.(map[string]interface{})
		address := vv["address"]
		function := vv["function"]
		key := vv["key"]
		value := vv["value"]
		if function == "read" {
			response, err = t.invokeChainCode(stub, address.(string), function.(string), key.(string), "")
		} else if function == "exception" {
			err = errors.New("bad function")
		} else {
			response, err = t.invokeChainCode(stub, address.(string), function.(string), key.(string), value.(string))
		}

		if err != nil {
			errStr := fmt.Sprintf("SHOP:Failed to invoke chaincode . Got error: %s", err.Error())
			fmt.Printf(errStr)
			response = []byte(errStr)
		}
		buffer.WriteString(address.(string)[0:5])
		buffer.WriteString("_")
		buffer.WriteString(function.(string))
		buffer.WriteString("_")
		buffer.WriteString(key.(string))
		buffer.WriteString("_")
		if function == "read" {
			buffer.WriteString(value.(string))
			buffer.WriteString(":")
		}
		buffer.Write(response)
	}
	fmt.Println("SHOP:response = " + buffer.String())
	return buffer.Bytes(), err
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
	} else {
		if chaincodeURL == "local" {
			err = errors.New("unknown function")
		} else {
			queryArgs = util.ToChaincodeArgs(function, key, value)
			response, err = stub.InvokeChaincode(chaincodeURL, queryArgs)
		}
	}
	return response, err
}
