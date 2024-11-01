package main

import (
	"C"
	"net/http"
	"sync"

	utls "github.com/refraction-networking/utls"
)

type session struct {
	request         *http.Request
	response        *http.Response
	isCertVerified  bool
	connectionState *utls.ConnectionState
}

var (
	mapWriteLock sync.Mutex
	sessionMap   = make(map[int64]*session)
)

func saveSessionToInteropMap(id int64, session *session) int64 {
	mapWriteLock.Lock()
	defer mapWriteLock.Unlock()
	sessionMap[id] = session
	return id
}

func removeSessionFromInteropMap(id int64) {
	mapWriteLock.Lock()
	defer mapWriteLock.Unlock()
	delete(sessionMap, id)
}

func setSessionRequest(id int64, req *http.Request) {
	mapWriteLock.Lock()
	defer mapWriteLock.Unlock()
	session, exists := sessionMap[id]
	if !exists {
		return
	}

	session.request = req
}

func getSessionRequest(id int64) *http.Request {
	mapWriteLock.Lock()
	defer mapWriteLock.Unlock()
	session, exists := sessionMap[id]
	if !exists {
		return nil
	}
	return session.request
}

func getSessionResponse(id int64) *http.Response {
	mapWriteLock.Lock()
	defer mapWriteLock.Unlock()
	session, exists := sessionMap[id]
	if !exists {
		return nil
	}
	return session.response
}

func getSessionConnectionState(id int64) *utls.ConnectionState {
	mapWriteLock.Lock()
	defer mapWriteLock.Unlock()
	session, exists := sessionMap[id]
	if !exists {
		return nil
	}
	return session.connectionState
}

func isSessionTlsVerified(id int64) bool {
	mapWriteLock.Lock()
	defer mapWriteLock.Unlock()
	session, exists := sessionMap[id]
	if !exists {
		return false
	}
	return session.isCertVerified
}
