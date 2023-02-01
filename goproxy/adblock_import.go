package main

import (
	"archive/zip"
	"bufio"
	"io/ioutil"
	"log"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/patriciy/adblock/adblock"
)

func (am *AdBlockMatcher) ParseRulesZipArchive(filePath string) {
	zipFile, e := zip.OpenReader(filePath)
	if e != nil {
		log.Printf("Error parsing zipfile %s", e)
		return
	}
	defer zipFile.Close()
	for _, file := range zipFile.File {
		am.ParseZipRulesFile(file)
	}
}

func (am *AdBlockMatcher) AddRule(rule string, category string, listType int) {
	rule = strings.ToLower(rule)

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
		am.addMatcher(category, listType)
	}

	//Check if it's just a domain rule
	if len(r.Parts) == 2 {
		if r.Parts[0].Type == adblock.DomainAnchor {
			if r.Parts[1].Type == adblock.Exact {
				am.lastCategory.BlockedDomains[string(r.Parts[1].Value)] = !r.Exception
				am.RulesCnt = am.RulesCnt + 1
				return
			}
		}
	}

	am.lastMatcher.AddRule(r, am.RulesCnt)

	am.RulesCnt = am.RulesCnt + 1
}

func (am *AdBlockMatcher) ParseZipRulesFile(file *zip.File) {
	fileDescriptor, err := file.Open()
	defer fileDescriptor.Close()

	if err != nil {
		log.Printf("Error open zip file %s", err)
		return
	}

	if strings.Contains(file.Name, "block.htm") {
		am.addBlockPageFromZipFile(file)
	} else {
		scanner := bufio.NewScanner(fileDescriptor)
		categoryName := file.Name
		if strings.Contains(file.Name, ".triggers") {
			log.Printf("Opening triggers %s", file.Name)
			am.addPhrasesFromScanner(scanner, categoryName)
		} else if strings.Contains(file.Name, ".bypass") {
			am.addMatcher(categoryName, BypassList)
			log.Printf("Opening bypass %s", file.Name)
			am.addRulesFromScanner(scanner, categoryName, BypassList)
		} else if strings.Contains(file.Name, ".rules") {
			am.addMatcher(categoryName, Blacklist)
			log.Printf("Opening rules %s", file.Name)
			am.addRulesFromScanner(scanner, categoryName, Blacklist)
		} else {
			log.Printf("File type recognition failed %s", file.Name)
		}
	}

}

func (am *AdBlockMatcher) addBlockPageFromZipFile(file *zip.File) {
	fileReader, e := file.Open()
	if e != nil {
		log.Printf("Error reading block page %s %s", e, file.Name)
	}
	defer fileReader.Close()
	content, e := ioutil.ReadAll(fileReader)
	if e != nil {
		log.Printf("Error reading block page %s %s", e, file.Name)
	}
	blockPageContent := string(content)
	am.BlockPageTemplate, _ = raymond.Parse(blockPageContent)
}

func (am *AdBlockMatcher) addRulesFromScanner(scanner *bufio.Scanner, categoryName string, listType int) {
	for scanner.Scan() {
		line := scanner.Text()
		if listType == Whitelist && !strings.HasPrefix(line, "@@") {
			line = "@@" + line
		}
		am.AddRule(line, categoryName, listType)
	}
}

func (am *AdBlockMatcher) addPhrasesFromScanner(scanner *bufio.Scanner, categoryName string) {
	for scanner.Scan() {
		line := scanner.Text()
		am.AddBlockedPhrase(line, categoryName)
	}
}
