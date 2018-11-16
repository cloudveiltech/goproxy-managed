package main

import (
	"C"
	"net/http"
)

var (
	requestMap    = make(map[int]*http.Request)
	lastRequestId = 0
)

func saveRequestToInteropMap(req *http.Request) int {
	requestMap[lastRequestId] = req
	lastRequestId++
	return lastRequestId - 1
}

func removeRequestFromInteropMap(id int) {
	delete(requestMap, id)
}

//export RequestGetUrl
func RequestGetUrl(id int, result *string) {
	request, exists := requestMap[id]
	if !exists {
		return
	}
	*result = request.RequestURI
}

//export RequestHeaderExists
func RequestHeaderExists(id int, name string) bool {
	request, exists := requestMap[id]
	if !exists {
		return false
	}
	_, headerExists := request.Header[name]
	return headerExists
}

//export RequestGetFirstHeader
func RequestGetFirstHeader(id int, name string, res *string) {
	request, exists := requestMap[id]
	if !exists {
		return
	}
	values, headerExists := request.Header[name]
	if !headerExists {
		return
	}
	*res = values[0]
}
