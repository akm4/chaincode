package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("examples")

//SimpleChaincode - default "class"
type SimpleChaincode struct {
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
	if function == "write" {
		return t.write(stub, args)
	} else if function == "read" { // read data by name from state
		return t.read(stub, args)
	}
	return shim.Error("Received unknown function invocation")
}

func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	key := args[0]
	Avalbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}
	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil value for " + key + "\"}"
		return shim.Error(jsonResp)
	}
	jsonResp := "{\"Key\":\"" + key + "\",\"Value\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	key := args[0]
	val := []byte(args[1])
	err := stub.PutState(key, val)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func main() {
	logger.SetLevel(shim.LogInfo)

	logLevel, _ := shim.LogLevel(os.Getenv("SHIM_LOGGING_LEVEL"))
	shim.SetLoggingLevel(logLevel)
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
