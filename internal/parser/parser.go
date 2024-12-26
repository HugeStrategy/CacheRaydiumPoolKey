package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Pool struct {
	ID         string `json:"id"`
	BaseMint   string `json:"baseMint"`
	BaseVault  string `json:"baseVault"`
	QuoteVault string `json:"quoteVault"`
	ProgramID  string `json:"programId"`
	QuoteMint  string `json:"quoteMint"`
}

// ParseAndFilter processes both "official" and "unOfficial" arrays from the JSON.
func ParseAndFilter(filepath string, programID, quoteMint string) ([]Pool, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Use a generic map to hold the parsed JSON
	var data map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v", err)
	}

	// Helper function to process pools
	processPools := func(rawPools []interface{}) ([]Pool, error) {
		var processed []Pool
		for _, item := range rawPools {
			poolData, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			// Extract required fields, using default values if not present
			id, _ := poolData["id"].(string)
			baseMint, _ := poolData["baseMint"].(string)
			baseVault, _ := poolData["baseVault"].(string)
			quoteVault, _ := poolData["quoteVault"].(string)
			programId, _ := poolData["programId"].(string)
			quoteMintValue, _ := poolData["quoteMint"].(string)

			// Apply filter criteria
			if programId == programID && quoteMintValue == quoteMint {
				processed = append(processed, Pool{
					ID:         id,
					BaseMint:   baseMint,
					BaseVault:  baseVault,
					QuoteVault: quoteVault,
					ProgramID:  programId,
					QuoteMint:  quoteMintValue,
				})
			}
		}
		return processed, nil
	}

	// Extract and process the "official" pools
	official, ok := data["official"].([]interface{})
	if !ok {
		return nil, errors.New(`missing or invalid "official" array in JSON`)
	}
	officialPools, err := processPools(official)
	if err != nil {
		return nil, fmt.Errorf("error processing official pools: %v", err)
	}

	// Extract and process the "unOfficial" pools
	unOfficial, ok := data["unOfficial"].([]interface{})
	if !ok {
		unOfficial = []interface{}{} // Handle missing unOfficial gracefully
	}
	unOfficialPools, err := processPools(unOfficial)
	if err != nil {
		return nil, fmt.Errorf("error processing unOfficial pools: %v", err)
	}

	// Combine results from both arrays
	return append(officialPools, unOfficialPools...), nil
}
