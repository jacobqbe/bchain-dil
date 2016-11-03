package main

import(
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {}

type PolicyHolder struct {
	ID string
	Policies []Policy
}

type CarrierTerms struct {
	CarrierID string
	Timestamp int64
	Country string
	Premium int32
	Value int32
}

type Policy struct {
	ID string
	Timestamp int64
	HolderID string
	Countries []string
	Terms []CarrierTerms
}

type AllPolicies struct {
	Catalog []Policy
}

var pendingPoliciesString = "_pendingPolicies"
var activePoliciesString = "_activePolicies"

func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting simple chaincode: %s", err)
	}
}

// Initialize the state of the 'Policies' variable
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("Initializing Policies")

	// Encode empty array of strings into json
	var blank []string
	blankBytes, _ := json.Marshal(&blank)

	// Set the state of the 'Cows' variable to blank
	err := stub.PutState(activePoliciesString, blankBytes)
	if err != nil {
		fmt.Println("Failed to initialize policies")
		return nil, err
	}

	err = stub.PutState(pendingPoliciesString, blankBytes)
	if err != nil {
		fmt.Println("Failed to initialize pending policies")
		return nil, err
	}

	fmt.Println("Initialization complete")
	return nil, nil
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("Invoking")

	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "generatePolicy" {
		return t.generatePolicy(stub, args)
	}
	
	fmt.Println("Invoke did not find a function: " + function)
	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) createTerms(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	/*
json 
{
    "carrier_id" : string,
    "country": string,
    "premium":0.00,
    "value" : 0.00
}

*/
	return nil, nil
}

func (t *SimpleChaincode) generatePolicy(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("Expected multiple arguments; arguments received: " +  strconv.Itoa(len(args)))
	}

	// Generate a new UUID for the new policy
	policyID := uuid.NewV4().String()

	holderID := args[0]
	
	// Compile the list of countries for which the policy needs to be covered
	countries := make([]string, len(args) - 1)
	i := 1
	for i < len(args) {
		countries[i] = args[i - 1]
		i = i + 1
	}

	// Generate new policy object
	var newPolicy Policy
	newPolicy.ID = policyID
	newPolicy.Timestamp = makeTimestamp()
	newPolicy.HolderID = holderID
	newPolicy.Countries = countries
	newPolicy.Terms = nil

	// Retrieve the current list of pending policies
	pendingAsBytes, err := stub.GetState(pendingPoliciesString)
	if err != nil {
		return nil, errors.New("Failed to get pending policies")
	}

	var pendingPolicies AllPolicies
	json.Unmarshal(pendingAsBytes, &pendingPolicies)

	// Add the new policy to the list of pending policies
	pendingPolicies.Catalog = append(pendingPolicies.Catalog, newPolicy)
	fmt.Println("New policy appended to pending policies. Pending policy count: " + strconv.Itoa(len(pendingPolicies.Catalog)))

	pendingAsBytes, _ = json.Marshal(pendingPolicies)
	err = stub.PutState(pendingPoliciesString, pendingAsBytes)
	if err != nil {
		return nil, err
	}
	fmt.Println("Policy successfully added to pending policies")
	return nil, nil
}

// Check the state of the chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("Querying: " + function)

	if function == "read" {
		return t.read(stub, args)
	}
	fmt.Println("Query did not find a function: " + function)
	return nil, errors.New("Received unknown function query")
}

// Read the state of a variable
func (t *SimpleChaincode) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Expecting one argument; arguments received: " + strconv.Itoa(len(args)))
	}

	name = args[0]
	valAsBytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\": \"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}
	
	return valAsBytes, nil
}
