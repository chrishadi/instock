package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/chrishadi/instock/common"
)

type fakeBody struct {
	content []byte
}

func (m fakeBody) Read(p []byte) (int, error) {
	copy(p, m.content)
	return len(m.content), io.EOF
}

func (m fakeBody) Close() error {
	return nil
}

func TestGetStocksGivenHttpGetReturnProperJson(t *testing.T) {
	jsonb := []byte(`[{"Code":"BBCA","Last":32000.0}]`)
	resp := http.Response{StatusCode: 200, Body: fakeBody{content: jsonb}}
	httpGet := func(url string) (*http.Response, error) {
		return &resp, nil
	}
	expected := []Stock{{Code: "BBCA", Last: 32000.0}}

	actual, err := getStocks("", httpGet, common.ReadResponse, json.Unmarshal)

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, got %v", expected, actual)
	}
}

func TestGetStocksGivenHttpGetReturnError(t *testing.T) {
	httpGet := func(url string) (resp *http.Response, err error) {
		return nil, errors.New("")
	}

	_, err := getStocks("", httpGet, nil, nil)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestGetStocksGivenReadResponseReturnError(t *testing.T) {
	resp := http.Response{StatusCode: 200, Body: fakeBody{content: []byte{}}}
	httpGet := func(url string) (*http.Response, error) {
		return &resp, nil
	}
	readResponse := func(*http.Response) ([]byte, error) {
		return nil, errors.New("")
	}

	_, err := getStocks("", httpGet, readResponse, nil)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestGetStocksGivenJsonUnmarshalReturnError(t *testing.T) {
	resp := http.Response{StatusCode: 200, Body: fakeBody{content: []byte{}}}
	httpGet := func(url string) (*http.Response, error) {
		return &resp, nil
	}
	jsonUnmarshal := func(data []byte, v interface{}) error {
		return errors.New("")
	}

	_, err := getStocks("", httpGet, common.ReadResponse, jsonUnmarshal)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestExtractCode(t *testing.T) {
	expected := []string{"A", "B"}

	actual := extractCode([]Stock{{Code: "A"}, {Code: "B"}})

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, got %v", expected, actual)
	}
}
