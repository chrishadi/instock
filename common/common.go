package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type HttpGetFunc func(string) (*http.Response, error)
type HttpPostFunc func(string, string, io.Reader) (*http.Response, error)

type MockRespBody struct {
	Content []byte
}

func (m MockRespBody) Read(p []byte) (int, error) {
	copy(p, m.Content)
	return len(m.Content), io.EOF
}

func (m MockRespBody) Close() error {
	return nil
}

func ReadResponse(resp *http.Response) ([]byte, error) {
	buffer, err := ioutil.ReadAll(resp.Body)
	if err == nil && resp.StatusCode != 200 {
		err = fmt.Errorf("Status: %d, Body: %s", resp.StatusCode, buffer)
	}

	return buffer, err
}
