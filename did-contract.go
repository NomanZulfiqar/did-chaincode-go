package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// DIDChaincode implements a simple chaincode to manage DIDs
// Version 1.2x - Enhanced for two-organization deployment with full DID requirements
type DIDChaincode struct {
	Version string
}

// GetVersion returns the chaincode version
func (t *DIDChaincode) GetVersion() string {
	return "1.2x"
}

// DIDDocument represents a DID document structure
// Enhanced for two-organization network (CompanyA & CompanyB)
type DIDDocument struct {
	DID         string    `json:"did"`
	LongFormDID string    `json:"longFormDid"`
	Document    string    `json:"document"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Version     int       `json:"version"`
	Recovered   bool      `json:"recovered,omitempty"`
	RecoveredAt time.Time `json:"recoveredAt,omitempty"`
	UpdateKey   string    `json:"updateKey,omitempty"`   // Public key for updates
	RecoveryKey string    `json:"recoveryKey,omitempty"` // Public key for recovery
	CreatedBy   string    `json:"createdBy,omitempty"`   // Organization that created the DID
	EndorsedBy  []string  `json:"endorsedBy,omitempty"`  // Organizations that endorsed operations
}

// validateSignature performs basic signature validation (simplified for demo)
func (t *DIDChaincode) validateSignature(message, signature, publicKey string) bool {
	// Simplified validation: check if signature contains hash of message + key
	// In production, use proper cryptographic signature verification
	if signature == "" || publicKey == "" {
		return false
	}
	
	// Create expected signature hash
	hash := sha256.Sum256([]byte(message + publicKey))
	expectedSig := hex.EncodeToString(hash[:])
	
	// Check if provided signature matches or contains expected pattern
	return strings.Contains(signature, expectedSig[:16]) // First 16 chars for demo
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
	case "GetVersion":
		return t.getVersion(stub)
	case "GetNetworkInfo":
		return t.getNetworkInfo(stub)
	default:
		return shim.Error("Invalid function name")
	}
}

// initLedger initializes the ledger
func (t *DIDChaincode) initLedger(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("DID Chaincode v1.2x initialized for two-organization network")
	fmt.Println("Supporting CompanyA (m-FQEEX22AZNEGDDJL4WCQP6KYHU) and CompanyB (m-JLGL2ZEX6BDIXIEFYD4RJVZSTI)")
	return shim.Success([]byte("DID Chaincode v1.2x initialized successfully"))
}

// createDID anchors a new DID Document on Fabric
func (t *DIDChaincode) createDID(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) < 3 || len(args) > 5 {
		return shim.Error("Incorrect number of arguments. Expecting 3-5: did, longFormDid, documentJSON, [updateKey], [recoveryKey]")
	}

	did := args[0]
	longFormDid := args[1]
	documentJSON := args[2]
	
	// Optional keys for signature validation
	var updateKey, recoveryKey string
	if len(args) >= 4 {
		updateKey = args[3]
	}
	if len(args) >= 5 {
		recoveryKey = args[4]
	}

	// Check if DID already exists
	existingDID, err := stub.GetState(did)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get DID: %s", err))
	}
	if existingDID != nil {
		return shim.Error(fmt.Sprintf("DID %s already exists", did))
	}

	// Get deterministic timestamp from transaction
	txTimestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get transaction timestamp: %s", err))
	}
	txTime := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos))

	// Get creator organization info
	creator, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get creator: %s", err))
	}
	
	// Extract organization from creator (simplified)
	createdBy := "unknown"
	if strings.Contains(string(creator), "m-FQEEX22AZNEGDDJL4WCQP6KYHU") {
		createdBy = "CompanyA"
	} else if strings.Contains(string(creator), "m-JLGL2ZEX6BDIXIEFYD4RJVZSTI") {
		createdBy = "CompanyB"
	}

	// Create DID document with organization tracking
	didDocument := DIDDocument{
		DID:         did,
		LongFormDID: longFormDid,
		Document:    documentJSON,
		CreatedAt:   txTime,
		UpdatedAt:   txTime,
		Version:     1,
		UpdateKey:   updateKey,
		RecoveryKey: recoveryKey,
		CreatedBy:   createdBy,
		EndorsedBy:  []string{createdBy},
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
	operationSignature := args[2]

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

	// Get deterministic timestamp from transaction
	txTimestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get transaction timestamp: %s", err))
	}
	txTime := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos))

	// Validate signature if updateKey exists
	if existingDID.UpdateKey != "" {
		message := fmt.Sprintf("%s:%s:%d", did, updatedDocumentJSON, existingDID.Version+1)
		if !t.validateSignature(message, operationSignature, existingDID.UpdateKey) {
			return shim.Error("Invalid operation signature for update")
		}
	}

	// Get updater organization info
	creator, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get creator: %s", err))
	}
	
	updatedBy := "unknown"
	if strings.Contains(string(creator), "m-FQEEX22AZNEGDDJL4WCQP6KYHU") {
		updatedBy = "CompanyA"
	} else if strings.Contains(string(creator), "m-JLGL2ZEX6BDIXIEFYD4RJVZSTI") {
		updatedBy = "CompanyB"
	}

	// Update DID document with endorsement tracking
	existingDID.Document = updatedDocumentJSON
	existingDID.UpdatedAt = txTime
	existingDID.Version++
	
	// Add to endorsed by list if not already present
	found := false
	for _, org := range existingDID.EndorsedBy {
		if org == updatedBy {
			found = true
			break
		}
	}
	if !found {
		existingDID.EndorsedBy = append(existingDID.EndorsedBy, updatedBy)
	}

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
	recoverySignature := args[2]

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

	// Get deterministic timestamp from transaction
	txTimestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get transaction timestamp: %s", err))
	}
	txTime := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos))

	// Validate recovery signature if recoveryKey exists
	if existingDID.RecoveryKey != "" {
		message := fmt.Sprintf("%s:recovery:%s:%d", did, newDocumentJSON, existingDID.Version+1)
		if !t.validateSignature(message, recoverySignature, existingDID.RecoveryKey) {
			return shim.Error("Invalid recovery signature")
		}
	}

	// Get recoverer organization info
	creator, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get creator: %s", err))
	}
	
	recoveredBy := "unknown"
	if strings.Contains(string(creator), "m-FQEEX22AZNEGDDJL4WCQP6KYHU") {
		recoveredBy = "CompanyA"
	} else if strings.Contains(string(creator), "m-JLGL2ZEX6BDIXIEFYD4RJVZSTI") {
		recoveredBy = "CompanyB"
	}

	// Recover DID document with endorsement tracking
	existingDID.Document = newDocumentJSON
	existingDID.UpdatedAt = txTime
	existingDID.Version++
	existingDID.Recovered = true
	existingDID.RecoveredAt = txTime
	
	// Add to endorsed by list if not already present
	found := false
	for _, org := range existingDID.EndorsedBy {
		if org == recoveredBy {
			found = true
			break
		}
	}
	if !found {
		existingDID.EndorsedBy = append(existingDID.EndorsedBy, recoveredBy)
	}

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

// getVersion returns the chaincode version
func (t *DIDChaincode) getVersion(stub shim.ChaincodeStubInterface) peer.Response {
	version := map[string]string{
		"version": "1.2x",
		"description": "DID Chaincode for two-organization network with full DID requirements",
		"organizations": "CompanyA (m-FQEEX22AZNEGDDJL4WCQP6KYHU), CompanyB (m-JLGL2ZEX6BDIXIEFYD4RJVZSTI)",
	}
	
	versionJSON, err := json.Marshal(version)
	if err != nil {
		return shim.Error(err.Error())
	}
	
	return shim.Success(versionJSON)
}

// getNetworkInfo returns network information
func (t *DIDChaincode) getNetworkInfo(stub shim.ChaincodeStubInterface) peer.Response {
	networkInfo := map[string]interface{}{
		"chaincode_version": "1.2x",
		"network_type": "two-organization",
		"organizations": []map[string]string{
			{
				"name": "CompanyA",
				"msp_id": "m-FQEEX22AZNEGDDJL4WCQP6KYHU",
				"peer": "nd-lhf6gjm2mrg2bkl4k2fycpwrd4.m-fqeex22aznegddjl4wcqp6kyhu.n-lhs7rblbt5drppe2pfry3il3yu.managedblockchain.us-east-1.amazonaws.com:30003",
			},
			{
				"name": "CompanyB", 
				"msp_id": "m-JLGL2ZEX6BDIXIEFYD4RJVZSTI",
				"peer": "nd-7sfv4dmoobf77guclpma7za2je.m-jlgl2zex6bdixiefyd4rjvzsti.n-lhs7rblbt5drppe2pfry3il3yu.managedblockchain.us-east-1.amazonaws.com:30006",
			},
		},
		"channel": "mychannel",
		"endorsement_policy": "MAJORITY (requires both organizations)",
	}
	
	infoJSON, err := json.Marshal(networkInfo)
	if err != nil {
		return shim.Error(err.Error())
	}
	
	return shim.Success(infoJSON)
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(DIDChaincode)); err != nil {
		fmt.Printf("Error starting DID chaincode: %s", err)
	}
}