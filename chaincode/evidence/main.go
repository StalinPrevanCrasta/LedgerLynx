package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// EvidenceContract provides functions for managing evidence
type EvidenceContract struct {
    contractapi.Contract
}

// Evidence describes basic details of an evidence item
type Evidence struct {
    ID               string    `json:"id"`
    CaseNumber       string    `json:"caseNumber"`
    Description      string    `json:"description"`
    CollectedBy      string    `json:"collectedBy"`
    CollectedAt      time.Time `json:"collectedAt"`
    CurrentCustodian string    `json:"currentCustodian"`
    CurrentOrg       string    `json:"currentOrg"`
    Status           string    `json:"status"`
    Location         string    `json:"location"`
    IPFSHash         string    `json:"ipfsHash"`
    EvidenceType     string    `json:"evidenceType"`
    CreatedAt        time.Time `json:"createdAt"`
    UpdatedAt        time.Time `json:"updatedAt"`
}

// HistoryRecord describes a custody transfer record
type HistoryRecord struct {
    TxID          string    `json:"txId"`
    Timestamp     time.Time `json:"timestamp"`
    FromOrg       string    `json:"fromOrg"`
    ToOrg         string    `json:"toOrg"`
    FromCustodian string    `json:"fromCustodian"`
    ToCustodian   string    `json:"toCustodian"`
    Action        string    `json:"action"`
    Evidence      Evidence  `json:"evidence"`
    Remarks       string    `json:"remarks"`
}

// InitLedger adds a base set of evidence to the ledger (for testing)
func (ec *EvidenceContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    fmt.Println("Initializing LedgerLynx Evidence Management System")
    return nil
}

// RegisterEvidence creates a new evidence record
func (ec *EvidenceContract) RegisterEvidence(ctx contractapi.TransactionContextInterface, 
    evidenceID string, caseNumber string, description string, collectedBy string, 
    location string, ipfsHash string, evidenceType string) error {
    
    // Check if evidence already exists
    evidenceJSON, err := ctx.GetStub().GetState(evidenceID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if evidenceJSON != nil {
        return fmt.Errorf("evidence %s already exists", evidenceID)
    }

    // Get submitting organization
    clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
    if err != nil {
        return fmt.Errorf("failed to get client MSP ID: %v", err)
    }

    // Validate that only authorized organizations can register evidence
    if clientMSPID != "PoliceDeptMSP" {
        return fmt.Errorf("only Police Department can register new evidence")
    }

    // Create evidence record
    evidence := Evidence{
        ID:               evidenceID,
        CaseNumber:       caseNumber,
        Description:      description,
        CollectedBy:      collectedBy,
        CollectedAt:      time.Now(),
        CurrentCustodian: collectedBy,
        CurrentOrg:       clientMSPID,
        Status:           "COLLECTED",
        Location:         location,
        IPFSHash:         ipfsHash,
        EvidenceType:     evidenceType,
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
    }

    evidenceJSON, err = json.Marshal(evidence)
    if err != nil {
        return err
    }

    // Create initial history record
    txID := ctx.GetStub().GetTxID()
    historyRecord := HistoryRecord{
        TxID:          txID,
        Timestamp:     time.Now(),
        FromOrg:       "",
        ToOrg:         clientMSPID,
        FromCustodian: "",
        ToCustodian:   collectedBy,
        Action:        "REGISTER_EVIDENCE",
        Evidence:      evidence,
        Remarks:       "Evidence initially registered",
    }

    historyJSON, err := json.Marshal(historyRecord)
    if err != nil {
        return err
    }

    // Store history record with composite key
    historyKey, err := ctx.GetStub().CreateCompositeKey("history", []string{evidenceID, txID})
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(historyKey, historyJSON)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(evidenceID, evidenceJSON)
}

// TransferCustody transfers evidence custody between organizations
func (ec *EvidenceContract) TransferCustody(ctx contractapi.TransactionContextInterface,
    evidenceID string, toCustodian string, toOrg string, remarks string) error {
    
    // Get current evidence state
    evidenceJSON, err := ctx.GetStub().GetState(evidenceID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if evidenceJSON == nil {
        return fmt.Errorf("evidence %s does not exist", evidenceID)
    }

    var evidence Evidence
    err = json.Unmarshal(evidenceJSON, &evidence)
    if err != nil {
        return err
    }

    // Get submitting organization
    clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
    if err != nil {
        return fmt.Errorf("failed to get client MSP ID: %v", err)
    }

    // Verify current custodian organization can transfer
    if evidence.CurrentOrg != clientMSPID {
        return fmt.Errorf("only current custodian organization can transfer evidence")
    }

    // Validate transfer logic
    validTransfers := map[string][]string{
        "PoliceDeptMSP":    {"ForensicsLabMSP", "CourtSystemMSP"},
        "ForensicsLabMSP":  {"PoliceDeptMSP", "CourtSystemMSP"},
        "CourtSystemMSP":   {"PoliceDeptMSP", "ForensicsLabMSP"},
    }

    if !contains(validTransfers[clientMSPID], toOrg) {
        return fmt.Errorf("invalid transfer from %s to %s", clientMSPID, toOrg)
    }

    // Update evidence record
    fromCustodian := evidence.CurrentCustodian
    fromOrg := evidence.CurrentOrg
    
    evidence.CurrentCustodian = toCustodian
    evidence.CurrentOrg = toOrg
    evidence.Status = "TRANSFERRED"
    evidence.UpdatedAt = time.Now()

    evidenceJSON, err = json.Marshal(evidence)
    if err != nil {
        return err
    }

    // Create history record
    txID := ctx.GetStub().GetTxID()
    historyRecord := HistoryRecord{
        TxID:          txID,
        Timestamp:     time.Now(),
        FromOrg:       fromOrg,
        ToOrg:         toOrg,
        FromCustodian: fromCustodian,
        ToCustodian:   toCustodian,
        Action:        "TRANSFER_CUSTODY",
        Evidence:      evidence,
        Remarks:       remarks,
    }

    historyJSON, err := json.Marshal(historyRecord)
    if err != nil {
        return err
    }

    // Store history record with composite key
    historyKey, err := ctx.GetStub().CreateCompositeKey("history", []string{evidenceID, txID})
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(historyKey, historyJSON)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(evidenceID, evidenceJSON)
}

// UpdateEvidenceStatus updates the status of evidence
func (ec *EvidenceContract) UpdateEvidenceStatus(ctx contractapi.TransactionContextInterface,
    evidenceID string, newStatus string, remarks string) error {
    
    evidenceJSON, err := ctx.GetStub().GetState(evidenceID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if evidenceJSON == nil {
        return fmt.Errorf("evidence %s does not exist", evidenceID)
    }

    var evidence Evidence
    err = json.Unmarshal(evidenceJSON, &evidence)
    if err != nil {
        return err
    }

    // Get submitting organization
    clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
    if err != nil {
        return fmt.Errorf("failed to get client MSP ID: %v", err)
    }

    // Verify current custodian organization can update
    if evidence.CurrentOrg != clientMSPID {
        return fmt.Errorf("only current custodian organization can update evidence status")
    }

    // Validate status transitions
    validStatuses := []string{"COLLECTED", "TRANSFERRED", "UNDER_ANALYSIS", "ANALYZED", "IN_COURT", "DISPOSED"}
    if !contains(validStatuses, newStatus) {
        return fmt.Errorf("invalid status: %s", newStatus)
    }

    oldStatus := evidence.Status
    evidence.Status = newStatus
    evidence.UpdatedAt = time.Now()

    evidenceJSON, err = json.Marshal(evidence)
    if err != nil {
        return err
    }

    // Create history record
    txID := ctx.GetStub().GetTxID()
    historyRecord := HistoryRecord{
        TxID:          txID,
        Timestamp:     time.Now(),
        FromOrg:       clientMSPID,
        ToOrg:         clientMSPID,
        FromCustodian: evidence.CurrentCustodian,
        ToCustodian:   evidence.CurrentCustodian,
        Action:        "UPDATE_STATUS",
        Evidence:      evidence,
        Remarks:       fmt.Sprintf("Status changed from %s to %s. %s", oldStatus, newStatus, remarks),
    }

    historyJSON, err := json.Marshal(historyRecord)
    if err != nil {
        return err
    }

    // Store history record with composite key
    historyKey, err := ctx.GetStub().CreateCompositeKey("history", []string{evidenceID, txID})
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(historyKey, historyJSON)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(evidenceID, evidenceJSON)
}

// GetEvidence returns the evidence record for given ID
func (ec *EvidenceContract) GetEvidence(ctx contractapi.TransactionContextInterface, evidenceID string) (*Evidence, error) {
    evidenceJSON, err := ctx.GetStub().GetState(evidenceID)
    if err != nil {
        return nil, fmt.Errorf("failed to read from world state: %v", err)
    }
    if evidenceJSON == nil {
        return nil, fmt.Errorf("evidence %s does not exist", evidenceID)
    }

    var evidence Evidence
    err = json.Unmarshal(evidenceJSON, &evidence)
    if err != nil {
        return nil, err
    }

    return &evidence, nil
}

// GetEvidenceHistory returns the history of custody transfers for evidence
func (ec *EvidenceContract) GetEvidenceHistory(ctx contractapi.TransactionContextInterface, evidenceID string) ([]*HistoryRecord, error) {
    // Query history records using composite key
    historyIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("history", []string{evidenceID})
    if err != nil {
        return nil, err
    }
    defer historyIterator.Close()

    var history []*HistoryRecord
    for historyIterator.HasNext() {
        response, err := historyIterator.Next()
        if err != nil {
            return nil, err
        }

        var record HistoryRecord
        err = json.Unmarshal(response.Value, &record)
        if err != nil {
            return nil, err
        }
        history = append(history, &record)
    }

    return history, nil
}

// GetAllEvidence returns all evidence records (for admin purposes)
func (ec *EvidenceContract) GetAllEvidence(ctx contractapi.TransactionContextInterface) ([]*Evidence, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var evidenceList []*Evidence
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        // Skip history records and other composite keys
        if len(queryResponse.Key) > 7 && queryResponse.Key[:7] == "history" {
            continue
        }

        var evidence Evidence
        err = json.Unmarshal(queryResponse.Value, &evidence)
        if err != nil {
            return nil, err
        }
        evidenceList = append(evidenceList, &evidence)
    }

    return evidenceList, nil
}

// GetEvidenceByCase returns all evidence for a specific case
func (ec *EvidenceContract) GetEvidenceByCase(ctx contractapi.TransactionContextInterface, caseNumber string) ([]*Evidence, error) {
    queryString := fmt.Sprintf(`{"selector":{"caseNumber":"%s"}}`, caseNumber)
    
    resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var evidenceList []*Evidence
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var evidence Evidence
        err = json.Unmarshal(queryResponse.Value, &evidence)
        if err != nil {
            return nil, err
        }
        evidenceList = append(evidenceList, &evidence)
    }

    return evidenceList, nil
}

// GetEvidenceByOrganization returns all evidence currently held by an organization
func (ec *EvidenceContract) GetEvidenceByOrganization(ctx contractapi.TransactionContextInterface, orgMSP string) ([]*Evidence, error) {
    queryString := fmt.Sprintf(`{"selector":{"currentOrg":"%s"}}`, orgMSP)
    
    resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var evidenceList []*Evidence
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var evidence Evidence
        err = json.Unmarshal(queryResponse.Value, &evidence)
        if err != nil {
            return nil, err
        }
        evidenceList = append(evidenceList, &evidence)
    }

    return evidenceList, nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

func main() {
    evidenceChaincode, err := contractapi.NewChaincode(&EvidenceContract{})
    if err != nil {
        log.Panicf("Error creating evidence chaincode: %v", err)
    }

    if err := evidenceChaincode.Start(); err != nil {
        log.Panicf("Error starting evidence chaincode: %v", err)
    }
}
