package common

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func ReadResponse(resp *http.Response) ([]byte, error) {
	buffer, err := ioutil.ReadAll(resp.Body)
	if err == nil && resp.StatusCode != 200 {
		err = fmt.Errorf("Status: %d, Body: %s", resp.StatusCode, buffer)
	}

	return buffer, err
}
