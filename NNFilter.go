package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Define the nested structure returned by /trsrv/stocks
type Contract struct {
	ConID    int    `json:"conid"`
	Exchange string `json:"exchange"`
	IsUS     bool   `json:"isUS"`
}

type SymbolData struct {
	Name       string     `json:"name"`
	AssetClass string     `json:"assetClass"`
	Contracts  []Contract `json:"contracts"`
}

func main() {
	symbols := "TSCO,AAPL,MSFT"
	url := fmt.Sprintf("https://localhost:5000/v1/api/trsrv/stocks?symbols=%s", symbols)

	// In a real app, use the authenticated client from your price.go file
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// The API returns a map where keys are the symbols (e.g., "AAPL": [...])
	var result map[string][]SymbolData
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Unmarshal error: %v\n", err)
		return
	}

	validConids := make([]int, 0)

	// Iterate through each symbol in the map
	for symbol, dataList := range result {
		fmt.Printf("Processing %s...\n", symbol)
		
		for _, data := range dataList {
			for _, contract := range data.Contracts {
				ex := strings.ToUpper(contract.Exchange)
				
				// Filter for NASDAQ (ISLAND) or NYSE
				if ex == "NASDAQ" || ex == "ISLAND" || ex == "NYSE" {
					fmt.Printf("  - Found %s on %s: %d\n", symbol, ex, contract.ConID)
					validConids = append(validConids, contract.ConID)
				}
			}
		}
	}

	fmt.Printf("\nFinal filtered conids: %v\n", validConids)
}