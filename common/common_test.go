package common

import (
	"bytes"
	"io"
	"net/http"
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

func TestReadResponseGivenStatusCodeIs200ThenOk(t *testing.T) {
	ok := []byte("ok")
	resp := http.Response{StatusCode: 200, Body: fakeBody{content: ok}}

	buffer, err := ReadResponse(&resp)

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

	buffer, err := ReadResponse(&resp)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
	if !bytes.Equal(buffer, oops) {
		t.Errorf("Expect 'buffer' to equal %v, got %v", oops, buffer)
	}
}
