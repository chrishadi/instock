package main

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

func TestReadResponseGivenStatusCodeIs200(t *testing.T) {
	body := fakeBody{content: []byte("ok")}
	resp := http.Response{StatusCode: 200, Body: body}

	buffer, err := readResponse(&resp)

	if err != nil {
		t.Errorf("Expect error to be nil, got: %v", err)
	}
	if !bytes.Equal(buffer, body.content) {
		t.Errorf("Expect 'buffer' to equal %v, got: %v", body.content, buffer)
	}
}

func TestReadResponseGivenStatusCodeIsNot200(t *testing.T) {
	body := fakeBody{content: []byte("oops")}
	resp := http.Response{StatusCode: 504, Body: body}

	buffer, err := readResponse(&resp)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
	if !bytes.Equal(buffer, body.content) {
		t.Errorf("Expect 'buffer' to equal %v, got: %v", body.content, buffer)
	}
}
