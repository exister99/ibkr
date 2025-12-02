package transaction

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"regexp"
	"strconv"
	"time"
)

type Transaction struct {
	gorm.Model
	Id          int
	Date        time.Time `gorm:"column:date"`
	Symbol      string    `gorm:"column:symbol"`
	Action      string    `gorm:"column:action"`
	Description string    `gorm:"column:description"`
	Quantity    float64   `gorm:"column:quantity"`
	Price       float64   `gorm:"column:price"`
	Amount      float64   `gorm:"column:amount"`
	Fee         float64   `gorm:"column:fee"`
	StockID     uint
}

func NewTransaction(symbol string, action string, description string, quantity float64, price float64, amount float64) *Transaction {
	return &Transaction{
		Symbol:      symbol,
		Action:      action,
		Description: description,
		Quantity:    quantity,
		Price:       price,
		Amount:      amount,
	}
}

func (t *Transaction) Strike() (float64, error) {
	// The regex pattern matches one or more digits (\d+) at the end of the string ($)
	re := regexp.MustCompile(`(\d+)$`)

	match := re.FindStringSubmatch(t.Symbol)

	if len(match) > 1 {
		// match[1] is the captured group (the number itself)
		numberStr := match[1]
		// Convert the matched string to an integer
		number, err := strconv.Atoi(numberStr)
		if err != nil {
			return 0, fmt.Errorf("could not convert number to integer: %w", err)
		}
		return float64(number), nil
	}

	return 0, fmt.Errorf("no number found at the end of the string: %s", t.Symbol)
}

func (t *Transaction) Expiration() (time.Time, error) {
	// The regex pattern matches:
	// - a hyphen (-)
	// - any non-digit characters (.*?)
	// - exactly six digits ((\d{6})) - this is our date, captured in group 1
	re := regexp.MustCompile(`-.*?(\d{6})`)

	match := re.FindStringSubmatch(t.Symbol)

	if len(match) > 1 {
		// match[1] is the captured date string (e.g., "260618")
		dateStr := match[1]

		// The expected layout is YYMMDD, which corresponds to "060102"
		// Go's time package uses the specific reference date: Mon Jan 2 15:04:05 MST 2006
		// Year (YY): "06"
		// Month (MM): "01"
		// Day (DD): "02"
		const layout = "060102"

		// Parse the date string into a time.Time object
		t, err := time.Parse(layout, dateStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("could not parse date string '%s' with layout '%s': %w", dateStr, layout, err)
		}
		return t, nil
	}

	return time.Time{}, fmt.Errorf("no YYMMDD date found in the string: %s", t.Symbol)
}

func (t *Transaction) ExpirationNanos() float64  {
	exp, err := t.Expiration()
	if err != nil {
		log.Fatalf("Error fetching last call expiration: %v", err)
	}
    duration := exp.Sub(time.Now())
	// 3. Sum all the durations

	// 1. Convert totalDuration to its nanosecond (int64) value
	nanosPerStandardYear := int64(31536000000000000)
	
	return float64(duration.Nanoseconds())/float64(nanosPerStandardYear)
}