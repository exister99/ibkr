package stock

import (
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/gorm"

	fx "skyblaze/ibkr/flexdata/trade"

	p "github.com/exister99/invest/price"
	t "github.com/exister99/invest/transaction"
)

type Stock struct {
	gorm.Model
	Symbol       string `gorm:"column:symbol"`
	Quantity     float64
	Price        float64
	Cost         float64
	Return       float64
	Calls		 float64
	Transactions []t.Transaction `gorm:"foreignKey:StockID"`
}

func NewStock(symbol string, cbp float64, q float64) *Stock {
	// 2. Fetch the current market price
	currentPrice, err := p.Price(symbol)
	if err != nil {
		fmt.Printf("Error fetching price for %s: %v\n", symbol, err)
	}
	return &Stock{
		Symbol:   symbol,
		Quantity: q,
		Price:    currentPrice,
		Cost:     cbp,
		Return:   100 * (currentPrice / cbp),
		Calls:    0,
	}
}

func (s *Stock) PrintTrx() {
	for i, t := range s.Transactions {
		fmt.Printf("The transaction at %d is %v\n", i, t)
	}
}

func (s *Stock) AddTrx(trade *fx.Trade) {
	//fmt.Printf("Adding %v to %v\n", trade, s )
	newTrx := t.NewTransaction(trade)
	s.Transactions = append(s.Transactions, *newTrx)
}

func (stock *Stock) CountCostBasis() float64 {
	total := 0.0
	//shares := 0.0
	for _, t := range stock.Transactions {
		total += t.Amount
		fmt.Printf("Just added %g to the total Amount\n", t.Amount)
	}
	return total
}

func (stock *Stock) CountShares() float64 {
	fmt.Printf("Counting Shares on %d transactions\n", len(stock.Transactions))
	shares := 0.0
	for _, t := range stock.Transactions {
		if t.Action == "BUY" {
			shares += t.Quantity
		}
		if t.Action == "SELL" {
			shares -= t.Quantity
		}
	}
	return shares
}

func (stock *Stock) CountDividends() float64 {
	total := 0.0
	for _, t := range stock.Transactions {
		if strings.HasPrefix(t.Action, "DIVIDEND RECEIVED") {
			total += t.Amount
		}
	}
	//fmt.Printf("The dividend total is %g\n", total )
	return total
}

func (stock *Stock) CountPremiums() float64 {
	total := 0.0
	for _, t := range stock.Transactions {
		if strings.HasPrefix(t.Symbol, "-"+stock.Symbol) {
			total += t.Amount
		}
	}
	//fmt.Printf("The premium total is %g\n", total )
	return total
}

func (stock *Stock) AverageAge() float64 {
	//Test and clean up this function
	now := time.Now()
	totalDuration := int64(0)
	totalShares := 0.0

	const dateFormat = "2006-01-02"
	for _, t := range stock.Transactions {
		if strings.HasPrefix(t.Action, "YOU BOUGHT "+t.Description) {
			fmt.Printf("The purches date for %g shares is %v\n", t.Quantity, t.Date)
			duration := now.Sub(t.Date)
			// 3. Sum all the durations
			durationNanos := duration.Nanoseconds()
			totalDuration += (int64(durationNanos) * int64(t.Quantity))
			totalShares += t.Quantity
		}
	}

	// 1. Convert totalDuration to its nanosecond (int64) value
	nanosPerStandardYear := int64(31536000000000000)
	//nanoseconds := float64(totalDuration.Nanoseconds())
	averageDurationNanos := totalDuration / int64(totalShares)
	return float64(averageDurationNanos) / float64(nanosPerStandardYear)
}

func (stock *Stock) LastCall() (bool, t.Transaction) {
	//Default to a century ago
	lc := time.Now().AddDate(-100, 0, 0)
	foundCall := false
	LastOne := stock.Transactions[0]
	for _, t := range stock.Transactions {
		if strings.HasPrefix(t.Action, "YOU SOLD OPENING TRANSACTION CALL ("+stock.Symbol) {
			if t.Date.After(lc) {
				lc = t.Date
				LastOne = t
				foundCall = true
			}
		}
	}

	return foundCall, LastOne
}

func (stock *Stock) CountFees() float64 {
	total := 0.0
	for _, t := range stock.Transactions {
		total += t.Fee
	}
	//fmt.Printf("The fee total is %g\n", total )
	return total
}
