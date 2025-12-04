package main

import (
	"encoding/xml"
	"fmt"
	"log" // Added for error logging
	"net/http"
	"time"

    // Import Koanf libraries for config loading
	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"

	fx "skyblaze/ibkr/flexdata"
	s "github.com/exister99/invest/stock"
)

// Configuration - FlexToken is now removed from constants
const (
	QueryID   = "1337940" // e.g., "12345"
	BaseURL   = "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService"
)

// Global variable to hold the token once loaded
var FlexToken string

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
	FlexToken = k.String("ib.flex_token")
	
	if FlexToken == "" {
		return fmt.Errorf("flex_token not found in configuration file")
	}
	
	return nil
}

func main() {
    positions := make(map[string]*s.Stock)

	// 1. Load the FlexToken from the TOML file
	if err := loadConfig(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	// Step 2: Request the Report (using the loaded global FlexToken)
	reqURL := fmt.Sprintf("%s/SendRequest?t=%s&q=%s&v=3", BaseURL, FlexToken, QueryID)
	resp, err := http.Get(reqURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var initResp fx.FlexStatementResponse
	if err := xml.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		panic(err)
	}

	if initResp.Status != "Success" {
		fmt.Printf("Error requesting report: %s\n", initResp.ErrorMessage)
		return
	}

	fmt.Printf("Report requested. Reference Code: %s. Waiting 10s...\n", initResp.ReferenceCode)
	
	// Step 3: Wait for generation (mandatory)
	time.Sleep(10 * time.Second)

	// Step 4: Retrieve the Report
	getURL := fmt.Sprintf("%s/GetStatement?q=%s&t=%s&v=3", BaseURL, initResp.ReferenceCode, FlexToken)
	reportResp, err := http.Get(getURL)
	if err != nil {
		panic(err)
	}
	defer reportResp.Body.Close()

	// Parse the actual trade data
	var data fx.FlexQueryResponse
	
	if err := xml.NewDecoder(reportResp.Body).Decode(&data); err != nil {
		fmt.Printf("Error parsing report (check if report is ready): %v\n", err)
		return
	}

	// Step 5: Output
	trades := data.FlexStatements.FlexStatement.Trades.Trade
	fmt.Printf("\nFound %d historical trades:\n", len(trades))
	
	for _, t := range trades {
		_, exists := positions[t.Symbol]
		if !exists {
			positions[t.Symbol] = s.NewStock(t.Symbol)
		} else {
			positions[t.Symbol].AddTrx(&t)
		}

		fmt.Printf("%s | %s %s | Qty: %.0f | Price: %.2f\n", 
			t.TradeDate, t.BuySell, t.Symbol, t.Quantity, t.Price)
	}

   positions["IBM"].PrintTrx()

}

