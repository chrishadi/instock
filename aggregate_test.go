package ingest

import (
	"reflect"
	"testing"
)

var a = Stock{Code: "A", LastUpdate: "2020-02-03T00:00:00", OneDay: 1.0}
var b = Stock{Code: "B", LastUpdate: "2020-02-02T00:00:00", OneDay: 0.5}
var c = Stock{Code: "C", LastUpdate: "2020-02-03T00:00:00", OneDay: 0.0}
var e = Stock{Code: "E", LastUpdate: "2020-02-03T00:00:00", OneDay: -0.5}

func TestAggregateWhenDBIsEmptyShouldReturnAllStocksAsNewAndActive(t *testing.T) {
	lastUpdates := []StockLastUpdate{}
	newStocks := []Stock{a, b, c}

	expected := &AggregateResult{
		Active: []Stock{a, b, c},
		New:    []Stock{a, b, c},
	}
	actual, err := aggregate(newStocks, lastUpdates)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("\n%v\nis not equal to\n%v", actual, expected)
	}
}

func TestAggregateGivenNewActiveAndStaleStocksShouldSplitThem(t *testing.T) {
	alu := StockLastUpdate{Code: "A", LastUpdate: "2020-02-02 00:00:00"}
	blu := StockLastUpdate{Code: "B", LastUpdate: "2020-02-02 00:00:00"}
	dlu := StockLastUpdate{Code: "D", LastUpdate: "2020-02-02 00:00:00"}
	elu := StockLastUpdate{Code: "E", LastUpdate: "2020-02-02 00:00:00"}

	lastUpdates := []StockLastUpdate{alu, blu, dlu, elu}
	newStocks := []Stock{a, b, c, e}

	expected := &AggregateResult{
		Active: []Stock{a, c, e},
		Stale:  []Stock{b},
		New:    []Stock{c},
	}

	actual, err := aggregate(newStocks, lastUpdates)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("\n%v\nis not equal to\n%v", actual, expected)
	}
}
