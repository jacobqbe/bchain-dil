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

	//TODO(jacob): initialize votes on policy
	
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

	pendingAsBytes, err := getPolicies(stub, pendingPoliciesString)
	if err != nil {
		return nil, err
	}

	var pendingPolicies AllPolicies
	pendingPolicies, err = bytesToAllPolicies(pendingAsBytes)
	if err != nil {
		return nil, err
	}

	var policyIndex int
	policyIndex, err = getPolicyByHash(pendingPolicies.Catalog, policyID)
	if err != nil {
		return nil, err
	}
	
	//TODO(jacob): confirm that carrierID matches a carrier on the policy & that policy has not been voted on by carrier
	i := 0
	for i < len(pendingPolicies.Catalog[policyIndex].Terms) {
		if pendingPolicies.Catalog[policyIndex].Terms[i].CarrierID == carrierID {
			err = vote(&pendingPolicies.Catalog[policyIndex], i, carrierID, voteCast)
		}
		i = i + 1
	}

	err = checkActive(&pendingPolicies.Catalog[policyIndex])
	if err != nil {
		return nil, err
	}
	//TODO(jacob): check to determine that no votes remain to be cast
	// this should probably be a function, so that it can be called from other functions as well
	
	//TODO(jacob): if all votes have been cast, check that there are no negative votes
	// if all votes are positive, remove policy from pendingPolicies and add to activePolicies
	// if not all votes are positive, check that this policy exists in activePolicies
	// if it does, simply remove from pendingPolicies. Otherwise move from pendingPolicies to
	// incompletePolicies
	
	return nil, nil
}

func vote(policy *Policy, index int, carrierID string, vote string) error {
	if policy.Votes[index].Vote != "" {
		return errors.New("vote has already been cast")
	}

	policy.Votes[index].CarrierID = carrierID
	policy.Votes[index].Vote = vote
	return nil
}

func checkActive(policy *Policy) error {
	return nil
}
