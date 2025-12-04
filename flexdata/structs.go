package flexdata

// FlexStatementResponse holds the initial response status/reference code
type FlexStatementResponse struct {
	Status        string `xml:"Status"`
	ReferenceCode string `xml:"ReferenceCode"`
	ErrorMessage  string `xml:"ErrorMessage"`
}

// FlexQueryResponse is the root structure for the retrieved report data
type FlexQueryResponse struct {
	FlexStatements struct {
		FlexStatement struct {
			Trades struct {
				Trade []Trade `xml:"Trade"`
			} `xml:"Trades"`
		} `xml:"FlexStatement"`
	} `xml:"FlexStatements"`
}

// Trade holds the details for a single trade transaction
// NOTE: Fields are exported (start with an uppercase letter) to be used outside this package
type Trade struct {
	Symbol     string  `xml:"symbol,attr"`
	Description     string  `xml:"description,attr"`
	BuySell    string  `xml:"buySell,attr"`
	Quantity   float64 `xml:"quantity,attr"`
	Price      float64 `xml:"tradePrice,attr"`
	Amount     float64 `xml:"cost,attr"`
	TradeDate  string  `xml:"tradeDate,attr"`
	TradeID    string  `xml:"tradeID,attr"`
}