package main

import (
	"fmt"
	//"os"
	"strconv"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	time "time"
)

var logger = shim.NewLogger("examples")

//SimpleChaincode - default "class"
type SimpleChaincode struct {
}

//Init - shim method
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("init method")
	//check arguments length
	function, args := stub.GetFunctionAndParameters()
	logger.Debugf("dInit is running this function :%s", function)
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	return shim.Success(nil)
}

//Invoke - shim method
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Invoke is running this function :" + function)
	logger.Debugf("dInvoke is running this function :%s", function)
	// Handle different functions
	if function == "write" {
		return t.write(stub, args)
	} else if function == "read" { // read data by name from state
		return t.read(stub, args)
	} else if function == "multipleWrite" {
		return t.multipleWrite(stub, args)
	} else if function == "readKeyHistory" {
		return t.readKeyHistory(stub, args)
	}
	return shim.Error("Received unknown function invocation")
}
func (t *SimpleChaincode) readKeyHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	var history []string
	key := args[0]
	resultsIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		historicValue, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if historicValue!=nil && historicValue.Value!=nil{
			history = append(history, time.Unix(historicValue.Timestamp.Seconds, int64(historicValue.Timestamp.Nanos)).String()," - ",string(historicValue.Value),";") //add this tx to the list
		}

	}
	fmt.Printf("- readKeyHistory returning:\n%s", history)

	//change to array of bytes
	historyAsBytes, _ := json.MarshalIndent(history,"", "\t") //convert to array of bytes
	return shim.Success(historyAsBytes)
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

func (t *SimpleChaincode) multipleWrite(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	key := args[0]
	count, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("incorrect count " + err.Error())
	}
	for i := 0; i < count; i++ {
		err := stub.PutState(key+strconv.Itoa(i), []byte(strconv.Itoa(i)))
		if err != nil {
			return shim.Error(err.Error())
		}

	}
	return shim.Success(nil)
}

func main() {
	fmt.Println("main method")
	logger.SetLevel(shim.LogDebug)
	logLevel, _ := shim.LogLevel("DEBUG")
	shim.SetLoggingLevel(logLevel)
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
