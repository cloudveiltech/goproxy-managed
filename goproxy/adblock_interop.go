package main

import "C"

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/aymerick/raymond"
)

var adBlockBlacklistCallback unsafe.Pointer

var adBlockMatchers map[int32]*AdBlockMatcher
var adBlockInteropSyncMutex sync.Mutex

var newAdBlockMatcher *AdBlockMatcher

//export AdBlockMatcherInitialize
func AdBlockMatcherInitialize() {
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	newAdBlockMatcher = CreateMatcher()

	if adBlockMatcher != nil {
		newAdBlockMatcher.bypassEnabled = adBlockMatcher.bypassEnabled
	}
}

//export AdBlockMatcherBuild
func AdBlockMatcherBuild() {
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	if newAdBlockMatcher != nil {
		newAdBlockMatcher.Build()

		adBlockMatcher = newAdBlockMatcher
		newAdBlockMatcher = nil
	}
}

//export AdBlockMatcherParseRuleFile
func AdBlockMatcherParseRuleFile(fileNameC *C.char, categoryIdC *C.char, listType int32) bool {
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	categoryId := C.GoString(categoryIdC)
	fileName := C.GoString(fileNameC)

	fileHandle, err := os.Open(fileName)
	if err != nil {
		return false
	}
	defer fileHandle.Close()

	if newAdBlockMatcher == nil {
		return false
	}

	scanner := bufio.NewScanner(fileHandle)
	log.Printf("Parsing category %s file %s", categoryId, fileName)

	newAdBlockMatcher.addMatcher(categoryId, int(listType))

	if listType == TextTrigger {
		newAdBlockMatcher.addPhrasesFromScanner(scanner, categoryId)
	} else {
		newAdBlockMatcher.addRulesFromScanner(scanner, categoryId, int(listType))
	}
	time.Sleep(time.Millisecond * 10)
	return true
}

//export AdBlockMatcherSetBlockedPageContent
func AdBlockMatcherSetBlockedPageContent(contentBlockPageC, contentCertPageC *C.char) {
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	if newAdBlockMatcher != nil {
		blockPagePath := C.GoString(contentBlockPageC)
		newAdBlockMatcher.BlockPageTemplate = parseTemplate(blockPagePath)

		certPagePath := C.GoString(contentCertPageC)
		newAdBlockMatcher.BlockCertTemplate = parseTemplate(certPagePath)
	}
}

//export AdBlockMatcherSetBlockPageContextTag
func AdBlockMatcherSetBlockPageContextTag(keyC, valueC *C.char) {
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	if newAdBlockMatcher != nil {
		key := C.GoString(keyC)
		value := C.GoString(valueC)

		if len(value) > 0 {
			newAdBlockMatcher.defaultBlockPageTags[key] = value
		} else {
			delete(newAdBlockMatcher.defaultBlockPageTags, key)
		}
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
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	adBlockMatcher.SaveToFile(fileName)
}

//export AdBlockMatcherLoad
func AdBlockMatcherLoad(fileName string) {
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	adBlockMatcher = LoadMatcherFromFile(fileName)
}

//export AdBlockMatcherEnableBypass
func AdBlockMatcherEnableBypass() {
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	if adBlockMatcher != nil {
		adBlockMatcher.bypassEnabled = true
	}
}

//export AdBlockMatcherDisableBypass
func AdBlockMatcherDisableBypass() {
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	if adBlockMatcher != nil {
		adBlockMatcher.bypassEnabled = false
	}
}

//export AdBlockMatcherGetBypassEnabled
func AdBlockMatcherGetBypassEnabled() bool {
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	if adBlockMatcher != nil {
		return adBlockMatcher.bypassEnabled
	} else {
		return false
	}
}

//export AdBlockMatcherIsDomainWhitelisted
func AdBlockMatcherIsDomainWhitelisted(hostC *C.char) bool {
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	if adBlockMatcher != nil {
		host := C.GoString(hostC)
		return adBlockMatcher.IsDomainWhitelisted(host)
	} else {
		return false
	}
}

//export AdBlockMatcherGetWhitelistedDomains
func AdBlockMatcherGetWhitelistedDomains() *C.char {
	adBlockInteropSyncMutex.Lock()
	defer adBlockInteropSyncMutex.Unlock()

	if adBlockMatcher != nil {
		domains := adBlockMatcher.GetWhitelistedDomains()
		res := strings.Join(domains, ";")
		return C.CString(res)
	} else {
		return C.CString("")
	}
}
