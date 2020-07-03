package main

import "C"

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"unsafe"

	"github.com/aymerick/raymond"
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

//export AdBlockMatcherBuild
func AdBlockMatcherBuild() {
	if adBlockMatcher != nil {
		adBlockMatcher.Build()
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

	log.Printf("Parsing category %s file %s", categoryId, fileName)

	adBlockMatcher.addMatcher(categoryId, listType == BypassList)

	if listType == TextTrigger {
		adBlockMatcher.addPhrasesFromScanner(scanner, categoryId)
	} else {
		adBlockMatcher.addRulesFromScanner(scanner, categoryId, listType == Whitelist, listType == BypassList)
	}
	return true
}

//export AdBlockMatcherSetBlockedPageContent
func AdBlockMatcherSetBlockedPageContent(contentBlockPageC, contentCertPageC *C.char) {
	blockPagePath := C.GoString(contentBlockPageC)
	adBlockMatcher.BlockPageTemplate = parseTemplate(blockPagePath)

	certPagePath := C.GoString(contentCertPageC)
	adBlockMatcher.BlockCertTemplate = parseTemplate(certPagePath)
}

//export AdBlockMatcherSetBlockPageContextTag
func AdBlockMatcherSetBlockPageContextTag(keyC, valueC *C.char) {
	key := C.GoString(keyC)
	value := C.GoString(valueC)

	if len(value) > 0 {
		adBlockMatcher.defaultBlockPageTags[key] = value
	} else {
		delete(adBlockMatcher.defaultBlockPageTags, key)
	}
}

func parseTemplate(pagePath string) *raymond.Template {
	fileHandle, err := os.Open(pagePath)
	if err != nil {
		log.Printf("Error reading block page %s", err)
		return nil
	}
	defer fileHandle.Close()
	content, e := ioutil.ReadAll(fileHandle)
	if e != nil {
		log.Printf("Error reading block page %s", e)
		return nil
	}

	pageString := string(content)
	template, err := raymond.Parse(pageString)
	if err != nil {
		log.Printf("Error parsing template %s, %v", pagePath, err)
		return nil
	}
	return template
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
	if adBlockMatcher != nil {
		adBlockMatcher.bypassEnabled = true
	}
}

//export AdBlockMatcherDisableBypass
func AdBlockMatcherDisableBypass() {
	if adBlockMatcher != nil {
		adBlockMatcher.bypassEnabled = false
	}
}

//export AdBlockMatcherGetBypassEnabled
func AdBlockMatcherGetBypassEnabled() bool {
	if adBlockMatcher != nil {
		return adBlockMatcher.bypassEnabled
	} else {
		return false
	}
}

//export AdBlockMatcherIsDomainWhitelisted
func AdBlockMatcherIsDomainWhitelisted(hostC *C.char) bool {
	if adBlockMatcher != nil {
		host := C.GoString(hostC)
		return adBlockMatcher.IsDomainWhitelisted(host)
	} else {
		return false
	}
}
