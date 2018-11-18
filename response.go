package main

import (
	"C"
)
import "bytes"

//export ResponseGetStatusCode
func ResponseGetStatusCode(id int) int {
	response := getSessionResponse(id)
	if response == nil {
		return 0
	}

	return response.StatusCode
}

//export ResponseGetBody
func ResponseGetBody(id int, res *[]byte) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	if response.Body == nil {
		return false
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	*res = buf.Bytes()
	return true
}

//export ResponseHasBody
func ResponseHasBody(id int) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	return response.Body != nil && response.ContentLength > 0
}

//export ResponseHeaderExists
func ResponseHeaderExists(id int, name string) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	_, headerExists := response.Header[name]
	return headerExists
}

//export ResponseGetFirstHeader
func ResponseGetFirstHeader(id int, name string, res *string) {
	response := getSessionResponse(id)
	if response == nil {
		return
	}

	values, headerExists := response.Header[name]
	if !headerExists {
		return
	}
	*res = values[0]
}

//export ResponseSetHeader
func ResponseSetHeader(id int, name string, value string) {
	response := getSessionResponse(id)
	if response == nil {
		return
	}

	response.Header.Set(name, value)
}
