package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type Policy struct {
	ID string `json:"id"`
	HolderID string `json:"holderID"`
	Countries []string `json:"countries"`
	Terms []CarrierTerms `json:"terms"`
}

type AllPolicies struct {
	Catalog []Policy `json:"policies"`
}

type PolicyHolder struct {
	ID string `json:"id"`
	Policies []Policy `json:"policies"`
}

func createPolicyObject(args []string) Policy {
	fmt.Println("Function: createPolicyObject")
	
	var policy Policy
	policy.ID = makeHash(args)
	policy.HolderID = args[0]

	countries := args[1:]
	policy.Countries = countries
	policy.Terms = make([]CarrierTerms, len(countries))

	var i = 0
	for i < len(countries){
		policy.Terms[i].Country = countries[i]
		i = i + 1
	}

	return policy
}

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


func removePolicy(policies *AllPolicies, index int) Policy {
	fmt.Println("Function: removePolicy")
	var pendingPolicy Policy
	pendingPolicy = policies.Catalog[index]

	copy(policies.Catalog[:index], policies.Catalog[index + 1:])
	policies.Catalog = policies.Catalog[:len(policies.Catalog) - 1]
	
	return pendingPolicy
}

func addPendingPolicy(stub *shim.ChaincodeStub, policy Policy) error {
	fmt.Println("Function: addPendingPolicy")
	
	pendingAsBytes, err := getPolicies(stub, pendingPoliciesString)
	if err != nil {
		return err
	}
	fmt.Println("pending policies retrieved")

	var pendingPolicies AllPolicies
	pendingPolicies, err = bytesToAllPolicies(pendingAsBytes)
	if err != nil {
		return err
	}
	fmt.Println("pending policies derived from bytes")

	pendingPolicies.Catalog = append(pendingPolicies.Catalog, policy)
	fmt.Println("policy appended to pending policies")
	
	pendingAsBytes, err = json.Marshal(pendingPolicies)
	err = write(stub, pendingPoliciesString, pendingAsBytes)
	if err != nil {
		return err
	}
	fmt.Println("pending policies successfully rewritten with complete policy")
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

	i := 0
	for i < len(policy.Terms) {
		if policy.Terms[i].ID == "" {
			return errors.New("Policy incomplete")
		}
		i = i + 1
	}
	fmt.Println("Policy complete")
	return nil
}

func getPolicyByHash(policies []Policy, hash string) (int, error) {
	fmt.Println("Function: getPolicyByHash")
	
	var i int
	i = 0
	for i < len(policies) {
		if policies[i].ID == hash {
			return i, nil
		}
		i = i + 1
	}
	
	return 0, errors.New("No policy found with hash: " + hash)
}

func insertTermsIntoPolicy(policy *Policy, terms CarrierTerms) error {
	fmt.Println("Function: insertTermsIntoPolicy")
	fmt.Println("country: " + terms.Country)
	i := 0
	for i < len(policy.Terms) {
		fmt.Println("policy.Terms.Country: " + policy.Terms[i].Country)
		if policy.Terms[i].Country == terms.Country {
			fmt.Println(policy.Terms[i].ID) //
			if policy.Terms[i].ID == "" {
				policy.Terms[i] = terms
				fmt.Println("Policy found; terms inserted")
				fmt.Println("Carrier: " + policy.Terms[i].CarrierID);
				return nil
			}
		}
		i = i + 1
	}
	return errors.New("Policy does not require country: " + terms.Country)
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
