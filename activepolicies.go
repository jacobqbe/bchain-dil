package main

import (
	"encoding/json"
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

	err = addPolicyToHolder(stub, policy, policy.HolderID)
	err = writePolicies(stub, activePoliciesString, activePolicies)
	if err != nil {
		return err
	}
	fmt.Println("active policies written with new active policy")

	return nil
}

func addPolicyToHolder(stub *shim.ChaincodeStub, policy Policy, holderID string) error {
	fmt.Println("Function: addPolicyToHolder")

	holdersAsBytes, err := stub.GetState(holdersString)
	if err != nil {
		return err
	}
	fmt.Println("holders retrieved as bytes")

	var holders AllHolders
	err = json.Unmarshal(holdersAsBytes, &holders)
	if err != nil {
		return err
	}
	fmt.Println("holders retrieved from bytes")

	i := 0
	var policyHolder PolicyHolder
	for i < len(holders.Catalog) {
		if holders.Catalog[i].ID == holderID {
			policyHolder = holders.Catalog[i]
			fmt.Println("policy holder found")
			break
		}
		i = i + 1
	}
	if policyHolder.ID == "" {
		policyHolder.ID = holderID
		tempPolicies := make([]Policy, 0)
		policyHolder.Policies = tempPolicies
		fmt.Println("new policy holder added")
	}

	i = 0
	for i < len(policyHolder.Policies) {
		if policyHolder.Policies[i].ID == policy.ID {
			copy(policyHolder.Policies[:i], policyHolder.Policies[i + 1:])
			policyHolder.Policies = policyHolder.Policies[:len(policyHolder.Policies) - 1]
			fmt.Println("matching policy removed from policy holder")
		}
		i = i + 1
	}

	policyHolder.Policies = append(policyHolder.Policies, policy)
	fmt.Println("policy added to policy holder")

	holdersAsBytes, err = json.Marshal(holders)
	if err != nil {
		return err
	}
	fmt.Println("holders written to bytes")

	err = write(stub, holdersString, holdersAsBytes)
	if err != nil {
		return err
	}
	fmt.Println("holders written to blockchain with new policy")
	
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
