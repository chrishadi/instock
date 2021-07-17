package ingest

import (
	"reflect"
	"testing"
)

var a = Stock{Code: "A", LastUpdate: "2020-02-03T00:00:00"}
var b = Stock{Code: "B", LastUpdate: "2020-02-02T00:00:00"}
var c = Stock{Code: "C", LastUpdate: "2020-02-03T00:00:00"}

func TestFilterGivenEmptyDBShouldReturnAllStocksAsNewAndActive(t *testing.T) {
	lastUpdates := []StockLastUpdate{}
	newStocks := []Stock{a, b, c}

	expected := &FilterResult{
		Active: []Stock{a, b, c},
		New:    []Stock{a, b, c},
	}
	actual, err := filter(newStocks, lastUpdates)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("\n%v\nis not equal to\n%v", actual, expected)
	}
}

func TestFilterGivenNewActiveAndStaleStocksShouldSplitThem(t *testing.T) {
	alu := StockLastUpdate{Code: "A", LastUpdate: "2020-02-02 00:00:00"}
	blu := StockLastUpdate{Code: "B", LastUpdate: "2020-02-02 00:00:00"}
	dlu := StockLastUpdate{Code: "D", LastUpdate: "2020-02-02 00:00:00"}

	lastUpdates := []StockLastUpdate{alu, blu, dlu}
	newStocks := []Stock{a, b, c}

	expected := &FilterResult{
		Active: []Stock{a, c},
		Stale:  []Stock{b},
		New:    []Stock{c},
	}
	actual, err := filter(newStocks, lastUpdates)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("\n%v\nis not equal to\n%v", actual, expected)
	}
}
