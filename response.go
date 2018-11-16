package main

import (
	"C"
	"net/http"
)

var (
	responseMap   = make(map[int]*http.Response)
	lastRequestId = 0
)

func saveResponseToInteropMap(req *http.Response) int {
	responseMap[lastRequestId] = req
	lastRequestId++
	return lastRequestId - 1
}

func removeResponseFromInteropMap(id int) {
	delete(responseMap, id)
}
