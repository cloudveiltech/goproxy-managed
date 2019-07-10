package main

import "C"

import (
	"log"
	"unsafe"
	"net/http"
	"net/textproto"
	"bufio"
	"strings"
)

var adBlockMatcher *AdBlockMatcher

var onWhitelistCallback unsafe.Pointer
var onBlacklistCallback unsafe.Pointer

var adBlockMatchers map[int32]*AdBlockMatcher

//export AdBlockMatcherInitialize
func AdBlockMatcherInitialize() {
	adBlockMatcher = CreateMatcher()
}

//export AdBlockMatcherParseRuleFile
func AdBlockMatcherParseRuleFile(fileName string, categoryId int32, listType int32) {
	log.Printf("AdBlockMatcherParseRuleFile(%s, %d, %d)", fileName, categoryId, listType)
	adBlockMatcher.ParseRuleFile(fileName, categoryId, listType)
}

//export AdBlockMatcherSave
func AdBlockMatcherSave(fileName string) {
	adBlockMatcher.SaveToFile(fileName)
}

//export AdBlockMatcherLoad
func AdBlockMatcherLoad(fileName string) {
	adBlockMatcher = LoadMatcherFromFile(fileName)
}

//export AdBlockMatcherTestUrlMatch
func AdBlockMatcherTestUrlMatch(url string, host string, headersRaw string) []int32 {
	var headers http.Header = nil

	if len(headersRaw) > 0 {
		reader := bufio.NewReader(strings.NewReader(headersRaw + "\r\n"))
		tp := textproto.NewReader(reader)

		mimeHeader, err := tp.ReadMIMEHeader()
		if err != nil {
			log.Printf("MIME Header parse error: %s", err)
		}

		headers = http.Header(mimeHeader)
	}

	return adBlockMatcher.TestUrlBlocked(url, host, headers)
}

//export AdBlockMatcherAreListsLoaded
func AdBlockMatcherAreListsLoaded() bool {
	if adBlockMatcher == nil {
		return false
	} else if adBlockMatcher.MatcherCategories == nil && adBlockMatcher.BypassMatcherCategories == nil {
		return false
	} else {
		return len(adBlockMatcher.MatcherCategories) > 0 || len(adBlockMatcher.BypassMatcherCategories) > 0
	}
}

//export AdBlockMatcherSetWhitelistCallback
func AdBlockMatcherSetWhitelistCallback(callback unsafe.Pointer) {
	onWhitelistCallback = callback
}

//export AdBlockMatcherSetBlacklistCallback
func AdBlockMatcherSetBlacklistCallback(callback unsafe.Pointer) {
	onBlacklistCallback = callback
}