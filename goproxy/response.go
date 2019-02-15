package main

import (
	"C"
)
import (
	"bytes"
	"io/ioutil"
	"strings"

	"github.com/cloudveiltech/goproxy"
)

//export ResponseGetStatusCode
func ResponseGetStatusCode(id int64) int {
	response := getSessionResponse(id)
	if response == nil {
		return 0
	}

	return response.StatusCode
}

//export ResponseGetBody
func ResponseGetBody(id int64, res *[]byte) bool {
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

	//since we'd read all body - we need to recreate reader for client here
	response.Body.Close()
	response.Body = ioutil.NopCloser(bytes.NewBuffer(*res))

	return true
}

//export ResponseGetBodyAsString
func ResponseGetBodyAsString(id int64, res *string) bool {
	var bytes []byte
	if !ResponseGetBody(id, &bytes) {
		return false
	}
	*res = string(bytes[:])

	return true
}

//export ResponseHasBody
func ResponseHasBody(id int64) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	return response.Body != nil && response.ContentLength != 0
}

//export ResponseHeaderExists
func ResponseHeaderExists(id int64, name string) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	_, headerExists := response.Header[name]
	return headerExists
}

//export ResponseGetFirstHeader
func ResponseGetFirstHeader(id int64, name string, res *string) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	values, headerExists := response.Header[name]
	if !headerExists {
		return false
	}
	*res = values[0]
	return true
}

//export ResponseSetHeader
func ResponseSetHeader(id int64, name string, value string) bool {
	response := getSessionResponse(id)
	if response == nil {
		return false
	}

	response.Header.Set(name, value)
	return true
}

//export ResponseGetHeaders
func ResponseGetHeaders(id int64, keys *string) int {
	response := getSessionResponse(id)
	if response == nil {
		return 0
	}
	var result strings.Builder
	for key, v := range response.Header {
		for _, value := range v {
			result.WriteString(key + ": " + value + "\r\n")
		}
	}

	*keys = result.String()
	return len(response.Header)
}

//export ResponseGetCertificatesCount
func ResponseGetCertificatesCount(id int64) int {
	response := getSessionResponse(id)
	if response == nil {
		return 0
	}
	if response.TLS == nil {
		return 0
	}
	return len(response.TLS.PeerCertificates)
}

//export ResponseGetCertificate
func ResponseGetCertificate(id int64, index int, certData *[]byte) int {
	response := getSessionResponse(id)
	if response == nil {
		return 0
	}
	if response.TLS == nil {
		return 0
	}

	cert := response.TLS.PeerCertificates[index]

	*certData = cert.Raw

	return 1
}

//export CreateResponse
func CreateResponse(id int64, status int, contentType string, body string) bool {
	session, exists := sessionMap[id]
	if !exists {
		return false
	}

	session.response = goproxy.NewResponse(session.request, contentType, status, body)
	return true
}
