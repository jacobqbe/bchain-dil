package main

import(
	//"bytes"
	"encoding/json"
	//"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {}

type PolicyHolder struct {
	ID string `json:"id"`
	Policies []Policy `json:"policies"`
}

type CarrierTerms struct {
	CarrierID string `json:"carrier"`
	ID string `json:"timestamp"`
	Country string `json:"country"`
	Premium int64 `json:"premium"`
	Value int64 `json:"value"`
}

type Policy struct {
	ID string `json:"id"`
	HolderID string `json:"holderID"`
	Countries []string `json:"countries"`
	Terms []CarrierTerms `json:"terms"`
}

type AllPolicies struct {
	Catalog []Policy `json:"policies"`
}

var incompletePoliciesString = "_incompletePolicies"
var pendingPoliciesString = "_pendingPolicies"
var activePoliciesString = "_activePolicies"

func makeHash(args []string) string {
	if len(args) < 0 {
		return "no_hash_can_be_generated"
	}

	i := 1
	s := ""
	for i < len(args){
		s = s + args[i]
		i = i + 1
	}
	return s
}

var logger = shim.NewLogger("debug log")

func main() {
	fmt.Println("Function: main")
	
	err := shim.Start(new(SimpleChaincode))
	logger.SetLevel(shim.LogDebug)
	shim.SetLoggingLevel(shim.LogDebug)
	
	if err != nil {
		fmt.Printf("Error starting simple chaincode: %s", err)
	}
}

/*
Methods for SimpleChaincode 
*/

// Initialize the state of the 'Policies' variable
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

// Manipulate the blockchain
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("Method: SimpleChaincode.Invoke; received: " + function)

	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "generatePolicy" {
		return generatePolicy(stub, args)
	} else if function == "assignTerms" {
		return assignTerms(stub, args)
	}
	
	fmt.Println("Invoke did not find a function: " + function)
	return nil, errors.New("Received unknown function invocation")
}

// Check the state of the blockchain
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

/*
Functions for carrying out complex actions on the blockchain.
These include adding values to arrays on the blockchain and verification of permissions
*/

func generatePolicy(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Println("Function: generatePolicy")

	if len(args) < 2 {
		return nil, errors.New("Expected multiple arguments; arguments received: " +  strconv.Itoa(len(args)))
	}
	newPolicy := createPolicyObject(args)
	
	// Retrieve the current list of pending policies
	incompleteAsBytes, err := getPolicies(stub, incompletePoliciesString)
	if err != nil {
		return nil, err
	}

	var incompletePolicies AllPolicies
	incompletePolicies, err = bytesToAllPolicies(incompleteAsBytes)
	if err != nil {
		return nil, err
	}
	
	// Add the new policy to the list of pending policies
	incompletePolicies.Catalog = append(incompletePolicies.Catalog, newPolicy)
	fmt.Println("New policy appended to incomplete policies. Incomplete policy count: " + strconv.Itoa(len(incompletePolicies.Catalog)))

	incompleteAsBytes, err = json.Marshal(incompletePolicies)
	if err != nil {
		return nil, err
	}
	fmt.Println("incomplete policies successfully converted to bytes")
	
	err = write(stub, incompletePoliciesString, incompleteAsBytes)
	if err != nil {
		return nil, err
	}
	fmt.Println("incomplete policies successfully rewritten with new policy")
	return nil, nil
}

func assignTerms(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Println("Function: assignTerms")
	
	// args should include the terms of the policy and the timestamp for the policy that it will be assigned to
	if len(args) != 5 {
		return nil, errors.New("Expecting 5 arguments; arguments received: " + strconv.Itoa(len(args)))
	}

	incompleteAsBytes, err := getPolicies(stub, incompletePoliciesString)
	if err != nil {
		return nil, err
	}

	var incompletePolicies AllPolicies
	incompletePolicies, err = bytesToAllPolicies(incompleteAsBytes)
	if err != nil {
		return nil, err
	}

	if len(incompletePolicies.Catalog) == 0 {
		return nil, errors.New("No incomplete policies were found")
	}
	
	policyHash := args[0]
	var targetPolicy Policy
	var index int
	targetPolicy, index, err = getPolicyByHash(incompletePolicies.Catalog, policyHash)
	if err != nil {
		return nil, err
	}

	carrierArgs := args[1:]
	var carrierTerms CarrierTerms
	carrierTerms, err = createTerms(carrierArgs)
	if err != nil {
		return nil, err
	}

	err = insertTermsIntoPolicy(&targetPolicy, carrierTerms)
	if err != nil {
		return nil, err
	}

	err = checkComplete(targetPolicy)
	if err != nil {
		return nil, nil
	}

	// TODO(jacob): remove completed policy from incompletePolicies and insert into pendingPolicies
	err = removePolicy(&incompletePolicies, index)
	return nil, nil
}

func removePolicy(policies *AllPolicies, index int) error {
	fmt.Println("Function: removePolicy")
	return nil
}

func bytesToAllPolicies(policiesAsBytes []byte) (AllPolicies, error) {
	fmt.Println("Function: bytesToAllPolicies")
	
	var policies AllPolicies

	err := json.Unmarshal(policiesAsBytes, &policies)
	fmt.Println("json.Unmarshal error:")
	fmt.Println(err)
	
	return policies, err
}

func checkComplete(policy Policy) error {
	fmt.Println("Function: checkComplete")

	i := 1
	for i < len(policy.Terms) {
		if policy.Terms[i].ID == "" {
			return errors.New("Policy incomplete")
		}
		i = i + 1
	}
	fmt.Println("Policy complete")
	return nil
}

func getPolicyByHash(policies []Policy, hash string) (Policy, int, error) {

	var i int
	i = 0
	for i < len(policies) {
		if policies[i].ID == hash {
			return policies[i], i, nil
		}
		i = i + 1
	}
	
	var noPolicy Policy
	return noPolicy, 0, errors.New("No policy found with hash: " + hash)
}

func insertTermsIntoPolicy(policy *Policy, terms CarrierTerms) error {
	fmt.Println("Function: insertTermsIntoPolicy")
	
	i := 1
	for i < len(policy.Terms) {
		if policy.Terms[i].Country == terms.Country && policy.Terms[i].ID == "" {
			policy.Terms[i] = terms
			fmt.Println("Policy found; terms inserted")
			return nil
		}
		i = i + 1
	}
	return errors.New("Policy does not require country: " + terms.Country)
}

/*
Functions for directly reading & writing to the blockchain
*/

func write(stub *shim.ChaincodeStub, name string, value []byte) error {
	fmt.Println("Function: write")
	
	err := stub.PutState(name, value)
	if err != nil {
		return err
	}
	return nil
}

func getPolicies(stub *shim.ChaincodeStub, policiesString string) ([]byte, error) {
	fmt.Println("Function: getPolicies (" + policiesString + ")")
	
	policiesAsBytes, err := stub.GetState(policiesString)
	if err != nil {
		jsonResp := "{\"Error\": \"Failed to get policies.\"}"
		return nil, errors.New(jsonResp)
	}

	return policiesAsBytes, nil
}

/*
Functions for creating objects 
*/

func createTerms(args []string) (CarrierTerms, error) {
	fmt.Println("Function: createTerms")
	
	var terms CarrierTerms
	if len(args) != 4 {
		return terms, errors.New("Expected 4 arguments; arguments received: " + strconv.Itoa(len(args)))
	}

	var err error
	terms.CarrierID = args[0]
	terms.ID = makeHash(args)
	terms.Country = args[1] 
	terms.Premium, err = strconv.ParseInt(args[2], 10, 64)
	terms.Value, err = strconv.ParseInt(args[3], 10, 64)
	
	return terms, err
}

func createPolicyObject(args []string) Policy {
	fmt.Println("Function: createPolicyObject")
	
	var policy Policy
	holder := args[0]
	countries := args[1:]
	policy.ID = makeHash(args)
	policy.HolderID = holder
	policy.Countries = countries
	policy.Terms = make([]CarrierTerms, len(countries))

	var i = 0
	for i < len(countries){
		policy.Terms[i].Country = countries[i]
		i = i + 1
	}

	return policy
}
