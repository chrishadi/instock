package main

import (
	"reflect"
	"testing"
)

var a = Stock{Code: "A", LastUpdate: "2020-02-03T00:00:00"}
var b = Stock{Code: "B", LastUpdate: "2020-02-02T00:00:00"}
var c = Stock{Code: "C", LastUpdate: "2020-02-03T00:00:00"}

func TestFilterGivenEmptyDB(t *testing.T) {
	savedStocks := []Stock{}
	newStocks := []Stock{a, b, c}

	expected := FilterResult{
		Active: []Stock{a, b, c},
		New:    []Stock{a, b, c},
	}
	actual, err := filter(newStocks, savedStocks)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("\n%v\nis not equal to\n%v", actual, expected)
	}
}

func TestFilterGivenNewActiveAndStaleStocks(t *testing.T) {
	_a := Stock{Code: "A", LastUpdate: "2020-02-02 00:00:00"}
	_b := Stock{Code: "B", LastUpdate: "2020-02-02 00:00:00"}
	_d := Stock{Code: "D", LastUpdate: "2020-02-02 00:00:00"}

	savedStocks := []Stock{_a, _b, _d}
	newStocks := []Stock{a, b, c}

	expected := FilterResult{
		Active: []Stock{a, c},
		Stale:  []Stock{b},
		New:    []Stock{c},
	}
	actual, err := filter(newStocks, savedStocks)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("\n%v\nis not equal to\n%v", actual, expected)
	}
}
