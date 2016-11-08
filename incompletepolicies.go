package main

import(
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
)

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

	//TODO: check that policy holder has been registered
	
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

	err = writePolicies(stub, incompletePoliciesString, incompletePolicies)
	if err != nil {
		return nil, err
	}
	fmt.Println("incomplete policies successfully rewritten with new policy")
	return nil, nil
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

	pendingPolicy := removePolicy(&incompletePolicies, index)
		
	err = addPendingPolicy(stub, pendingPolicy)
	if err != nil {
		return nil, err
	}
	fmt.Println("policy successfully added to pending policies")
	fmt.Println("pendingPolicy removed from incomplete policies")

	err = writePolicies(stub, incompletePoliciesString, incompletePolicies)
	if err != nil {
		return nil, err
	}
	fmt.Println("incomplete policies successfully written")
	return nil, nil
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
