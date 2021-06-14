package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type HttpGetFn func(string) (*http.Response, error)
type HttpPostFn func(string, string, io.Reader) (*http.Response, error)
type JsonUnmarshalFn func([]byte, interface{}) error
type JsonMarshalFn func(interface{}) ([]byte, error)
type ReadResponseFn func(*http.Response) ([]byte, error)

func ReadResponse(resp *http.Response) ([]byte, error) {
	buffer, err := ioutil.ReadAll(resp.Body)
	if err == nil && resp.StatusCode != 200 {
		err = fmt.Errorf("Status: %d, Body: %s", resp.StatusCode, buffer)
	}

	return buffer, err
}
