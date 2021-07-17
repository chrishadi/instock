package main

type Stock struct {
	Name              string  `json:"Name"`
	Code              string  `json:"Code"`
	SubSectorId       uint    `json:"StockSubSectorId"`
	SubSectorName     string  `json:"SubSectorName"`
	SectorId          uint    `json:"StockSectorId"`
	SectorName        string  `json:"SectorName"`
	Last              float32 `json:"Last"`
	PrevClosingPrice  float32 `json:"PrevClosingPrice"`
	AdjustedOpenPrice float32 `json:"AdjustedOpenPrice"`
	AdjustedHighPrice float32 `json:"AdjustedHighPrice"`
	AdjustedLowPrice  float32 `json:"AdjustedLowPrice"`
	Volume            float64 `json:"Volume"`
	Frequency         float64 `json:"Frequency"`
	Value             float64 `json:"Value"`
	LastUpdate        string  `json:"LastUpdate"`
}

type StockLastUpdate struct {
	Code       string `json:"Code"`
	LastUpdate string `json:"LastUpdate"`
}
