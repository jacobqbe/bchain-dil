package main

import(
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	"string"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {}

type CarrierTerms struct {
	CarrierID string
	Country string
	Premium int32
	Value int32
}

type Policy struct {
	ID string
	Terms []CarrierTerms
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil{
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
	err := stub.PutState("Policies", blankBytes)
	if err != nil {
		fmt.Println("Failed to initialize policies")
		return nil, err
	}

	fmt.Println("Initialization complete")
	return nil, nil
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("Invoking")

	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.Write(stub, args)
	}
	
	fmt.Println("Invoke did not find a function: " + function)
	return nil, errors.New("Received unknown function invocation")
}

// Check the state of the chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("Querying: " + function)

	if function == "read" {
		return t.Read(stub, args)
	}
	fmt.Println("Query did not find a function: " + function)
	return nil, error.New("Received unknown function query")
}

// Write a value to a variable
func (t *SimpleChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Println("Writing")

	var name, value string
	var err error

	if len(args) != 2 {
		return nil, errors.New("Expecting two arguments; arguments received: " + len(args))
	}

	name = args[0]
	value = args[1]
	err = stub.Putstate(name, []byte(value))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Read the state of a variable
func (t *SimpleChaincode) Read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments; arguments received: " + len(args))
	}

	name = args[0]
	valAsBytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\": \"Failed to get state for " + name + "\"}"
		return nil, errors.new(jsonResp)
	}
	
	return valAsBytes, nil
}
