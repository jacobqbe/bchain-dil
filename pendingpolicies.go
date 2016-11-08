package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
)

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

	votes := make([]Approval, len(policy.Terms))
	policy.Votes = votes
	
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

func castVote(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Println("Function: castVote")

	if len(args) != 3 {
		return nil, errors.New("Expected three arguments; arguments received: " + strconv.Itoa(len(args)))
	}
	
	policyID := args[0]
	carrierID := args[1]
	voteCast := args[2]

	if voteCast != "approve" && voteCast != "disapprove" {
		return nil, errors.New("Invalid vote: " + voteCast)
	}
	
	pendingPolicies, err := readPolicies(stub, pendingPoliciesString)
	if err != nil {
		return nil, err
	}
	
	var policyIndex int
	policyIndex, err = getPolicyByHash(pendingPolicies.Catalog, policyID)
	if err != nil {
		return nil, err
	}
	
	i := 0
	for i < len(pendingPolicies.Catalog[policyIndex].Terms) {
		if pendingPolicies.Catalog[policyIndex].Terms[i].CarrierID == carrierID {
			err = vote(&pendingPolicies.Catalog[policyIndex], i, carrierID, voteCast)
		}
		i = i + 1
	}

	err = checkActive(&pendingPolicies.Catalog[policyIndex])
	if err != nil {
		err = writePolicies(stub, pendingPoliciesString, pendingPolicies)
		if err != nil {
			return nil, err
		}
		fmt.Println("pending policies successfully written with new vote(s)")
		return nil, nil
	}

	i = 0
	for i < len(pendingPolicies.Catalog[policyIndex].Votes) {
		if pendingPolicies.Catalog[policyIndex].Votes[i].Vote != "approve" {
			incompletePolicy := removePolicy(&pendingPolicies, policyIndex)

			policyArgs := make([]string, len(incompletePolicy.Countries) + 1)
			policyArgs[0] = incompletePolicy.HolderID
			j := 0
			for j < len(incompletePolicy.Countries) {
				policyArgs[j + 1] = incompletePolicy.Countries[j]
				j = j + 1
			}
		
			_, err = generatePolicy(stub, policyArgs)
			if err != nil {
				return nil, err
			}
			err = writePolicies(stub, pendingPoliciesString, pendingPolicies)
			if err != nil {
				return nil, err
			}
			fmt.Println("pending policies successfully written after removal of incomplete policy")
			return nil, nil
		}
		i = i + 1
	}
	activePolicy := removePolicy(&pendingPolicies, policyIndex)
	err = addActivePolicy(stub, activePolicy)
	if err != nil {
		return nil, err
	}
	err = writePolicies(stub, pendingPoliciesString, pendingPolicies)
	if err != nil {
		return nil, err
	}
	
	fmt.Println("pending policies successfully written after moving policy to active policies")	
	return nil, nil
}

func vote(policy *Policy, index int, carrierID string, vote string) error {
	fmt.Println("Function: vote")
	
	if policy.Votes[index].Vote != "" {
		return errors.New("vote has already been cast")
	}

	policy.Votes[index].CarrierID = carrierID
	policy.Votes[index].Vote = vote
	return nil
}

func checkActive(policy *Policy) error {
	fmt.Println("Function: checkActive")
	
	i := 0
	for i < len(policy.Votes) {
		if policy.Votes[i].CarrierID == "" {
			return errors.New("Not all votes have been cast")
		}
		i = i + 1
	}	
	return nil
}
