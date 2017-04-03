package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	//"strconv"
	//"strings"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
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
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	//check arguments length
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	return shim.Success(nil)
}

//Invoke - shim method
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Invoke is running this function :" + function)
	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub)
	} else if function == "write" {
		return t.doActions(stub, args[0])
	} else if function == "read" { // read data by name from state
		return t.doActions(stub, args[0])
	}
	return shim.Error("Received unknown function invocation")
}

func (t *SimpleChaincode) doActions(stub shim.ChaincodeStubInterface, args string) pb.Response {
	var buffer bytes.Buffer
	var err error
	var response []byte
	//parse input json
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(args), &parsed)
	if err != nil {
		return shim.Error(err.Error())
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
		} else if function == "exception" && address == "local" {
			err = errors.New("bad function")
		} else {
			response, err = t.invokeChainCode(stub, address.(string), function.(string), key.(string), value.(string))
		}

		if err != nil {
			errStr := fmt.Sprintf("Got error- %s", err.Error())
			fmt.Printf(errStr)
			response = []byte(errStr)
		}
		buffer.WriteString(address.(string)[0:5])
		buffer.WriteString("_")
		buffer.WriteString(function.(string))
		buffer.WriteString("_")
		buffer.WriteString(key.(string))
		buffer.WriteString("_")
		if function == "write" {
			buffer.WriteString("_")
			buffer.WriteString(value.(string))
		}
		buffer.WriteString(":")
		buffer.Write(response)
		buffer.WriteString(";")
	}
	fmt.Println("SHOP:response = " + buffer.String())
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(buffer.Bytes())
}

func (t *SimpleChaincode) invokeChainCode(stub shim.ChaincodeStubInterface, chaincodeURL string, function string, key string, value string) ([]byte, error) {
	var queryArgs [][]byte
	var response []byte
	var err error
	var bpRes pb.Response
	if function == "write" {
		if chaincodeURL == "local" {
			err = stub.PutState(key, []byte(value))
		} else {
			queryArgs = util.ToChaincodeArgs(function, key, value)
			bpRes = stub.InvokeChaincode(chaincodeURL, queryArgs, "")
			if bpRes.Status != shim.OK {
				err = fmt.Errorf("Failed to invoke chaincode. Got error: %s", bpRes.Payload)
			}
			response = bpRes.Payload
		}
	} else if function == "read" {
		if chaincodeURL == "local" {
			response, err = stub.GetState(key)
		} else {
			queryArgs = util.ToChaincodeArgs(function, key)
			bpRes = stub.InvokeChaincode(chaincodeURL, queryArgs, "")
			if bpRes.Status != shim.OK {
				err = fmt.Errorf("Failed to invoke chaincode. Got error: %s", bpRes.Payload)
			}
			response = bpRes.Payload
		}
	} else {
		if chaincodeURL == "local" {
			err = errors.New("unknown function")
		} else {
			queryArgs = util.ToChaincodeArgs(function, key, value)
			bpRes = stub.InvokeChaincode(chaincodeURL, queryArgs, "")
			if bpRes.Status != shim.OK {
				err = fmt.Errorf("Failed to invoke chaincode. Got error: %s", bpRes.Payload)
			}
			response = bpRes.Payload
		}
	}
	return response, err
}
