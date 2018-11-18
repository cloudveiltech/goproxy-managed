package main

import (
	"C"
	"bytes"
)

//export RequestGetUrl
func RequestGetUrl(id int, result *string) {
	request := getSessionRequest(id)
	if request == nil {
		return
	}
	*result = request.RequestURI
}

//export RequestGetBody
func RequestGetBody(id int, res *[]byte) bool {
	request := getSessionRequest(id)
	if request == nil {
		return false
	}

	if request.Body == nil {
		return false
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)

	*res = buf.Bytes()
	return true
}

//export RequestHasBody
func RequestHasBody(id int) bool {
	request := getSessionRequest(id)
	if request == nil {
		return false
	}

	return request.Body != nil && request.ContentLength > 0
}

//export RequestHeaderExists
func RequestHeaderExists(id int, name string) bool {
	request := getSessionRequest(id)
	if request == nil {
		return false
	}

	// for k := range request.Header {
	// 	fmt.Fprintf(os.Stderr, "key[%s] value[%s]\n", k, request.Header[k])
	// }

	_, headerExists := request.Header[name]
	return headerExists
}

//export RequestGetFirstHeader
func RequestGetFirstHeader(id int, name string, res *string) {
	request := getSessionRequest(id)
	if request == nil {
		return false
	}

	values, headerExists := request.Header[name]
	if !headerExists {
		return
	}
	*res = values[0]
}

//export RequestSetHeader
func RequestSetHeader(id int, name string, value string) {
	request := getSessionRequest(id)
	if request == nil {
		return false
	}

	request.Header.Set(name, value)
}
