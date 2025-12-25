package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	// Import Koanf libraries for config loading
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

)

// Constants - Replace with your actual credentials
const (
	//Token   = "YOUR_IBKR_TOKEN"
	QueryID = "1356519"
	BaseURL = "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService"
)

// Response from SendRequest
type FlexAuthResponse struct {
	XMLName       xml.Name `xml:"FlexStatementResponse"`
	Status        string   `xml:"Status"`
	ReferenceCode string   `xml:"ReferenceCode"`
	Url           string   `xml:"Url"`
	ErrorCode     string   `xml:"ErrorCode"`
	ErrorMessage  string   `xml:"ErrorMessage"`
}

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
	// 1. Load the FlexToken from the TOML file
	if err := loadConfig(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	fmt.Printf("The token is %s\n", Token)

	// 2. Request the report generation
	refCode, err := requestReport()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Report requested successfully. Reference Code: %s\n", refCode)
	fmt.Println("Waiting for IBKR to generate the file...")
	
	// IBKR often takes a few seconds to process the request
	time.Sleep(5 * time.Second)

	// 3. Fetch the actual statement data
	data, err := getStatement(refCode)
	if err != nil {
		panic(err)
	}

	fmt.Println("--- Query Results ---")
	fmt.Println(string(data))
}

func requestReport() (string, error) {
	url := fmt.Sprintf("%s/SendRequest?t=%s&q=%s&v=3", BaseURL, Token, QueryID)
	
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var auth FlexAuthResponse
	if err := xml.Unmarshal(body, &auth); err != nil {
		return "", err
	}

	if auth.Status == "Fail" {
		return "", fmt.Errorf("IBKR Error %s: %s", auth.ErrorCode, auth.ErrorMessage)
	}

	return auth.ReferenceCode, nil
}

func getStatement(refCode string) ([]byte, error) {
	url := fmt.Sprintf("%s/GetStatement?t=%s&q=%s&v=3", BaseURL, Token, refCode)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}