package main

import (
    "crypto/tls"
    "fmt"
    "encoding/json"
    "net/http"
    "net/http/cookiejar"
    "time"
)

// Global client to maintain the session
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

// Position represents a single holding in the portfolio
type Position struct {
    ConId     int     `json:"conid"`       // Contract ID
    AccountID string  `json:"acctId"`      // Account Number
    Symbol    string  `json:"contractDesc"`// Ticker/Description
    Position  float64 `json:"position"`    // Quantity held (can be negative for shorts)
    AvgCost   float64 `json:"avgCost"`     // Average cost per share
    MktValue  float64 `json:"mktValue"`    // Market value of the position
    UnrealizedPnL float64 `json:"unrealizedPnl"` // Unrealized P&L
}

func getPositions(accountID string) ([]Position, error) {
    url := "https://localhost:5000/v1/api/portfolio/" + accountID + "/positions"
    
    // Create the request
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    // Execute the request using the client with the cookie jar
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
    }

    var positions []Position
    // Decode the JSON response into the slice of Position structs
    if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
        return nil, err
    }

    return positions, nil
}

// Remember to import "encoding/json" and "fmt"
func main() {
    // Replace with your actual account ID
    myAccountID := "U2609035" 

    positions, err := getPositions(myAccountID)
    if err != nil {
        fmt.Println("Error retrieving positions:", err)
        return
    }

    fmt.Println("--- Portfolio Positions ---")
    var totalPortfolioValue float64
    var totalUnrealizedPnL float64

    for _, p := range positions {
        fmt.Printf("Symbol: %s | Qty: %.0f | Mkt Value: $%.2f | Unrealized P&L: $%.2f\n", 
            p.Symbol, p.Position, p.MktValue, p.UnrealizedPnL)
        
        totalPortfolioValue += p.MktValue
        totalUnrealizedPnL += p.UnrealizedPnL
    }

    fmt.Println("---------------------------")
    fmt.Printf("Total Portfolio Market Value: $%.2f\n", totalPortfolioValue)
    fmt.Printf("Total Unrealized P&L: $%.2f\n", totalUnrealizedPnL)
}