package main

import(
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil{
		fmt.Printf("Error starting simple chaincode: %s", err)
	}
}

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	return nil, nil
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	return nil, nil
}

func (t *SimpleChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	return nil, nil
}
