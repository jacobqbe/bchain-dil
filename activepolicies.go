package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
)

func addActivePolicy(stub *shim.ChaincodeStub, policy Policy) error {
	fmt.Println("Function: addActivePolicy")

	i := 0
	for i < len(policy.Votes) {
		if policy.Votes[i].Vote != "approve" {
			return errors.New("policy is not active; contains at least one vote that is not \"approve\"")
		}
		i = i + 1
	}
	fmt.Println("policy is confirmed to be active")
	
	activePolicies, err := readPolicies(stub, activePoliciesString)
	if err != nil {
		return err
	}
	fmt.Println("active policies retrieved")
	
	i = 0
	for i < len(activePolicies.Catalog) {
		if activePolicies.Catalog[i].ID == policy.ID {
			removePolicy(&activePolicies, i)
			fmt.Println("active policy with matching ID has been found and removed")
		}
		i = i + 1
	}

	activePolicies.Catalog = append(activePolicies.Catalog, policy)
	fmt.Println("policy appended to active policies")

	err = writePolicies(stub, activePoliciesString, activePolicies)
	if err != nil {
		return err
	}
	fmt.Println("active policies written with new active policy")

	return nil
}

func modifyActivePolicy(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Println("Function: modifyActivePolicy")

	if len(args) != 5 {
		return nil, errors.New("Expected 5 arguments; arguments received: " + strconv.Itoa(len(args)))
	}

	policyID := args[0]
	carrierTerms, err := createTerms(args[1:])
	if err != nil {
		return nil, err
	}
	fmt.Println("new terms created successfully")
	
	var activePolicies AllPolicies
	activePolicies, err = readPolicies(stub, activePoliciesString)
	if err != nil {
		return nil, err
	}
	fmt.Println("active policies successfully read")

	var policyIndex int
	policyIndex, err = getPolicyByHash(activePolicies.Catalog, policyID)
	if err != nil {
		return nil, err
	}
	

	err = modifyPolicy(stub, activePolicies.Catalog[policyIndex], carrierTerms) 
	if err != nil {
		return nil, err
	}
	fmt.Println("policy modified successfully")
	
	return nil, nil
}

func modifyPolicy(stub *shim.ChaincodeStub, policy Policy, terms CarrierTerms) error {
	fmt.Println("Function: modifyPolicy")
	
	i := 0
	termsIndex := -1
	for i < len(policy.Terms) {
		if policy.Terms[i].CarrierID == terms.CarrierID && policy.Terms[i].Country == terms.Country {
			termsIndex = i
			break
		}
		i = i + 1
	}
	if termsIndex == -1 {
		return errors.New("carrier " + terms.CarrierID + " not found for policy " + policy.ID + ", country of " + terms.Country)
	}
	fmt.Println("terms to modify found")
	
	if terms.ID == policy.Terms[termsIndex].ID {
		return errors.New("terms submitted are not different than existing terms")
	}
	policy.Terms[termsIndex] = terms;
	fmt.Println("terms have been modified")

	pendingPolicies, err := readPolicies(stub, pendingPoliciesString)
	if err != nil {
		return err
	}
	fmt.Println("pending policies have been retrieved")

	i = 0
	for i < len(pendingPolicies.Catalog) {
		if pendingPolicies.Catalog[i].ID == policy.ID {
			removePolicy(&pendingPolicies, i)
			fmt.Println("policy has been found in pending policies & removed")
			break
		}
		i = i + 1
	}

	i = 0
	for i < len(policy.Votes){
		policy.Votes[i].Vote = ""
		i = i + 1
	}

	pendingPolicies.Catalog = append(pendingPolicies.Catalog, policy)
	fmt.Println("modified policy appended to pending policies")

	err = writePolicies(stub, pendingPoliciesString, pendingPolicies)
	if err != nil {
		return err
	}
	fmt.Println("pending policies successfully written with modified policy")
	
	return nil
}
