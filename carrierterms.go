package main

import(
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type CarrierTerms struct {
	CarrierID string `json:"carrier"`
	ID string `json:"id"`
	Country string `json:"country"`
	Premium int64 `json:"premium"`
	Value int64 `json:"value"`
}

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
	//var targetPolicy Policy
	var index int
	index, err = getPolicyByHash(incompletePolicies.Catalog, policyHash)
	if err != nil {
		return nil, err
	}

	carrierArgs := args[1:]
	var carrierTerms CarrierTerms
	carrierTerms, err = createTerms(carrierArgs)
	if err != nil {
		return nil, err
	}

	err = insertTermsIntoPolicy(&incompletePolicies.Catalog[index], carrierTerms)
	if err != nil {
		return nil, err
	}

	err = checkComplete(incompletePolicies.Catalog[index])
	if err != nil {
		incompleteAsBytes, err = json.Marshal(incompletePolicies)
		if err != nil {
			return nil, err
		}
		err = write(stub, incompletePoliciesString, incompleteAsBytes)
		if err != nil {
			return nil, err
		}
		fmt.Println("incomplete policies successfully written with new terms")
		return nil, nil
	}

	// TODO(jacob): remove completed policy from incompletePolicies and insert into pendingPolicies
	pendingPolicy := removePolicy(&incompletePolicies, index)
	fmt.Println("pendingPolicy removed from incomplete policies")

	incompleteAsBytes, err = json.Marshal(incompletePolicies)
	if err != nil {
		return nil, err
	}
	fmt.Println("incomplete policies converted to bytes")

	err = write(stub, incompletePoliciesString, incompleteAsBytes)
	if err != nil {
		return nil, err
	}
	fmt.Println("incomplete policies successfully written")
	
	err = addPendingPolicy(stub, pendingPolicy)
	if err != nil {
		return nil, err
	}
	fmt.Println("policy successfully added to pending policies")
	return nil, nil
}
