package common

import (
	"io"
	"net/http"
	"testing"
)

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

func TestReadResponseGivenStatusCodeIsNot200ShouldReturnError(t *testing.T) {
	oops := []byte("oops")
	resp := http.Response{StatusCode: 504, Body: MockRespBody{Content: oops}}

	buf, err := ReadResponse(&resp)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
	if string(buf) != string(oops) {
		t.Errorf("Expect 'buf' to equal %s, got %s", oops, buf)
	}
}

func TestReadResponseGivenStatusCodeIs200ShouldReturnResponseBody(t *testing.T) {
	ok := []byte("ok")
	resp := http.Response{StatusCode: 200, Body: MockRespBody{Content: ok}}

	buf, err := ReadResponse(&resp)

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
	if string(buf) != string(ok) {
		t.Errorf("Expect 'buf' to equal %s, got: %s", ok, buf)
	}
}
