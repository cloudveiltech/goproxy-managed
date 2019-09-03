package main

import "C"

import (
	"bufio"
	"log"
	"net/http"
	"net/textproto"
	"strings"
	"unsafe"
)

var adBlockMatcher *AdBlockMatcher = nil

var onWhitelistCallback unsafe.Pointer
var onBlacklistCallback unsafe.Pointer

var adBlockMatchers map[int32]*AdBlockMatcher

//export AdBlockMatcherInitialize
func AdBlockMatcherInitialize() {
	var oldMatcher *AdBlockMatcher = nil

	if adBlockMatcher != nil {
		oldMatcher = adBlockMatcher
	}

	adBlockMatcher = CreateMatcher()

	if oldMatcher != nil {
		adBlockMatcher.bypassEnabled = oldMatcher.bypassEnabled
	}
}

//export AdBlockMatcherParseRuleFile
func AdBlockMatcherParseRuleFile(fileName string, categoryId int32, listType int32) {
	log.Printf("AdBlockMatcherParseRuleFile11(%s, %d, %d)", fileName, categoryId, listType)
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

	return adBlockMatcher.TestUrlBlocked(url, host, headers.Get("referer"))
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

//export AdBlockMatcherEnableBypass
func AdBlockMatcherEnableBypass() {
	adBlockMatcher.bypassEnabled = true
}

//export AdBlockMatcherDisableBypass
func AdBlockMatcherDisableBypass() {
	adBlockMatcher.bypassEnabled = false
}

//export AdBlockMatcherGetBypassEnabled
func AdBlockMatcherGetBypassEnabled() bool {
	if adBlockMatcher != nil {
		return adBlockMatcher.bypassEnabled
	} else {
		return false
	}
}
