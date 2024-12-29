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

	decoder := json.NewDecoder(file)

	// Move to the first token (start of the object)
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, errors.New("invalid JSON format: expected object")
	}

	var filteredPools []Pool

	// Iterate over the keys in the top-level object
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to read key: %v", err)
		}

		// Check for "official" or "unOfficial"
		if key == "official" || key == "unOfficial" {
			if err := processPools(decoder, programID, solAddress, &filteredPools); err != nil {
				return nil, fmt.Errorf("error processing pools for key %s: %v", key, err)
			}
		} else {
			var skip json.RawMessage
			if err := decoder.Decode(&skip); err != nil {
				return nil, fmt.Errorf("failed to skip key %s: %v", key, err)
			}
		}
	}

	// Second filtering step: Remove pools with duplicate BaseMint or QuoteMint values
	fmt.Println("Total pools:", len(filteredPools))
	//result := removeDuplicatePools(filteredPools)
	//fmt.Println("Duplicate pools:", len(filteredPools)-len(result))
	return filteredPools, nil
}

// processPools processes a JSON array of pools, filtering and appending matching entries.
func processPools(decoder *json.Decoder, programID, solAddress string, result *[]Pool) error {
	token, err := decoder.Token()
	if err != nil || token != json.Delim('[') {
		return errors.New("invalid JSON format: expected array")
	}

	for decoder.More() {
		var poolData map[string]interface{}
		if err := decoder.Decode(&poolData); err != nil {
			return fmt.Errorf("failed to decode pool data: %v", err)
		}

		// Filtering logic for programID and solAddress
		// Extract and filter the pool data
		id, _ := poolData["id"].(string)
		baseMint, _ := poolData["baseMint"].(string)
		baseVault, _ := poolData["baseVault"].(string)
		quoteMint, _ := poolData["quoteMint"].(string)
		quoteVault, _ := poolData["quoteVault"].(string)
		programId, _ := poolData["programId"].(string)

		if programId == programID {
			// Append the pool if the base or quote mint matches the given address
			// Set Base as Solana

			if baseMint == solAddress {
				*result = append(*result, Pool{
					ID:         id,
					ProgramID:  programId,
					BaseMint:   baseMint,
					BaseVault:  baseVault,
					QuoteMint:  quoteMint,
					QuoteVault: quoteVault,
				})
			} else if quoteMint == solAddress {
				*result = append(*result, Pool{
					ID:         id,
					ProgramID:  programId,
					BaseMint:   quoteMint,
					BaseVault:  quoteVault,
					QuoteMint:  baseMint,
					QuoteVault: baseVault,
				})
			}
		}
	}

	token, err = decoder.Token()
	if err != nil || token != json.Delim(']') {
		return errors.New("invalid JSON format: expected end of array")
	}

	return nil
}

// removeDuplicatePools removes pools with duplicate BaseMint or QuoteMint values.
func removeDuplicatePools(pools []Pool) []Pool {
	seenCount := make(map[string]int)
	// Count occurrences of BaseMint and QuoteMint
	for _, pool := range pools {
		seenCount[pool.QuoteMint]++
	}
	// Filter out pools with duplicate BaseMint or QuoteMint
	var result []Pool
	for _, pool := range pools {
		// Only append pools with unique QuoteMint
		if seenCount[pool.QuoteMint] == 1 {
			result = append(result, pool)
		}
	}

	return result
}
