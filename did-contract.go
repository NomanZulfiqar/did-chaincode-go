package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// DIDChaincode implements a simple chaincode to manage DIDs
type DIDChaincode struct {
}

// DIDDocument represents a DID document structure
type DIDDocument struct {
	DID         string    `json:"did"`
	LongFormDID string    `json:"longFormDid"`
	Document    string    `json:"document"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Version     int       `json:"version"`
	Recovered   bool      `json:"recovered,omitempty"`
	RecoveredAt time.Time `json:"recoveredAt,omitempty"`
}

// Init is called during chaincode instantiation
func (t *DIDChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode
func (t *DIDChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	
	switch function {
	case "InitLedger":
		return t.initLedger(stub)
	case "CreateDID":
		return t.createDID(stub, args)
	case "UpdateDID":
		return t.updateDID(stub, args)
	case "RecoverDID":
		return t.recoverDID(stub, args)
	case "GetDID":
		return t.getDID(stub, args)
	case "ListDIDs":
		return t.listDIDs(stub)
	default:
		return shim.Error("Invalid function name")
	}
}

// initLedger initializes the ledger
func (t *DIDChaincode) initLedger(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("DID Chaincode initialized")
	return shim.Success(nil)
}

// createDID anchors a new DID Document on Fabric
func (t *DIDChaincode) createDID(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3: did, longFormDid, documentJSON")
	}

	did := args[0]
	longFormDid := args[1]
	documentJSON := args[2]

	// Check if DID already exists
	existingDID, err := stub.GetState(did)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get DID: %s", err))
	}
	if existingDID != nil {
		return shim.Error(fmt.Sprintf("DID %s already exists", did))
	}

	// Create DID document
	didDocument := DIDDocument{
		DID:         did,
		LongFormDID: longFormDid,
		Document:    documentJSON,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
	}

	didJSON, err := json.Marshal(didDocument)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(did, didJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(didJSON)
}

// updateDID updates a DID Document
func (t *DIDChaincode) updateDID(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3: did, updatedDocumentJSON, operationSignature")
	}

	did := args[0]
	updatedDocumentJSON := args[1]
	// operationSignature := args[2] // TODO: Implement signature validation

	// Get existing DID document
	didJSON, err := stub.GetState(did)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get DID: %s", err))
	}
	if didJSON == nil {
		return shim.Error(fmt.Sprintf("DID %s does not exist", did))
	}

	var existingDID DIDDocument
	err = json.Unmarshal(didJSON, &existingDID)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Update DID document
	existingDID.Document = updatedDocumentJSON
	existingDID.UpdatedAt = time.Now()
	existingDID.Version++

	updatedJSON, err := json.Marshal(existingDID)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(did, updatedJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(updatedJSON)
}

// recoverDID recovers a lost DID
func (t *DIDChaincode) recoverDID(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3: did, newDocumentJSON, recoverySignature")
	}

	did := args[0]
	newDocumentJSON := args[1]
	// recoverySignature := args[2] // TODO: Implement signature validation

	// Get existing DID document
	didJSON, err := stub.GetState(did)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get DID: %s", err))
	}
	if didJSON == nil {
		return shim.Error(fmt.Sprintf("DID %s does not exist", did))
	}

	var existingDID DIDDocument
	err = json.Unmarshal(didJSON, &existingDID)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Recover DID document
	existingDID.Document = newDocumentJSON
	existingDID.UpdatedAt = time.Now()
	existingDID.Version++
	existingDID.Recovered = true
	existingDID.RecoveredAt = time.Now()

	recoveredJSON, err := json.Marshal(existingDID)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(did, recoveredJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(recoveredJSON)
}

// getDID retrieves a DID Document
func (t *DIDChaincode) getDID(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1: did")
	}

	did := args[0]
	didJSON, err := stub.GetState(did)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get DID: %s", err))
	}
	if didJSON == nil {
		return shim.Error(fmt.Sprintf("DID %s does not exist", did))
	}

	return shim.Success(didJSON)
}

// listDIDs returns all DIDs
func (t *DIDChaincode) listDIDs(stub shim.ChaincodeStubInterface) peer.Response {
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var dids []DIDDocument
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var did DIDDocument
		err = json.Unmarshal(queryResponse.Value, &did)
		if err != nil {
			return shim.Error(err.Error())
		}
		dids = append(dids, did)
	}

	didsJSON, err := json.Marshal(dids)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(didsJSON)
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(DIDChaincode)); err != nil {
		fmt.Printf("Error starting DID chaincode: %s", err)
	}
}