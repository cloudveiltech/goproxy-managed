package main

import "C"

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"unsafe"
)

var adBlockBlacklistCallback unsafe.Pointer

var adBlockMatchers map[int32]*AdBlockMatcher

const (
	Blacklist   = 1
	Whitelist   = 2
	BypassList  = 3
	TextTrigger = 4
)

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
func AdBlockMatcherParseRuleFile(fileNameC *C.char, categoryIdC *C.char, listType int32) bool {
	fileName := C.GoString(fileNameC)
	categoryId := C.GoString(categoryIdC)

	fileHandle, err := os.Open(fileName)
	if err != nil {
		return false
	}
	defer fileHandle.Close()

	scanner := bufio.NewScanner(fileHandle)

	if listType == TextTrigger {
		adBlockMatcher.addPhrasesFromScanner(scanner, categoryId)
	} else {
		adBlockMatcher.addRulesFromScanner(scanner, categoryId, listType == Whitelist, listType == BypassList)
	}
	return true
}

//export AdBlockMatcherSetBlockedPageContent
func AdBlockMatcherSetBlockedPageContent(contentC *C.char) {
	blockPagePath := C.GoString(contentC)
	fileHandle, err := os.Open(blockPagePath)
	if err != nil {
		log.Printf("Error reading block page %s", err)
		return
	}
	defer fileHandle.Close()
	content, e := ioutil.ReadAll(fileHandle)
	if e != nil {
		log.Printf("Error reading block page %s", e)
		return
	}
	adBlockMatcher.BlockPageContent = string(content)
}

//export AdBlockMatcherSave
func AdBlockMatcherSave(fileName string) {
	adBlockMatcher.SaveToFile(fileName)
}

//export AdBlockMatcherLoad
func AdBlockMatcherLoad(fileName string) {
	adBlockMatcher = LoadMatcherFromFile(fileName)
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
