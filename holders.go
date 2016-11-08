package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func registerNewHolder(stub *shim.ChaincodeStub, args []string) error {
	//TODO(jacob): Check that holder does not yet exist

	//TODO(jacob): Add holder to allHolders

	return nil
}

func getPoliciesByHolder(stub *shim.ChaincodeStub, holder string) []Policy {
	//TODO(jacob): Check that holder has been registered

	//TODO(jacob): return policies held by holder
	
	var policies []Policy
	return policies
}
