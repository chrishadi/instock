package ingest

import (
	"container/list"
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
		Active:     []Stock{a, c, e},
		Stale:      []Stock{b},
		New:        []Stock{c},
		TopGainers: []string{"A"},
		TopLosers:  []string{"E"},
	}

	actual, err := aggregate(newStocks, lastUpdates)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("\n%v\nis not equal to\n%v", actual, expected)
	}
}

func TestUpdateTopRankGivenEmptyListAndNewStockShouldInsertTheStock(t *testing.T) {
	ls := list.New()

	updateTopRank(ls, Stock{Code: "A", OneDay: 1.0}, func(a, b float64) bool { return a > b })

	if ls.Len() != 1 {
		t.Error("Expect list length to be 1, got", ls.Len())
	}
	first := ls.Front().Value.(StockGain)
	if first.Code != "A" {
		t.Error("Expect A as the first element, got", first)
	}
}

func TestUpdateTopRankGivenNewStockGainIsGreaterThanTopShouldInsertAtTheTop(t *testing.T) {
	ls := list.New()
	ls.PushBack(StockGain{"A", 3.0})
	ls.PushBack(StockGain{"B", 2.0})

	updateTopRank(ls, Stock{Code: "C", OneDay: 5.0}, func(a, b float64) bool { return a > b })

	top := ls.Front().Value.(StockGain)
	if top.Code != "C" {
		t.Error("Expect C at the top, got", top)
	}
}

func TestUpdateTopRankGivenNewStockGainIsSomewhereBetweenShouldInsertSomewhereBetween(t *testing.T) {
	ls := list.New()
	ls.PushBack(StockGain{"A", -0.3})
	ls.PushBack(StockGain{"B", -0.2})
	ls.PushBack(StockGain{"C", -0.1})

	updateTopRank(ls, Stock{Code: "D", OneDay: -0.25}, func(a, b float64) bool { return a < b })

	second := ls.Front().Next().Value.(StockGain)
	if second.Code != "D" {
		t.Error("Expect D at second position, got", second)
	}
}

func TestUpdateTopRankGivenNewStockGainIsLessThanBottomShouldInsertAtTheBottom(t *testing.T) {
	ls := list.New()
	ls.PushBack(StockGain{"A", 3.0})
	ls.PushBack(StockGain{"B", 2.0})

	updateTopRank(ls, Stock{Code: "C", OneDay: 1.0}, func(a, b float64) bool { return a > b })

	bottom := ls.Back().Value.(StockGain)
	if bottom.Code != "C" {
		t.Error("Expect C at the bottom, got", bottom)
	}
}

func TestUpdateTopRankGivenTheListIsFullAndNewStockGainIsGreaterThanSomeShouldRemoveLast(t *testing.T) {
	_numOfGL := numOfGL
	numOfGL = 5
	ls := list.New()
	ls.PushBack(StockGain{"A", 3.0})
	ls.PushBack(StockGain{"B", 2.7})
	ls.PushBack(StockGain{"C", 2.0})
	ls.PushBack(StockGain{"D", 1.3})
	ls.PushBack(StockGain{"E", 1.0})

	updateTopRank(ls, Stock{Code: "F", OneDay: 1.5}, func(a, b float64) bool { return a > b })

	if ls.Len() != numOfGL {
		t.Error("Expect list len to be 5, got", ls.Len())
	}
	bottom := ls.Back().Value.(StockGain)
	if bottom.Code != "D" {
		t.Error("Expect D at the bottom, got", bottom)
	}

	numOfGL = _numOfGL
}

func TestExtractTopRankCodesGivenListOfStockGainShouldReturnStockCodes(t *testing.T) {
	ls := list.New()
	ls.PushBack(StockGain{"A", 3.0})
	ls.PushBack(StockGain{"B", 2.0})
	ls.PushBack(StockGain{"C", 1.0})
	expected := []string{"A", "B", "C"}

	actual := extractTopRankCodes(ls)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, got %v", expected, actual)
	}
}
