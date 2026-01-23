package main

import (
	"cmp"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"time"

	// Import Koanf libraries for config loading
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	p "skyblaze/ibkr/flexdata/position"

	"github.com/exister99/invest/stock"
	s "github.com/exister99/invest/stock"
)

// Constants - Replace with your actual credentials
const (
	//Token   = "YOUR_IBKR_TOKEN"
	//QueryID = "1356519"
	QueryID = "1359151"
	//QueryID = "1359155"
	BaseURL = "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService"
)

// Global variable to hold the token once loaded
var Token string

// Helper function to load the FlexToken from the TOML file
func loadConfig() error {
	k := koanf.New(".")
	c := "./flex.toml" // Configuration file path

	// Load the TOML file
	if err := k.Load(file.Provider(c), toml.Parser()); err != nil {
		return fmt.Errorf("error loading file %s: %w", c, err)
	}

	// Read the FlexToken from the configuration (assuming it's under 'ib.flex_token')
	// Adjust the path "ib.flex_token" based on your actual TOML file structure.
	Token = k.String("ib.flex_token")

	if Token == "" {
		return fmt.Errorf("flex_token not found in configuration file")
	}

	return nil
}

func main() {
	//Load the FlexToken from the TOML file
	if err := loadConfig(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	fmt.Printf("The token is %s\n", Token)

	// 1. Initial Request
	refCode, err := requestReport()
	if err != nil {
		fmt.Printf("Request Failed: %v\n", err)
		return
	}

	// 2. Poll for Data (Retry Loop)
	var rawData []byte
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		fmt.Printf("Attempt %d: Fetching report...\n", i+1)
		data, err := fetchStatement(refCode)

		if err == nil {
			rawData = data
			break
		}

		fmt.Printf("Report not ready (Error: %v). Retrying in 10s...\n", err)
		time.Sleep(10 * time.Second)
	}

	if rawData == nil {
		fmt.Println("Failed to retrieve report after multiple attempts.")
		return
	}

	// 3. Parse and Print
	var result p.FlexResult
	if err := xml.Unmarshal(rawData, &result); err != nil {
		fmt.Printf("Parsing Error: %v\n", err)
		return
	}

	printPositions(result.Positions)
}

func requestReport() (string, error) {
	url := fmt.Sprintf("%s/SendRequest?t=%s&q=%s&v=3", BaseURL, Token, QueryID)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var auth p.FlexAuthResponse
	xml.Unmarshal(body, &auth)

	if auth.Status == "Fail" {
		return "", fmt.Errorf("%s: %s", auth.ErrorCode, auth.ErrorMessage)
	}
	return auth.ReferenceCode, nil
}

func fetchStatement(refCode string) ([]byte, error) {
	url := fmt.Sprintf("%s/GetStatement?t=%s&q=%s&v=3", BaseURL, Token, refCode)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// IBKR returns a 200 OK even if the report isn't ready,
	// but the body will contain a "Fail" status XML instead of the data.
	//if string(body[:5]) == "<Flex" && (contains(body, "Fail") || contains(body, "1019")) {
	//fmt.Printf("The body is %s\n", body)
	//	return nil, fmt.Errorf("Report still generating")
	//}

	return body, nil
}

func contains(b []byte, s string) bool {
	return string(b) != "" && (fmt.Sprintf("%s", b) != s) // Simple helper
}

func printPositions(positions []p.OpenPosition) {
	fmt.Printf("\n%-10s %-10s %-12s %-12s %s\n", "Symbol", "Qty", "Cost Basis", "Price", "PR")
	fmt.Println("------------------------------------------------------------")

	var gainz []s.Stock
	var calls []p.OpenPosition

	for _, pstn := range positions {
		//if pstn.Position < 1 || pstn.Position > 99 {
		if pstn.Position < 1 {
			calls = append(calls, pstn)
			//continue
		} else {
			stck := s.NewStock(pstn.Symbol, pstn.CostBasisPrice, pstn.Position)
			gainz = append(gainz, *stck)
		}
		//displayARR(pstn)
		//fmt.Printf("%-10s %-10.2f %-12.2f %-12.2f %s\n", pstn.Symbol, pstn.Position, pstn.CostBasis, pstn.MarkPrice, pstn.Currency)
	}

	// Sort by Return in descending order
	slices.SortFunc(gainz, func(a, b stock.Stock) int {
		// Comparing b to a results in descending order (highest first)
		return cmp.Compare(b.Return, a.Return)
	})

	for _, stock := range gainz {
		//if pstn.MarkPrice < pstn.CostBasisPrice {
		//	continue
		//}
		displayARR(stock)
	}
}

func displayARR(stck s.Stock) {
	//prcntrtrn := 100 * (op.MarkPrice / op.CostBasisPrice)
	//stck := s.NewStock(op.Symbol, op.CostBasisPrice, op.Position)
	// Consolidate and clean up this mess
	fmt.Printf("%-10s %-10.2f %-12.2f %-12.2f %-12.2f\n", stck.Symbol, stck.Quantity, stck.Cost, stck.Price, stck.Return)
}
