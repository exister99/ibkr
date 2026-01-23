package position

// Structs for XML Parsing
type OpenPosition struct {
	AssetCategory     string  `xml:"assetCategory,attr"`
	UnderlyingSymbol	string `xml:"underlyingSymbol,attr"`
	Symbol    string  `xml:"symbol,attr"`
	Position  float64 `xml:"position,attr"`
	MarkPrice float64 `xml:"markPrice,attr"`
	CostBasisPrice float64 `xml:"costBasisPrice,attr"`
	CostBasisMoney float64 `xml:"costBasisMoney,attr"`
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