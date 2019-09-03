package main

import (
	"compress/gzip"
	"encoding/gob"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/patriciy/adblock/adblock"
)

const (
	Included = adblock.Included
	Excluded = adblock.Excluded
)

const MAX_RULES_PER_MATCHER = 1000
const MAX_CONTENT_SIZE_SCAN = 200 * 1024 //200kb max to scan
var adblockMatcher *AdBlockMatcher

var defaultBlockPageContent = "%url% is blocked. Category %category%. Reason %reason%"

type MatcherCategory struct {
	CategoryId     int32
	ListType       int32
	Matchers       []*adblock.RuleMatcher
	BlockedDomains map[string]bool
}

type PhraseCategory struct {
	Category string
	Phrases  []string
	regexp   *regexp.Regexp
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
	BlockPageContent        string
}

func CreateMatcher() *AdBlockMatcher {
	adblockMatcher = &AdBlockMatcher{
		RulesCnt:         0,
		BlockPageContent: defaultBlockPageContent,
	}

	return adblockMatcher
}

func (am *AdBlockMatcher) addMatcher(categoryId int32, listType int32, bypass bool) {
	matcher := adblock.NewMatcher()
	var categoryMatcher *MatcherCategory
	for _, element := range adblockMatcher.MatcherCategories {
		if element.CategoryId == categoryId {
			categoryMatcher = element
			break
		}
	}

	if categoryMatcher == nil {
		categoryMatcher = &MatcherCategory{
			CategoryId:     categoryId,
			ListType:       listType,
			BlockedDomains: make(map[string]bool),
		}

		if bypass {
			am.BypassMatcherCategories = append(am.BypassMatcherCategories, categoryMatcher)
		} else {
			am.MatcherCategories = append(am.MatcherCategories, categoryMatcher)
		}
	}

	categoryMatcher.Matchers = append(categoryMatcher.Matchers, matcher)
	adblockMatcher.lastMatcher = matcher
	adblockMatcher.lastCategory = categoryMatcher
}

func (am *AdBlockMatcher) GetBlockPage(url string, category string, reason string) string {
	tagsReplacer := strings.NewReplacer("%url%", url,
		"%category%", category,
		"%reason%", reason)
	return tagsReplacer.Replace(am.BlockPageContent)
}

func (am *AdBlockMatcher) TestUrlBlockedWithMatcherCategories(url string, host string, referer string) ([]*MatcherCategory, []int) {
	res1, res2 := am.matchRulesCategories(am.MatcherCategories, url, host, referer)
	if len(res1) > 0 {
		return res1, res2
	}

	if am.bypassEnabled {
		return make([]*MatcherCategory, 0), make([]int, 0)
	}

	return am.matchRulesCategories(am.BypassMatcherCategories, url, host, referer)
}

func TransformMatcherCategoryArrayToIntArray(categories []*MatcherCategory) []int32 {
	ret := make([]int32, len(categories))

	for i, category := range categories {
		ret[i] = category.CategoryId
	}

	return ret
}

func (am *AdBlockMatcher) TestUrlBlocked(url string, host string, referer string) []int32 {
	categories, _ := am.TestUrlBlockedWithMatcherCategories(url, host, referer)
	return TransformMatcherCategoryArrayToIntArray(categories)
}

func (am *AdBlockMatcher) matchRulesCategories(matcherCategories []*MatcherCategory, url string, host string, referer string) ([]*MatcherCategory, []int) {
	rq := &adblock.Request{
		URL:     url,
		Domain:  host,
		Referer: referer,
	}

	var matchedCategories []*MatcherCategory
	var catergoriesMatchType []int //Included, Excluded

	domainParts := strings.Split(host, ".")

	for _, matcherCategory := range matcherCategories {
		categoryMatched := false
		for _, matcher := range matcherCategory.Matchers {
			matched, categoryType, err := matcher.Match(rq)
			if err != nil {
				log.Printf("Error matching rule %s", err)
			}

			if matched {
				categoryMatched = true
				matchedCategories = append(matchedCategories, matcherCategory)
				catergoriesMatchType = append(catergoriesMatchType, categoryType)
				break
			}
		}

		if !categoryMatched && matchDomain(domainParts, matcherCategory) {
			matchedCategories = append(matchedCategories, matcherCategory)
			catergoriesMatchType = append(catergoriesMatchType, Included)
		}
	}

	return matchedCategories, catergoriesMatchType
}

func matchDomain(domainParts []string, matcherCatergory *MatcherCategory) bool {
	partsLen := len(domainParts)
	if partsLen < 2 {
		log.Printf("Domain too short")
		return false
	}
	domainName := domainParts[partsLen-1]
	for i := len(domainParts) - 2; i >= 0; i-- {
		domainName = domainParts[i] + "." + domainName
		_, ok := matcherCatergory.BlockedDomains[domainName]
		if ok {
			return true
		}
	}
	return false
}

func (am *AdBlockMatcher) TestContentTypeIsFiltrable(contentType string) bool {
	return strings.Contains(contentType, "html") ||
		strings.Contains(contentType, "json") ||
		strings.Contains(contentType, "text")
}

func (am *AdBlockMatcher) IsContentSmallEnoughToFilter(contentSize int64) bool {
	return contentSize < MAX_CONTENT_SIZE_SCAN
}

func (am *AdBlockMatcher) TestContainsForbiddenPhrases(str []byte) *string {
	for _, phraseCategory := range am.PhraseCategories {
		if phraseCategory.regexp != nil {
			if phraseCategory.regexp.Find(str) != nil {
				return &phraseCategory.Category
			}
		}
	}

	return nil
}

func (am *AdBlockMatcher) AddBlockedPhrase(phrase string, category string) {
	var phraseCategory *PhraseCategory = nil
	for _, element := range adblockMatcher.PhraseCategories {
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

	phraseCategory.Phrases = append(phraseCategory.Phrases, regexp.QuoteMeta(phrase))
}

func (am *AdBlockMatcher) Build() {
	am.phrasesCount = 0
	for _, phraseCategory := range am.PhraseCategories {
		regexString := strings.Join(phraseCategory.Phrases, "|")

		var e error
		phraseCategory.regexp, e = regexp.Compile("(?i)" + regexString)
		if e != nil {
			log.Printf("Error compiling matcher %s", e)
		}
		am.phrasesCount += len(phraseCategory.Phrases)
	}

	matchers := am.MatcherCategories[len(am.MatcherCategories)-1].Matchers
	am.lastMatcher = matchers[len(matchers)-1]
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

	adblockMatcher = &AdBlockMatcher{
		RulesCnt: 0,
	}
	err = decoder.Decode(&adblockMatcher)
	if err != nil {
		log.Printf("Decoder error %s", err)
	}
	return adblockMatcher
}

func (am *AdBlockMatcher) EnableBypass() {
	am.bypassEnabled = true
}

func (am *AdBlockMatcher) DisaleBypass() {
	am.bypassEnabled = false
}
