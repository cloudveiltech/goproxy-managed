package main

import (
	//"archive/zip"
	"bufio"
	//"io/ioutil"
	"log"
	//"strings"
	"os"

	"github.com/pmezard/adblock/adblock"
)

const (
	Blacklist = 1
	Whitelist = 2
	BypassList = 3
)

/*func (am *AdBlockMatcher) ParseRulesZipArchive(filePath string) {
	zipFile, e := zip.OpenReader(filePath)
	if e != nil {
		log.Printf("Error parsing zipfile %s", e)
		return
	}
	defer zipFile.Close()
	for _, file := range zipFile.File {
		am.ParseZipRulesFile(file)
	}
}*/

func (am *AdBlockMatcher) AddRule(rule string, categoryId int32, listType int32) {
	bypass := listType == BypassList

	r, e := adblock.ParseRule(rule)

	if e != nil {
		log.Printf("Error parsing rule %s %s", rule, e)
		return
	}
	if r == nil {
		log.Printf("Error parsing rule is nil")
		return
	}

	if am.RulesCnt%MAX_RULES_PER_MATCHER == 0 {
		am.addMatcher(categoryId, listType, bypass)
	}

	am.lastMatcher.AddRule(r, am.RulesCnt)

	am.RulesCnt = am.RulesCnt + 1
}

func (am *AdBlockMatcher) ParseRuleFile(fileName string, categoryId int32, listType int32) {
	file, err := os.Open(fileName)
	defer file.Close()

	if err != nil {
		log.Printf("Error opening rule file %s with error %s", fileName, err)
		return
	}

	scanner := bufio.NewScanner(file)

	bypass := listType == BypassList

	am.addMatcher(categoryId, listType, bypass)
	log.Printf("Opening rules %s", fileName)
	am.addRulesFromScanner(scanner, categoryId, listType)
}

func (am *AdBlockMatcher) addRulesFromScanner(scanner *bufio.Scanner, categoryId int32, listType int32) {
	for scanner.Scan() {
		line := scanner.Text()

		am.AddRule(line, categoryId, listType)
	}
}

/*func (am *AdBlockMatcher) addPhrasesFromScanner(scanner *bufio.Scanner, categoryId int32) {
	for scanner.Scan() {
		line := scanner.Text()
		am.AddBlockedPhrase(line, categoryId)
	}
}*/
