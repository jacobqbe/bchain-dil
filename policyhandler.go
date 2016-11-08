package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func getPolicies(stub *shim.ChaincodeStub, policiesString string) ([]byte, error) {
	fmt.Println("Function: getPolicies (" + policiesString + ")")
	
	policiesAsBytes, err := stub.GetState(policiesString)
	if err != nil {
		jsonResp := "{\"Error\": \"Failed to get policies.\"}"
		return nil, errors.New(jsonResp)
	}

	return policiesAsBytes, nil
}

func removePolicy(policies *AllPolicies, index int) Policy {
	fmt.Println("Function: removePolicy")
	var pendingPolicy Policy
	pendingPolicy = policies.Catalog[index]

	copy(policies.Catalog[:index], policies.Catalog[index + 1:])
	policies.Catalog = policies.Catalog[:len(policies.Catalog) - 1]
	
	return pendingPolicy
}

func bytesToAllPolicies(policiesAsBytes []byte) (AllPolicies, error) {
	fmt.Println("Function: bytesToAllPolicies")
	
	var policies AllPolicies

	err := json.Unmarshal(policiesAsBytes, &policies)
	fmt.Println("json.Unmarshal error:")
	fmt.Println(err)
	
	return policies, err
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
