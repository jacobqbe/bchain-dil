package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {}

var incompletePoliciesString = "_incompletePolicies"
var pendingPoliciesString = "_pendingPolicies"
var activePoliciesString = "_activePolicies"

func main() {
	fmt.Println("Function: main")
	
	err := shim.Start(new(SimpleChaincode))	
	if err != nil {
		fmt.Printf("Error starting simple chaincode: %s", err)
	}
}

func makeHash(args []string) string {
	if len(args) < 0 {
		return "no_hash_can_be_generated"
	}

	i := 0
	s := ""
	for i < len(args){
		s = s + args[i]
		i = i + 1
	}
	return s
}

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("Method: SimpleChaincode.Init")

	// Initialize the catalogs for both pending and active policies
	incompleteCatalog := make([]Policy, 0)
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

	// Create and marshal incomplete policies
	var incompletePolicies AllPolicies
	incompletePolicies.Catalog = incompleteCatalog
	var incompleteAsBytes []byte
	incompleteAsBytes, err = json.Marshal(incompletePolicies)
	if err != nil {
		return nil, err
	}

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

	err = stub.PutState(incompletePoliciesString, incompleteAsBytes)
	if err != nil {
		fmt.Println("Failed to initialize incomplete policies")
		return nil, err
	}

	fmt.Println("Initialization complete")
	return nil, nil
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("Method: SimpleChaincode.Invoke; received: " + function)

	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "generatePolicy" {
		return generatePolicy(stub, args)
	} else if function == "assignTerms" {
		return assignTerms(stub, args)
	} else if function == "castVote" {
		return castVote(stub, args)
	}
	
	fmt.Println("Invoke did not find a function: " + function)
	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("Method: SimpleChaincode.Query; received: " + function)

	if function == "getPendingPolicies" {
		return getPolicies(stub, pendingPoliciesString)
	} else if function == "getIncompletePolicies" {
		return getPolicies(stub, incompletePoliciesString)
	} else if function == "getActivePolicies" {
		return getPolicies(stub, activePoliciesString)
	}

	fmt.Println("Query did not find a function: " + function)
	return nil, errors.New("Received unknown function query")
}

func write(stub *shim.ChaincodeStub, name string, value []byte) error {
	fmt.Println("Function: write")
	
	err := stub.PutState(name, value)
	if err != nil {
		return err
	}
	return nil
}
