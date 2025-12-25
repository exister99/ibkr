package position

// Structs for XML Parsing
type OpenPosition struct {
	Symbol    string  `xml:"symbol,attr"`
	Position  float64 `xml:"position,attr"`
	MarkPrice float64 `xml:"markPrice,attr"`
	CostBasis float64 `xml:"costBasisPrice,attr"`
	Currency  string  `xml:"currency,attr"`
}

type FlexResult struct {
	Positions []OpenPosition `xml:"FlexStatements>FlexStatement>OpenPositions>OpenPosition"`
}

type FlexAuthResponse struct {
	Status        string `xml:"Status"`
	ReferenceCode string `xml:"ReferenceCode"`
	ErrorCode     string `xml:"ErrorCode"`
	ErrorMessage  string `xml:"ErrorMessage"`
}