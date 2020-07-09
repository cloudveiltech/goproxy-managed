package main

import (
	"compress/gzip"
	"encoding/gob"
	"log"
	"net/url"
	"os"
	"runtime/debug"
	"strings"

	"github.com/patriciy/adblock/adblock"

	"encoding/base64"

	"unicode"

	goahocorasick "github.com/anknown/ahocorasick"
	"github.com/aymerick/raymond"
)

const (
	Included = adblock.Included
	Excluded = adblock.Excluded
)

const MAX_RULES_PER_MATCHER = 1000
const MAX_CONTENT_SIZE_SCAN = 1000 * 1024 //500kb max to scan
var adBlockMatcher *AdBlockMatcher

var defaultBlockPageContent = "{{url_text}} is blocked. Category {{matching_category}}. Reason {{message}}"

type cacheItem struct {
	category        *string
	matchType       int
	isRelaxedPolicy bool
}

type MatcherCategory struct {
	Category       string
	Matchers       []*adblock.RuleMatcher
	BlockedDomains map[string]bool
}

type PhraseCategory struct {
	Category  string
	Phrases   []string
	processor *goahocorasick.Machine
}

type AdBlockMatcher struct {
	MatcherCategories       []*MatcherCategory
	BypassMatcherCategories []*MatcherCategory
	PhraseCategories        []*PhraseCategory
	lastMatcher             *adblock.RuleMatcher
	lastCategory            *MatcherCategory
	RulesCnt                int
	phrasesCount            int
	bypassEnabled           bool
	BlockPageTemplate       *raymond.Template
	BlockCertTemplate       *raymond.Template
	defaultBlockPageTags    map[string]string
}

func CreateMatcher() *AdBlockMatcher {
	adBlockMatcher = &AdBlockMatcher{
		RulesCnt:             0,
		defaultBlockPageTags: make(map[string]string),
	}

	return adBlockMatcher
}

func (am *AdBlockMatcher) addMatcher(category string, bypass bool) {
	matcher := adblock.NewMatcher()
	var categoryMatcher *MatcherCategory
	for _, element := range adBlockMatcher.MatcherCategories {
		if element.Category == category {
			categoryMatcher = element
			break
		}
	}

	if categoryMatcher == nil {
		categoryMatcher = &MatcherCategory{
			Category:       category,
			BlockedDomains: make(map[string]bool),
		}

		if bypass {
			am.BypassMatcherCategories = append(am.BypassMatcherCategories, categoryMatcher)
		} else {
			am.MatcherCategories = append(am.MatcherCategories, categoryMatcher)
		}
	}

	categoryMatcher.Matchers = append(categoryMatcher.Matchers, matcher)
	adBlockMatcher.lastMatcher = matcher
	adBlockMatcher.lastCategory = categoryMatcher
}

func (am *AdBlockMatcher) GetBlockPage(blockedUrl, category string, isRelaxedPolicy bool) string {
	tags := am.defaultBlockPageTags

	tags["url_text"] = blockedUrl
	tags["friendly_url_text"] = blockedUrl
	tags["message"] = ""
	tags["matching_category"] = category
	if isRelaxedPolicy {
		tags["isRelaxedPolicy"] = "1"
	} else {
		tags["isRelaxedPolicy"] = ""
	}
	tags["showUnblockRequestButton"] = "1"

	tags["unblockRequest"] = tags["unblockRequestBase"] + "&category_name=" + url.QueryEscape(category) + "&blocked_request=" + base64.StdEncoding.EncodeToString([]byte(blockedUrl))

	res, err := am.BlockPageTemplate.Exec(tags)
	if err != nil {
		log.Printf("Error render block block page %v", err)
		return "Blocked default page"
	}
	return res
}

func (am *AdBlockMatcher) GetBadCertPage(blockedUrl, host, certThumbPrint string) string {
	tags := am.defaultBlockPageTags
	if len(certThumbPrint) > 0 {
		tags["certThumbprintExists"] = "1"
	}

	tags["url_text"] = blockedUrl
	tags["friendly_url_text"] = blockedUrl
	tags["certThumbprint"] = certThumbPrint
	tags["host"] = host

	if am.BlockCertTemplate == nil {
		return "Blocked cert default page"
	}
	res, err := am.BlockCertTemplate.Exec(tags)
	if err != nil {
		log.Printf("Error render block cert page %v", err)
		return "Blocked cert default page"
	}
	return res
}

func (am *AdBlockMatcher) IsDomainWhitelisted(host string) bool {
	category, matchType, _ := am.TestUrlBlocked("https://"+host, host, "")
	if category != nil && matchType == Excluded {
		log.Printf("Testing early host - true %s", host)
		return true
	}

	log.Printf("Testing early host - false %s", host)
	return false
}

func (am *AdBlockMatcher) TestUrlBlocked(url string, host string, referer string) (category *string, matchType int, isRelaxedPolicy bool) {
	if am.RulesCnt == 0 {
		return nil, Included, false
	}

	res1, res2 := am.matchRulesCategories(am.MatcherCategories, url, host, referer)
	if res1 != nil {
		return res1, res2, false
	}

	if am.bypassEnabled {
		return nil, Included, true
	}

	res1, res2 = am.matchRulesCategories(am.BypassMatcherCategories, url, host, referer)

	return res1, res2, true
}

func (am *AdBlockMatcher) matchRulesCategories(matcherCategories []*MatcherCategory, url string, host string, referer string) (*string, int) {
	rq := &adblock.Request{
		URL:     url,
		Domain:  host,
		Referer: referer,
	}

	domainParts := strings.Split(host, ".")
	for _, matcherCategory := range matcherCategories {
		for _, matcher := range matcherCategory.Matchers {
			matched, matchType, err := matcher.Match(rq)
			if err != nil {
				log.Printf("Error matching rule %s", err)
			}

			if matched {
				return &matcherCategory.Category, matchType
			}
		}

		matched, matchType := matchDomain(domainParts, matcherCategory)
		if matched {
			return &matcherCategory.Category, matchType
		}
	}

	return nil, Included
}

func matchDomain(domainParts []string, matcherCatergory *MatcherCategory) (bool, int) {
	partsLen := len(domainParts)
	if partsLen < 2 {
		return false, Included
	}
	domainName := domainParts[partsLen-1]
	for i := len(domainParts) - 2; i >= 0; i-- {
		domainName = domainParts[i] + "." + domainName
		value, ok := matcherCatergory.BlockedDomains[domainName]
		if ok {
			if value {
				return true, Included
			} else {
				return true, Excluded
			}
		}
	}
	return false, Included
}

func (am *AdBlockMatcher) TestContentTypeIsFiltrable(contentType string) bool {
	return strings.Contains(contentType, "html") ||
		strings.Contains(contentType, "json") ||
		strings.Contains(contentType, "text")
}

func (am *AdBlockMatcher) IsContentSmallEnoughToFilter(contentSize int64) bool {
	return contentSize > 0 && contentSize < MAX_CONTENT_SIZE_SCAN
}

func (am *AdBlockMatcher) TestContainsForbiddenPhrases(str []byte) (*string, []string) {
	originalText := strings.ToLower(string(str))
	text := []rune(originalText)

	for _, phraseCategory := range am.PhraseCategories {
		if phraseCategory.processor == nil {
			log.Printf("Searching text trigger: nil")
			continue
		}

		res := phraseCategory.processor.MultiPatternSearch(text, true)
		if len(res) > 0 {
			words := make([]string, 0)
			for _, term := range res {
				startIndex := term.Pos
				endIndex := term.Pos + len(string(term.Word))

				//check if there's whole word match
				wholewordMatched := false
				if startIndex > 0 && isNonLetterAndDigitRune(text[startIndex-1]) {
					if endIndex < len(text)-2 && isNonLetterAndDigitRune(text[endIndex+1]) {
						wholewordMatched = true
					}
				}
				if wholewordMatched {
					words = append(words, string(term.Word))
				}
			}
			if len(words) > 0 {
				return &phraseCategory.Category, words
			} else {
				return nil, nil
			}
		}
	}

	return nil, nil
}

func isNonLetterAndDigitRune(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsDigit(r)
}

func (am *AdBlockMatcher) AddBlockedPhrase(phrase string, category string) {
	var phraseCategory *PhraseCategory = nil
	for _, element := range adBlockMatcher.PhraseCategories {
		if element.Category == category {
			phraseCategory = element
			break
		}
	}

	if phraseCategory == nil {
		phraseCategory = &PhraseCategory{
			Category: category,
		}

		am.PhraseCategories = append(am.PhraseCategories, phraseCategory)
	}

	phraseCategory.Phrases = append(phraseCategory.Phrases, phrase)
}

func (am *AdBlockMatcher) Build() {
	am.phrasesCount = 0
	for _, phraseCategory := range am.PhraseCategories {
		processor := new(goahocorasick.Machine)

		dict := [][]rune{}
		for _, phrase := range phraseCategory.Phrases {
			dict = append(dict, []rune(strings.ToLower(phrase)))
		}
		processor.Build(dict)
		phraseCategory.processor = processor

		am.phrasesCount += len(phraseCategory.Phrases)
	}

	if len(am.MatcherCategories) == 0 {
		return
	}
	matchers := am.MatcherCategories[len(am.MatcherCategories)-1].Matchers
	am.lastMatcher = matchers[len(matchers)-1]

	debug.FreeOSMemory()
}

func (am *AdBlockMatcher) RulesCount() int {
	return am.RulesCnt
}

func (am *AdBlockMatcher) PhrasesCount() int {
	return am.phrasesCount
}

func (am *AdBlockMatcher) SaveToFile(filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error opening file %s %s", filePath, err)
		return
	}
	defer file.Close()

	stream := gzip.NewWriter(file)
	defer stream.Close()

	encoder := gob.NewEncoder(stream)
	err = encoder.Encode(am)
	if err != nil {
		log.Printf("Encoder error %s", err)
	}
}

func LoadMatcherFromFile(filePath string) *AdBlockMatcher {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file %s %s", filePath, err)
		return nil
	}
	defer file.Close()

	stream, err := gzip.NewReader(file)
	if err != nil {
		log.Printf("Error opening file %s %s", filePath, err)
		return nil
	}
	defer stream.Close()

	decoder := gob.NewDecoder(stream)

	adBlockMatcher = &AdBlockMatcher{
		RulesCnt: 0,
	}
	err = decoder.Decode(&adBlockMatcher)
	if err != nil {
		log.Printf("Decoder error %s", err)
	}
	return adBlockMatcher
}

func (am *AdBlockMatcher) EnableBypass() {
	am.bypassEnabled = true
}

func (am *AdBlockMatcher) DisaleBypass() {
	am.bypassEnabled = false
}
