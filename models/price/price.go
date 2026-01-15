package price

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv" // Added for string to int conversion
	"strings"
	"time"
)

// The base URL for the IBKR Client Portal Gateway
const BaseURL = "https://localhost:5000/v1/api"

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

// --- API Response Structs ---

// SecdefSearchResponse represents an item returned from the contract search.
type SecdefSearchResponse []struct {
	// IMPORTANT FIX: Changed ConID type from 'int' to 'string'.
	// The IBKR API can sometimes return non-numeric error strings here,
	// causing an unmarshal error if it's expecting an integer.
	ConID    string `json:"conid"` // The unique Contract ID we need
	Symbol   string `json:"symbol"`
	SecType  string `json:"secType"`
	Exchange string `json:"exchange"`
	Description string `json:"description"`
}

// MarketDataSnapshotResponse represents the response from the market data snapshot.
// The response is an an array of objects, where the keys are tick type numbers (e.g., "31" for Last Price).
type MarketDataSnapshotResponse []map[string]interface{}

// --- Utility Functions ---

// apiCall performs a GET request to the specified IBKR endpoint.
func apiCall(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", BaseURL, endpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Execute the request using the client with the cookie jar
	resp, err := client.Do(req)
	if err != nil {
		// Note: This often fails if the Gateway isn't running or auth failed.
		return nil, fmt.Errorf("API request failed (is IBKR Gateway running and authenticated?): %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %s. Body: %s", resp.Status, string(body))
	}

	return body, nil
}

// getConid takes a stock symbol and returns its Contract ID (conid) as an integer.
// It prioritizes results from NASDAQ (ISLAND) and NYSE.
func getConid(symbol string) (int, error) {
	// Endpoint: /iserver/secdef/search?symbol={symbol}&name=false&secType=STK
	endpoint := fmt.Sprintf("iserver/secdef/search?symbol=%s&name=false&secType=STK", symbol)

//what about using /trsrv/stocks
//https://www.interactivebrokers.com/campus/ibkr-api-page/cpapi-v1/#trsrv-stock-contract

	body, err := apiCall(endpoint)
	if err != nil {
		return 0, err
	}

	var results SecdefSearchResponse
	if err := json.Unmarshal(body, &results); err != nil {
		return 0, fmt.Errorf("error decoding search response: %w", err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("no contract found for symbol: %s", symbol)
	}

	var selectedConid string
	//var ex string
	
	// 1. Priority Loop: Look for NASDAQ (ISLAND) or NYSE
	for _, res := range results {
		fmt.Printf("The response is %v\n", res)
		//ex = strings.ToUpper(res.Exchange)
		// IBKR refers to NASDAQ as "ISLAND" in many API responses
		//if true {//ex == "NASDAQ" || ex == "ISLAND" || ex == "NYSE" {
			//selectedConid = res.ConID
			//fmt.Printf("We have conid %s on exchange %s for symbol %s\n", selectedConid, ex, symbol)
			//fmt.Printf("The response is %v\n", res)
			//break 
		//}
	}



	// 2. Fallback: If no US major exchange found, take the first available result
	if selectedConid == "" {
		selectedConid = results[0].ConID
	}

	// Convert the string conid to an integer before returning
	conid, err := strconv.Atoi(selectedConid)
	if err != nil {
		return 0, fmt.Errorf("failed to convert conid string '%s' to integer: %w", selectedConid, err)
	}

	return conid, nil
}

// getCurrentPrice takes a conid and returns the last traded price.
func getCurrentPrice(conid int) (float64, error) {
	// Endpoint: /iserver/marketdata/snapshot?conids={conid}&fields=31
	// Field 31 is the 'Last Price'
	endpoint := fmt.Sprintf("iserver/marketdata/snapshot?conids=%d&fields=31", conid)

	body, err := apiCall(endpoint)
	//attempts := 100
	//for err != nil && attempts > 0 {
	//		//time.Sleep(1 * time.Second)
	//		attempts--
	//		body, err = apiCall(endpoint)
	//	}

	if err != nil {
		return 0.0, err
	}

	var snapshot MarketDataSnapshotResponse
	if err := json.Unmarshal(body, &snapshot); err != nil {
		return 0.0, fmt.Errorf("error decoding market data snapshot: %w", err)
	}

	if len(snapshot) == 0 {
		return 0.0, fmt.Errorf("no market data returned for conid: %d", conid)
	}

	fmt.Printf("The snapshot is %v\n", snapshot)

	// The response is an array of maps, where the key "31" holds the last price.
	priceVal, ok := snapshot[0]["31"]

	// --- FIX: Improved error handling for missing data/subscriptions ---
	if !ok {
		// If "31" (Last Price) is missing, check if the API returned an explicit error
		if errorVal, exists := snapshot[0]["error"]; exists {
			return 0.0, fmt.Errorf("API returned error for conid %d: %v", conid, errorVal)
		}
		return 0.0, fmt.Errorf("field 31 (Last Price) not found in snapshot response. This strongly suggests missing market data permissions or closed markets")
	}
	// --- End FIX ---

	price, ok := priceVal.(float64)
	if !ok {
		// Handle case where price might be returned as a string and needs parsing
		priceStr, ok := priceVal.(string)
		if ok {
			// Attempt to parse the string to float64
			// 1. Remove the "C" prefix
			cleanPrice := strings.TrimPrefix(priceStr, "C")
			parsedPrice, parseErr := strconv.ParseFloat(cleanPrice, 64)
			if parseErr != nil {
				return 0.0, fmt.Errorf("price value is string '%s' but failed to parse: %w", priceStr, parseErr)
			}
			return parsedPrice, nil
		}
		return 0.0, fmt.Errorf("price value is neither a float64 nor a string (type was %T)", priceVal)
	}

	return price, nil
}

func Price(symbol string) (float64, error) {
	// Check for a symbol argument
	//if len(os.Args) < 2 {
	//	fmt.Println("Usage: go run stock_price.go <STOCK_SYMBOL>")
	//	fmt.Println("Example: go run stock_price.go AAPL")
	//	os.Exit(1)
	//}
	//symbol := os.Args[1]

	//fmt.Printf("--- Fetching data for %s ---\n", symbol)

	// 1. Get Contract ID (conid)
	conid, err := getConid(symbol)
	if err != nil {
		fmt.Printf("Error finding Contract ID for %s: %v\n", symbol, err)
		// Check for specific Gateway connection error hints
		if err.Error() == "API request failed (is IBKR Gateway running and authenticated?): API request failed with status 503 Service Unavailable. Body: Server error" {
			fmt.Println("\n-- IMPORTANT NOTE --")
			fmt.Println("It looks like the IBKR Client Portal Gateway is not running or not reachable.")
			fmt.Println("Please ensure the Client Portal Gateway is running locally on port 5000 and you have successfully logged in.")
			fmt.Println("--------------------")
		}
		return 0.0, err
	}
	//fmt.Printf("Found Contract ID (conid): %d\n", conid)

	// 2. Get Current Price
	price, err := getCurrentPrice(conid)
	attempts := 100
	for err != nil && attempts > 0 {
		//time.Sleep(1 * time.Second)
		attempts--
		price, err = getCurrentPrice(conid)
	}

	if err != nil {
		fmt.Printf("Error fetching price for conid %d: %v\n", conid, err)
		return 0.0, err
	}

	//fmt.Printf("\n============================================\n")
	//fmt.Printf("Current Last Traded Price for %s: $%.2f\n", symbol, price)
	//fmt.Printf("============================================\n")

	// Note: You must ensure the IBKR Client Portal Gateway is running and authenticated
	// before running this program.

	return price, nil

}
