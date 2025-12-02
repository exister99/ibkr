package stock

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"

	t "github.com/exister99/invest/transaction"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Stock struct {
	gorm.Model
	Symbol       string          `gorm:"column:symbol"`
	Transactions []t.Transaction `gorm:"foreignKey:StockID"`
}

func NewStock(symbol string) (Stock) {
	return Stock{
		Symbol:       symbol,
	}
}

func GetTrxs(s string) ([]t.Transaction, error) {
	var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	dsn, tt, err := getConnTrx()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	defer db.Close()

    likePattern := "-" + s + "%"

	fmt.Printf("The symbol is %s\n", s)
	// Build the SELECT query
	rows, err := psql.Select("transaction_date", "action", "symbol", "description", "quantity", "price", "fees_and_commission", "amount").
		From(tt).
		Where(sq.Or{
			// Condition 1: The title contains the search term (using the LIKE logic from Query 1)
			sq.Expr("symbol LIKE ?", likePattern), 
			// Condition 2: The author is exactly "John Doe" (using sq.Eq for equality)
			sq.Eq{"symbol": s},                
		}).
		RunWith(db).
		Query()

	defer rows.Close()

	if err != nil {
		log.Fatal(err)
	}

	// 4. Iterate through the rows
	var trxs []t.Transaction
	for rows.Next() {
		var trx t.Transaction
		// Use rows.Scan() to read the data from the current row into the struct fields
		if err := rows.Scan(&trx.Date, &trx.Action, &trx.Symbol, &trx.Description, &trx.Quantity, &trx.Price, &trx.Fee, &trx.Amount); err != nil {
			// Handle error and continue or break
			log.Println("Error scanning row:", err)
			continue
		}
		trxs = append(trxs, trx)
	}

	return trxs, nil
}

func GetSymbols() ([]string, error) {
	dsn, pt, err := getConnStr()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	defer db.Close()

	// 1. Use Squirrel to build the SELECT query
	queryBuilder := sq.Select("symbol").
		Distinct().
		From(pt)

	// 2. Get the raw SQL and arguments
	sqlString, args, err := queryBuilder.ToSql()
	if err != nil {
		// Handle error
		log.Fatal(err)
	}

	// 3. Execute the query using the standard library
	rows, err := db.Query(sqlString, args...)
	if err != nil {
		// Handle error
		log.Fatal(err)
	}
	defer rows.Close() // ALWAYS close the rows object when you're done

	var symbols []string

	// 4. Iterate through the rows
	for rows.Next() {
		var s string
		// Use rows.Scan() to read the data from the current row into the struct fields
		if err := rows.Scan(&s); err != nil {
			// Handle error and continue or break
			log.Println("Error scanning row:", err)
			continue
		}
		symbols = append(symbols, s)
	}

	return symbols, nil

}

func getConnStr() (string, string, error) {
	var k = koanf.New(".")

	c := "./config.toml"
	if err := k.Load(file.Provider(c), toml.Parser()); err != nil {
		log.Fatalf("error loading file: %v", err)
	}

	dsn := k.String("pg.dsn")
	pt := k.String("pg.pt")
	return dsn, pt, nil
}

func getConnTrx() (string, string, error) {
	var k = koanf.New(".")

	c := "./config.toml"
	if err := k.Load(file.Provider(c), toml.Parser()); err != nil {
		log.Fatalf("error loading file: %v", err)
	}

	dsn := k.String("pg.dsn")
	tt := k.String("pg.tt")
	return dsn, tt, nil
}

func (old *Stock) AddTrxs(new *Stock) {
	fmt.Printf("Adding %v to %v\n", new, old )
}

func (stock *Stock) CountCostBasis() float64 {
	total := 0.0
	shares := 0.0
	for _, t := range stock.Transactions {
			if strings.HasPrefix(t.Action, "YOU BOUGHT " + t.Description) {
			total += t.Amount
			shares += t.Quantity
			fmt.Printf("Just added %g to the total\n", t.Amount )
		}
	}
	return -(total/shares)
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
		if strings.HasPrefix(t.Symbol, "-" + stock.Symbol) {
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
		if strings.HasPrefix(t.Action, "YOU BOUGHT " + t.Description) {
			fmt.Printf("The purches date for %g shares is %v\n", t.Quantity, t.Date )
			duration := now.Sub(t.Date)
			// 3. Sum all the durations
			durationNanos := duration.Nanoseconds() 
			totalDuration += (int64(durationNanos)  * int64(t.Quantity))
			totalShares += t.Quantity
		}
	}

	// 1. Convert totalDuration to its nanosecond (int64) value
	nanosPerStandardYear := int64(31536000000000000)
	//nanoseconds := float64(totalDuration.Nanoseconds())
	averageDurationNanos := totalDuration / int64(totalShares) 
	return float64(averageDurationNanos)/float64(nanosPerStandardYear)
}

func (stock *Stock) LastCall() (bool, t.Transaction) {
	//Default to a century ago
    lc := time.Now().AddDate(-100,0,0)
    foundCall := false
    LastOne := stock.Transactions[0]
	for _, t := range stock.Transactions {
		if strings.HasPrefix(t.Action, "YOU SOLD OPENING TRANSACTION CALL (" + stock.Symbol) {
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