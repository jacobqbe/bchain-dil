package main

import(
	"bytes"
	"encoding/json"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"
	
//	"github.com/satori/go.uuid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {}

type PolicyHolder struct {
	ID string `json:"id"`
	Policies []Policy `json:"policies"`
}

type CarrierTerms struct {
	CarrierID string `json:"carrier"`
	Timestamp int64 `json:"timestamp"`
	Country string `json:"country"`
	Premium int64 `json:"premium"`
	Value int64 `json:"value"`
}

type Policy struct {
//	ID string
	Timestamp int64 `json:"timestamp"`
	HolderID string `json:"holderID"`
	Countries []string `json:"countries"`
	Terms []CarrierTerms `json:"terms"`
}

type AllPolicies struct {
	Catalog []Policy `json:"policies"`
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

func createTerms(args []string) (CarrierTerms, error) {
	var terms CarrierTerms
	if len(args) != 4 {
		return terms, errors.New("Expected 4 arguments; arguments received: " + strconv.Itoa(len(args)))
	}

	var err error
	terms.CarrierID = args[0]
	terms.Timestamp = makeTimestamp()
	terms.Country = args[1] 
	terms.Premium, err = strconv.ParseInt(args[2], 10, 64)
	terms.Value, err = strconv.ParseInt(args[3], 10, 64)
	
	return terms, err
}

func createPolicyObject(holder string, countries []string) Policy {
	var policy Policy
	policy.Timestamp = makeTimestamp()
	policy.HolderID = holder
	policy.Countries = countries
	policy.Terms = make([]CarrierTerms, len(countries))

	var i = 1
	for i < len(countries){
		policy.Terms[i].Country = countries[i]
	}

	return policy
}

func (t *SimpleChaincode) generatePolicy(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("Expected multiple arguments; arguments received: " +  strconv.Itoa(len(args)))
	}

	holderID := args[0]

	countries := args[1:]
	newPolicy := createPolicyObject(holderID, countries)
	
	// Retrieve the current list of pending policies
	pendingAsBytes, err := t.getPendingPolicies(stub) //stub.GetState(pendingPoliciesString)
	if err != nil {
		return nil, err
	}

	var pendingPolicies AllPolicies
	buf := bytes.NewReader(pendingAsBytes)
	err = binary.Read(buf, binary.LittleEndian, &pendingPolicies)

	// Add the new policy to the list of pending policies
	pendingPolicies.Catalog = append(pendingPolicies.Catalog, newPolicy)
	fmt.Println("New policy appended to pending policies. Pending policy count: " + strconv.Itoa(len(pendingPolicies.Catalog)))

	pendingAsBytes, err = json.Marshal(pendingPolicies)
	if err != nil {
		return nil, err
	}
	
	err = t.write(stub, pendingPoliciesString, pendingAsBytes)
	if err != nil {
		return nil, err
	}
	fmt.Println("Policy successfully added to pending policies")
	return nil, nil
}

func (t *SimpleChaincode) write(stub *shim.ChaincodeStub, name string, value []byte) error {
	err := stub.PutState(name, value)
	if err != nil {
		return err
	}
	return nil
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
	pendingAsBytes, err := stub.GetState(pendingPoliciesString)
	if err != nil {
		jsonResp := "{\"Error\": \"Failed to get pending policies.\"}"
		return nil, errors.New(jsonResp)
	}

	return pendingAsBytes, nil
}
