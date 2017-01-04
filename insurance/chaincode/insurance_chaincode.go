package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"time"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var clientsIndexStr = "_clientsIndex"				//name for the key/value that will store a list of all known clients-hashes
var contractsIndexStr = "_contractsIndex"				//name for the key/value that will store all contracts

type Client struct{
	ClientHash string `json:"clientHash"`					//the fieldtags are needed to keep case from bouncing around
	Status string `json:"status"`								// fraud / suspicious / ok
	ContractHistroy []Contract `json:contractHistory`
}

type Contract struct{
	ContractHash string `json:"contractHash"`
	ObjectHash string `json:"objectHash"`
	ModifyDate int64 `json:"modifyDate"`			// datetime of contract modification
	Timestamp int64 `json:"timestamp"`			//utc timestamp of creation
}

type AllContracts struct{
	AllContracts []Contract `json:"contracts"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Insurance fraud DB chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}

	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(clientsIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	var contracts AllContracts
	jsonAsBytes, _ = json.Marshal(contracts)								//clear the contracts index
	err = stub.PutState(contractsIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "init_client" {									//create a new client
		return t.init_client(stub, args)
	}
	/*else if function == "delete" {										//deletes an entity from its state
		res, err := t.Delete(stub, args)
		cleanTrades(stub)													//lets make sure all open trades are still valid
		return res, err
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "set_user" {										//change owner of a client
		res, err := t.set_user(stub, args)
		cleanTrades(stub)													//lets make sure all open trades are still valid
		return res, err
	} else if function == "open_trade" {									//create a new trade order
		return t.open_trade(stub, args)
	} else if function == "perform_trade" {									//forfill an open trade order
		res, err := t.perform_trade(stub, args)
		cleanTrades(stub)													//lets clean just in case
		return res, err
	} else if function == "remove_trade" {									//cancel an open trade order
		return t.remove_trade(stub, args)
	}*/
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

// ============================================================================================================================
// Delete - remove a key/value pair from state
// ============================================================================================================================
func (t *SimpleChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	name := args[0]
	err := stub.DelState(name)													//remove the key from chaincode state
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	//get the client index
	clientsAsBytes, err := stub.GetState(clientsIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get client index")
	}
	var clientIndex []string
	json.Unmarshal(clientsAsBytes, &clientIndex)								//un stringify it aka JSON.parse()

	//remove client from index
	for i,val := range clientIndex{
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + name)
		if val == name{															//find the correct client
			fmt.Println("found client")
			clientIndex = append(clientIndex[:i], clientIndex[i+1:]...)			//remove it
			for x:= range clientIndex{											//debug prints...
				fmt.Println(string(x) + " - " + clientIndex[x])
			}
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(clientIndex)									//save new index
	err = stub.PutState(clientsIndexStr, jsonAsBytes)
	return nil, nil
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init Client - create a new client, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_client(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0       1       2     3
	// "01069306DB19E50B5A4CEB44EA4BFD43", "fraud", "contract history", "modify date"
	if len(args) != 2 {		// TODO add last 2 args
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	fmt.Println("- start init client")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}/*
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}*/
	clientHash := strings.ToLower(args[0])
	status := strings.ToLower(args[1])
	/*user := strings.ToLower(args[3])
	size, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}*/

	//check if client already exists
	clientAsBytes, err := stub.GetState(clientHash)
	if err != nil {
		return nil, errors.New("Failed to get clientHash")
	}
	res := Client{}
	json.Unmarshal(clientAsBytes, &res)
	if res.ClientHash == clientHash{
		fmt.Println("This client arleady exists: " + clientHash)
		fmt.Println(res);
		return nil, errors.New("This client arleady exists")				//all stop. a client by this hash exists
	}

	//build the client json string manually
	str := `{"clientHash": "` + clientHash + `", "status": "` + status + `"}`
	err = stub.PutState(clientHash, []byte(str))									//store client with id as key
	if err != nil {
		return nil, err
	}

	//get the client index
	clientsAsBytes, err := stub.GetState(clientsIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get client index")
	}
	var clientIndex []string
	json.Unmarshal(clientsAsBytes, &clientIndex)							//un stringify it aka JSON.parse()

	//append
	clientIndex = append(clientIndex, clientHash)									//add client name to index list
	fmt.Println("! client index: ", clientIndex)
	jsonAsBytes, _ := json.Marshal(clientIndex)
	err = stub.PutState(clientsIndexStr, jsonAsBytes)						//store name of client

	fmt.Println("- end init client")
	return nil, nil
}

// ============================================================================================================================
// Set User Permission on Client
// ============================================================================================================================
func (t *SimpleChaincode) set_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	fmt.Println("- start set user")
	fmt.Println(args[0] + " - " + args[1])
	clientAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := Client{}
	json.Unmarshal(clientAsBytes, &res)										//un stringify it aka JSON.parse()
//	res.User = args[1]														//change the user

	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the client with id as key
	if err != nil {
		return nil, err
	}

	fmt.Println("- end set user")
	return nil, nil
}

// ============================================================================================================================
// Open Trade - create an open trade for a client you want with clients you have
// ============================================================================================================================
/*func (t *SimpleChaincode) open_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	var will_size int
	var trade_away Description

	//	0        1      2     3      4      5       6
	//["bob", "blue", "16", "red", "16"] *"blue", "35*
	if len(args) < 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting like 5?")
	}
	if len(args)%2 == 0{
		return nil, errors.New("Incorrect number of arguments. Expecting an odd number")
	}

	size1, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}

	open := AnOpenTrade{}
	open.User = args[0]
	open.Timestamp = makeTimestamp()											//use timestamp as an ID
	open.Want.Color = args[1]
	open.Want.Size =  size1
	fmt.Println("- start open trade")
	jsonAsBytes, _ := json.Marshal(open)
	err = stub.PutState("_debug1", jsonAsBytes)

	for i:=3; i < len(args); i++ {												//create and append each willing trade
		will_size, err = strconv.Atoi(args[i + 1])
		if err != nil {
			msg := "is not a numeric string " + args[i + 1]
			fmt.Println(msg)
			return nil, errors.New(msg)
		}

		trade_away = Description{}
		trade_away.Color = args[i]
		trade_away.Size =  will_size
		fmt.Println("! created trade_away: " + args[i])
		jsonAsBytes, _ = json.Marshal(trade_away)
		err = stub.PutState("_debug2", jsonAsBytes)

		open.Willing = append(open.Willing, trade_away)
		fmt.Println("! appended willing to open")
		i++;
	}

	//get the open trade struct
	tradesAsBytes, err := stub.GetState(contractsIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)										//un stringify it aka JSON.parse()

	trades.OpenTrades = append(trades.OpenTrades, open);						//append to open trades
	fmt.Println("! appended open to trades")
	jsonAsBytes, _ = json.Marshal(trades)
	err = stub.PutState(contractsIndexStr, jsonAsBytes)								//rewrite open orders
	if err != nil {
		return nil, err
	}
	fmt.Println("- end open trade")
	return nil, nil
}
*/
// ============================================================================================================================
// Perform Trade - close an open trade and move ownership
// ============================================================================================================================

/*
func (t *SimpleChaincode) perform_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//	0		1					2					3				4					5
	//[data.id, data.closer.user, data.closer.name, data.opener.user, data.opener.color, data.opener.size]
	if len(args) < 6 {
		return nil, errors.New("Incorrect number of arguments. Expecting 6")
	}

	fmt.Println("- start close trade")
	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}

	size, err := strconv.Atoi(args[5])
	if err != nil {
		return nil, errors.New("6th argument must be a numeric string")
	}

	//get the open trade struct
	tradesAsBytes, err := stub.GetState(contractsIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)															//un stringify it aka JSON.parse()

	for i := range trades.OpenTrades{																//look for the trade
		fmt.Println("looking at " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
		if trades.OpenTrades[i].Timestamp == timestamp{
			fmt.Println("found the trade");


			clientAsBytes, err := stub.GetState(args[2])
			if err != nil {
				return nil, errors.New("Failed to get thing")
			}
			closersClient := Client{}
			json.Unmarshal(clientAsBytes, &closersClient)											//un stringify it aka JSON.parse()

			//verify if client meets trade requirements
			if closersClient.Color != trades.OpenTrades[i].Want.Color || closersClient.Size != trades.OpenTrades[i].Want.Size {
				msg := "client in input does not meet trade requriements"
				fmt.Println(msg)
				return nil, errors.New(msg)
			}

			client, e := findClient4Trade(stub, trades.OpenTrades[i].User, args[4], size)			//find a client that is suitable from opener
			if(e == nil){
				fmt.Println("! no errors, proceeding")

				t.set_user(stub, []string{args[2], trades.OpenTrades[i].User})						//change owner of selected client, closer -> opener
				t.set_user(stub, []string{client.Name, args[1]})									//change owner of selected client, opener -> closer

				trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)		//remove trade
				jsonAsBytes, _ := json.Marshal(trades)
				err = stub.PutState(contractsIndexStr, jsonAsBytes)										//rewrite open orders
				if err != nil {
					return nil, err
				}
			}
		}
	}
	fmt.Println("- end close trade")
	return nil, nil
}
*/
// ============================================================================================================================
// findClient4Trade - look for a matching client that this user owns and return it
// ============================================================================================================================
/*
func findClient4Trade(stub shim.ChaincodeStubInterface, user string, color string, size int )(m Client, err error){
	var fail Client;
	fmt.Println("- start find client 4 trade")
	fmt.Println("looking for " + user + ", " + color + ", " + strconv.Itoa(size));

	//get the client index
	clientsAsBytes, err := stub.GetState(clientsIndexStr)
	if err != nil {
		return fail, errors.New("Failed to get client index")
	}
	var clientIndex []string
	json.Unmarshal(clientsAsBytes, &clientIndex)								//un stringify it aka JSON.parse()

	for i:= range clientIndex{													//iter through all the clients
		//fmt.Println("looking @ client name: " + clientIndex[i]);

		clientAsBytes, err := stub.GetState(clientIndex[i])						//grab this client
		if err != nil {
			return fail, errors.New("Failed to get client")
		}
		res := Client{}
		json.Unmarshal(clientAsBytes, &res)										//un stringify it aka JSON.parse()
		//fmt.Println("looking @ " + res.User + ", " + res.Color + ", " + strconv.Itoa(res.Size));

		//check for user && color && size
		if strings.ToLower(res.User) == strings.ToLower(user) && strings.ToLower(res.Color) == strings.ToLower(color) && res.Size == size{
			fmt.Println("found a client: " + res.Name)
			fmt.Println("! end find client 4 trade")
			return res, nil
		}
	}

	fmt.Println("- end find client 4 trade - error")
	return fail, errors.New("Did not find client to use in this trade")
}
*/
// ============================================================================================================================
// Make Timestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

// ============================================================================================================================
// Remove Open Trade - close an open trade
// ============================================================================================================================
/*
func (t *SimpleChaincode) remove_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//	0
	//[data.id]
	if len(args) < 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	fmt.Println("- start remove trade")
	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}

	//get the open trade struct
	tradesAsBytes, err := stub.GetState(contractsIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)																//un stringify it aka JSON.parse()

	for i := range trades.OpenTrades{																	//look for the trade
		//fmt.Println("looking at " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
		if trades.OpenTrades[i].Timestamp == timestamp{
			fmt.Println("found the trade");
			trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)				//remove this trade
			jsonAsBytes, _ := json.Marshal(trades)
			err = stub.PutState(contractsIndexStr, jsonAsBytes)												//rewrite open orders
			if err != nil {
				return nil, err
			}
			break
		}
	}

	fmt.Println("- end remove trade")
	return nil, nil
}
*/
// ============================================================================================================================
// Clean Up Open Trades - make sure open trades are still possible, remove choices that are no longer possible, remove trades that have no valid choices
// ============================================================================================================================
/*
func cleanTrades(stub shim.ChaincodeStubInterface)(err error){
	var didWork = false
	fmt.Println("- start clean trades")

	//get the open trade struct
	tradesAsBytes, err := stub.GetState(contractsIndexStr)
	if err != nil {
		return errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)																		//un stringify it aka JSON.parse()

	fmt.Println("# trades " + strconv.Itoa(len(trades.OpenTrades)))
	for i:=0; i<len(trades.OpenTrades); {																		//iter over all the known open trades
		fmt.Println(strconv.Itoa(i) + ": looking at trade " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10))

		fmt.Println("# options " + strconv.Itoa(len(trades.OpenTrades[i].Willing)))
		for x:=0; x<len(trades.OpenTrades[i].Willing); {														//find a client that is suitable
			fmt.Println("! on next option " + strconv.Itoa(i) + ":" + strconv.Itoa(x))
			_, e := findClient4Trade(stub, trades.OpenTrades[i].User, trades.OpenTrades[i].Willing[x].Color, trades.OpenTrades[i].Willing[x].Size)
			if(e != nil){
				fmt.Println("! errors with this option, removing option")
				didWork = true
				trades.OpenTrades[i].Willing = append(trades.OpenTrades[i].Willing[:x], trades.OpenTrades[i].Willing[x+1:]...)	//remove this option
				x--;
			}else{
				fmt.Println("! this option is fine")
			}

			x++
			fmt.Println("! x:" + strconv.Itoa(x))
			if x >= len(trades.OpenTrades[i].Willing) {														//things might have shifted, recalcuate
				break
			}
		}

		if len(trades.OpenTrades[i].Willing) == 0 {
			fmt.Println("! no more options for this trade, removing trade")
			didWork = true
			trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)					//remove this trade
			i--;
		}

		i++
		fmt.Println("! i:" + strconv.Itoa(i))
		if i >= len(trades.OpenTrades) {																	//things might have shifted, recalcuate
			break
		}
	}

	if(didWork){
		fmt.Println("! saving open trade changes")
		jsonAsBytes, _ := json.Marshal(trades)
		err = stub.PutState(contractsIndexStr, jsonAsBytes)														//rewrite open orders
		if err != nil {
			return err
		}
	}else{
		fmt.Println("! all open trades are fine")
	}

	fmt.Println("- end clean trades")
	return nil
}
*/
