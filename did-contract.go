package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// DIDContract provides functions for managing DIDs
type DIDContract struct {
	contractapi.Contract
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

// InitLedger initializes the ledger
func (s *DIDContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("DID Chaincode initialized")
	return nil
}

// CreateDID anchors a new DID Document on Fabric
func (s *DIDContract) CreateDID(ctx contractapi.TransactionContextInterface, did string, longFormDid string, documentJSON string) error {
	// Check if DID already exists
	existingDID, err := ctx.GetStub().GetState(did)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if existingDID != nil {
		return fmt.Errorf("DID %s already exists", did)
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
		return err
	}

	return ctx.GetStub().PutState(did, didJSON)
}

// UpdateDID updates a DID Document
func (s *DIDContract) UpdateDID(ctx contractapi.TransactionContextInterface, did string, updatedDocumentJSON string, operationSignature string) error {
	// Get existing DID document
	didJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if didJSON == nil {
		return fmt.Errorf("DID %s does not exist", did)
	}

	var existingDID DIDDocument
	err = json.Unmarshal(didJSON, &existingDID)
	if err != nil {
		return err
	}

	// Update DID document
	existingDID.Document = updatedDocumentJSON
	existingDID.UpdatedAt = time.Now()
	existingDID.Version++

	updatedJSON, err := json.Marshal(existingDID)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(did, updatedJSON)
}

// RecoverDID recovers a lost DID
func (s *DIDContract) RecoverDID(ctx contractapi.TransactionContextInterface, did string, newDocumentJSON string, recoverySignature string) error {
	// Get existing DID document
	didJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if didJSON == nil {
		return fmt.Errorf("DID %s does not exist", did)
	}

	var existingDID DIDDocument
	err = json.Unmarshal(didJSON, &existingDID)
	if err != nil {
		return err
	}

	// Recover DID document
	existingDID.Document = newDocumentJSON
	existingDID.UpdatedAt = time.Now()
	existingDID.Version++
	existingDID.Recovered = true
	existingDID.RecoveredAt = time.Now()

	recoveredJSON, err := json.Marshal(existingDID)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(did, recoveredJSON)
}

// GetDID retrieves a DID Document
func (s *DIDContract) GetDID(ctx contractapi.TransactionContextInterface, did string) (*DIDDocument, error) {
	didJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if didJSON == nil {
		return nil, fmt.Errorf("DID %s does not exist", did)
	}

	var didDocument DIDDocument
	err = json.Unmarshal(didJSON, &didDocument)
	if err != nil {
		return nil, err
	}

	return &didDocument, nil
}

// ListDIDs returns all DIDs
func (s *DIDContract) ListDIDs(ctx contractapi.TransactionContextInterface) ([]*DIDDocument, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var dids []*DIDDocument
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var did DIDDocument
		err = json.Unmarshal(queryResponse.Value, &did)
		if err != nil {
			return nil, err
		}
		dids = append(dids, &did)
	}

	return dids, nil
}

func main() {
	didChaincode, err := contractapi.NewChaincode(&DIDContract{})
	if err != nil {
		fmt.Printf("Error creating DID chaincode: %v", err)
		return
	}

	if err := didChaincode.Start(); err != nil {
		fmt.Printf("Error starting DID chaincode: %v", err)
	}
}