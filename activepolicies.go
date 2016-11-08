package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func addActivePolicy(stub *shim.ChaincodeStub, policy Policy) error {
	//TODO(jacob): confirm that all votes are positive

	//TODO(jacob): check that the policy does not already exist in activePolicies
	// if it does, remove existing policy from activePolicies
	
	//TODO(jacob): add policy to activePolicies

	return nil
}

func modifyActivePolicy(stub *shim.ChaincodeStub, policy Policy, terms CarrierTerms) (Policy, error) {
	//TODO(jacob): check that the carrier for terms is a carrier for the policy

	//TODO(jacob): check that the terms are different from the current terms on the policy

	//TODO(jacob): check that this policy is not currently in pendingPolicies
	// if it is, clear all votes and modify pending policy with these terms
	// if it is not, then change the policy and add to pendingPolicies
	
	var resultPolicy Policy
	return resultPolicy, nil
}
