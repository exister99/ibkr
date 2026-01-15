package main

import (
	"encoding/xml"
	"fmt"
	"math"
	"sort"
	"time"
)

type OpenPosition struct {
	Symbol        string  `xml:"symbol,attr"`
	Quantity      float64 `xml:"position,attr"`
	PurchasePrice float64 `xml:"purchasePrice,attr"`
	PurchaseDate  string  `xml:"dateTime,attr"`
	MarkPrice     float64 `xml:"markPrice,attr"`
}

type FlexResult struct {
	Positions []OpenPosition `xml:"FlexStatements>FlexStatement>OpenPositions>OpenPosition"`
}

func main() {
	// Sample data with different holding periods
	xmlData := []byte(`
	<FlexQueryResult>
		<FlexStatements>
			<FlexStatement>
				<OpenPositions>
					<OpenPosition symbol="AAPL" position="10" purchasePrice="150.00" markPrice="230.00" dateTime="20230101"/>
					<OpenPosition symbol="AAPL" position="10" purchasePrice="210.00" markPrice="230.00" dateTime="20250601"/>
					<OpenPosition symbol="NVDA" position="5" purchasePrice="400.00" markPrice="1200.00" dateTime="20240101"/>
				</OpenPositions>
			</FlexStatement>
		</FlexStatements>
	</FlexQueryResult>`)

	var result FlexResult
	xml.Unmarshal(xmlData, &result)

	grouped := make(map[string][]OpenPosition)
	for _, pos := range result.Positions {
		grouped[pos.Symbol] = append(grouped[pos.Symbol], pos)
	}

	printAnnualizedReturns(grouped)
}

func printAnnualizedReturns(grouped map[string][]OpenPosition) {
	symbols := make([]string, 0, len(grouped))
	for s := range grouped {
		symbols = append(symbols, s)
	}
	sort.Strings(symbols)

	fmt.Printf("%-8s %-12s %-8s %-10s %-12s\n", "Symbol", "Date", "Qty", "Total Gain", "Annualized %")
	fmt.Println("------------------------------------------------------------")

	now := time.Now()

	for _, symbol := range symbols {
		var totalSymbolCost, totalSymbolValue float64
		
		for _, lot := range grouped[symbol] {
			pDate, _ := time.Parse("20060102", lot.PurchaseDate)
			
			// Calculate time held in years (fractional)
			daysHeld := now.Sub(pDate).Hours() / 24
			yearsHeld := daysHeld / 365.25

			initialCost := lot.PurchasePrice * lot.Quantity
			currentValue := lot.MarkPrice * lot.Quantity
			totalGain := (currentValue / initialCost) - 1
			
			// Calculate Annualized Return (CAGR)
			// Handle cases held for less than a day to avoid division by zero/extremes
			annualized := 0.0
			if yearsHeld > 0 && initialCost > 0 {
				annualized = (math.Pow(currentValue/initialCost, 1/yearsHeld) - 1) * 100
			}

			fmt.Printf("%-8s %-12s %-8.1f %-10.1f%% %-12.2f%%\n", 
				symbol, pDate.Format("2006-01-02"), lot.Quantity, totalGain*100, annualized)
			
			totalSymbolCost += initialCost
			totalSymbolValue += currentValue
		}
		
		// Note: Aggregating Annualized Return across multiple lots requires 
		// calculating the Internal Rate of Return (IRR), which is more complex.
		fmt.Println()
	}
}