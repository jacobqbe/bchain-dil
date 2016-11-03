package main

import(
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	
	"github.com/satori/go.uuid"
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

	// Initialize the catalogs for both pending and active policies
	pendingCatalog := make([]Policy, 0)
	activeCatalog := make([]Policy, 0)

	//Create and marshal the active policies
	var activePolicies AllPolicies
	activePolicies.Catalog = activeCatalog
	activeAsBytes, err := json.Marshal(activePolicies)
	if err != nil {
		return nil, err
	}
	
	// Create and marshal the pending policies
	var pendingPolicies AllPolicies
	pendingPolicies.Catalog = pendingCatalog
	var pendingAsBytes []byte
	pendingAsBytes, err = json.Marshal(pendingPolicies)
	if err != nil {
		return nil, err
	}
	
	// Set the state of the 'Cows' variable to blank
	err = stub.PutState(activePoliciesString, activeAsBytes)
	if err != nil {
		fmt.Println("Failed to initialize policies")
		return nil, err
	}

	err = stub.PutState(pendingPoliciesString, pendingAsBytes)
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

	if function == "getPendingPolicies" {
		return t.getPendingPolicies(stub)
	}
	fmt.Println("Query did not find a function: " + function)
	return nil, errors.New("Received unknown function query")
}

func (t *SimpleChaincode) getPendingPolicies(stub *shim.ChaincodeStub) ([]byte, error) {
	valAsBytes, err := stub.GetState(pendingPoliciesString)
	if err != nil {
		jsonResp := "{\"Error\": \"Failed to get pending policies.\"}"
		return nil, errors.New(jsonResp)
	}

	var pendingPolicies AllPolicies
	json.Unmarshal(valAsBytes, &pendingPolicies)
	numPoliciesAsBytes, erro := json.Marshal(len(pendingPolicies.Catalog))
	if erro != nil {
		return nil, errors.New("Unable to marshal pendingPolicies.Catalog size")
	}
	
	return numPoliciesAsBytes, nil
}
