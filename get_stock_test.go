package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"
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

func TestGetStockGivenHttpGetReturnProperJson(t *testing.T) {
	jsonb := []byte(`[{"Code":"BBCA","Last":32000.0}]`)
	resp := http.Response{StatusCode: 200, Body: fakeBody{content: jsonb}}
	httpGet := func(url string) (*http.Response, error) {
		return &resp, nil
	}
	expected := []Stock{{Code: "BBCA", Last: 32000.0}}

	actual, err := getStocks("", httpGet, json.Unmarshal)

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, got %v", expected, actual)
	}
}

func TestGetStockGivenHttpGetReturnError(t *testing.T) {
	httpGet := func(url string) (resp *http.Response, err error) {
		return nil, errors.New("")
	}

	_, err := getStocks("", httpGet, nil)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestGetStockGivenJsonUnmarshalReturnError(t *testing.T) {
	resp := http.Response{StatusCode: 200, Body: fakeBody{content: []byte{}}}
	httpGet := func(url string) (*http.Response, error) {
		return &resp, nil
	}
	jsonUnmarshal := func(data []byte, v interface{}) error {
		return errors.New("")
	}

	_, err := getStocks("", httpGet, jsonUnmarshal)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestReadResponseGivenStatusCodeIs200ThenOk(t *testing.T) {
	ok := []byte("ok")
	resp := http.Response{StatusCode: 200, Body: fakeBody{content: ok}}

	buffer, err := readResponse(&resp)

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
	if !bytes.Equal(buffer, ok) {
		t.Errorf("Expect 'buffer' to equal %v, got: %v", ok, buffer)
	}
}

func TestReadResponseGivenStatusCodeIsNot200ThenError(t *testing.T) {
	oops := []byte("oops")
	resp := http.Response{StatusCode: 504, Body: fakeBody{content: oops}}

	buffer, err := readResponse(&resp)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
	if !bytes.Equal(buffer, oops) {
		t.Errorf("Expect 'buffer' to equal %v, got %v", oops, buffer)
	}
}
