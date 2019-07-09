package main

import "log"
import "C"

var adBlockMatcher *AdBlockMatcher

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
func AdBlockMatcherTestUrlMatch(url string, host string) int32 {
	return adBlockMatcher.TestUrlBlocked(url, host)
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