package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

// Global HTTP Client initialized with cookie jar and insecure transport
var client *http.Client

func init() {
	// 1. Create a cookie jar to persist session cookies
	jar, _ := cookiejar.New(nil)

	// 2. Configure a Transport to skip SSL verification (necessary for localhost:5000)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 3. Create the HTTP Client
	client = &http.Client{
		Jar:       jar,
		Timeout:   30 * time.Second,
		Transport: tr,
	}
}

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
	symbols := "TSCO"
	url := fmt.Sprintf("https://localhost:5000/v1/api/trsrv/stocks?symbols=%s", symbols)

	// In a real app, use the authenticated client from your price.go file
	//resp, err := http.Get(url)
	//if err != nil {
	//	fmt.Printf("Request failed: %v\n", err)
	//	return
	//}
	//defer resp.Body.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("error creating request: %w", err)
		return
	}

	// Execute the request using the client with the cookie jar
	resp, err := client.Do(req)
	if err != nil {
		// Note: This often fails if the Gateway isn't running or auth failed.
		fmt.Printf("API request failed (is IBKR Gateway running and authenticated?): %w", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading response body: %w", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API request failed with status %s. Body: %s", resp.Status, string(body))
		return
	}

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