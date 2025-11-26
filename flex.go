package main

import (
	"encoding/xml"
	"fmt"
	//"io"
	"net/http"
	"time"
)

// Configuration
const (
	FlexToken = "147372257548178188653193"
	QueryID   = "1337940" // e.g., "12345"
	BaseURL   = "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService"
)

// --- XML Structs for Parsing ---

// Step 1 Response: Status and Reference Code
type FlexStatementResponse struct {
	Status        string `xml:"Status"`
	ReferenceCode string `xml:"ReferenceCode"`
	ErrorMessage  string `xml:"ErrorMessage"`
}

// Step 2 Response: The actual data
type FlexQueryResponse struct {
	FlexStatements struct {
		FlexStatement struct {
			Trades struct {
				Trade []Trade `xml:"Trade"`
			} `xml:"Trades"`
		} `xml:"FlexStatement"`
	} `xml:"FlexStatements"`
}

type Trade struct {
	Symbol     string  `xml:"symbol,attr"`
	BuySell    string  `xml:"buySell,attr"`
	Quantity   float64 `xml:"quantity,attr"`
	Price      float64 `xml:"tradePrice,attr"`
	Amount     float64 `xml:"cost,attr"`     // Total Value
	TradeDate  string  `xml:"tradeDate,attr"`  // YYYYMMDD
	TradeID    string  `xml:"tradeID,attr"`
}

func main() {
	// Step 1: Request the Report
	reqURL := fmt.Sprintf("%s/SendRequest?t=%s&q=%s&v=3", BaseURL, FlexToken, QueryID)
	resp, err := http.Get(reqURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var initResp FlexStatementResponse
	if err := xml.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		panic(err)
	}

	if initResp.Status != "Success" {
		fmt.Printf("Error requesting report: %s\n", initResp.ErrorMessage)
		return
	}

	fmt.Printf("Report requested. Reference Code: %s. Waiting 10s...\n", initResp.ReferenceCode)
	
	// Step 2: Wait for generation (mandatory)
	time.Sleep(10 * time.Second)

	// Step 3: Retrieve the Report
	getURL := fmt.Sprintf("%s/GetStatement?q=%s&t=%s&v=3", BaseURL, initResp.ReferenceCode, FlexToken)
	reportResp, err := http.Get(getURL)
	if err != nil {
		panic(err)
	}
	defer reportResp.Body.Close()

	// Parse the actual trade data
	var data FlexQueryResponse
	// Optional: read body to string first if you want to save to file
	// bodyBytes, _ := io.ReadAll(reportResp.Body)
	// os.WriteFile("trades.xml", bodyBytes, 0644)
	
	if err := xml.NewDecoder(reportResp.Body).Decode(&data); err != nil {
		fmt.Printf("Error parsing report (check if report is ready): %v\n", err)
		return
	}

	// Step 4: Output
	trades := data.FlexStatements.FlexStatement.Trades.Trade
	//acctID := data.FlexStatements.FlexStatement.accountId
	fmt.Printf("\nFound %d historical trades:\n", len(trades))
	
	for _, t := range trades {
		fmt.Printf("%s | %s %s | Qty: %.0f | Price: %.2f\n", 
			t.TradeDate, t.BuySell, t.Symbol, t.Quantity, t.Price)
	}
}