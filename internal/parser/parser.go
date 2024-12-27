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

// ParseAndFilter processes JSON data in a memory-efficient way.
func ParseAndFilter(filepath string, programID, solAddress string) ([]Pool, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create a JSON decoder for streaming
	decoder := json.NewDecoder(file)

	// Move to the first token (start of the object)
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, errors.New("invalid JSON format: expected object")
	}

	var result []Pool

	// Iterate over the keys in the top-level object
	for decoder.More() {
		// Read the key
		key, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to read key: %v", err)
		}

		// Check if key is "official" or "unOfficial"
		if key == "official" || key == "unOfficial" {
			// Read the array
			if err := processPools(decoder, programID, solAddress, &result); err != nil {
				return nil, fmt.Errorf("error processing pools for key %s: %v", key, err)
			}
		} else {
			// Skip unknown key's value by decoding into json.RawMessage
			var skip json.RawMessage
			if err := decoder.Decode(&skip); err != nil {
				return nil, fmt.Errorf("failed to skip key %s: %v", key, err)
			}
		}
	}

	return result, nil
}

// processPools processes a JSON array of pools, filtering and appending matching entries.
func processPools(decoder *json.Decoder, programID, solAddress string, result *[]Pool) error {
	// Ensure the next token is the start of an array
	token, err := decoder.Token()
	if err != nil || token != json.Delim('[') {
		return errors.New("invalid JSON format: expected array")
	}

	// Iterate through the array
	for decoder.More() {
		var poolData map[string]interface{}
		if err := decoder.Decode(&poolData); err != nil {
			return fmt.Errorf("failed to decode pool data: %v", err)
		}

		// Extract and filter the pool data
		id, _ := poolData["id"].(string)
		baseMint, _ := poolData["baseMint"].(string)
		baseVault, _ := poolData["baseVault"].(string)
		quoteVault, _ := poolData["quoteVault"].(string)
		programId, _ := poolData["programId"].(string)
		quoteMint, _ := poolData["quoteMint"].(string)

		if programId == programID && (baseMint == solAddress || quoteMint == solAddress) {
			*result = append(*result, Pool{
				ID:         id,
				BaseMint:   baseMint,
				BaseVault:  baseVault,
				QuoteVault: quoteVault,
				ProgramID:  programId,
				QuoteMint:  quoteMint,
			})
		}

	}

	// Ensure the next token is the end of the array
	token, err = decoder.Token()
	if err != nil || token != json.Delim(']') {
		return errors.New("invalid JSON format: expected end of array")
	}

	return nil
}
