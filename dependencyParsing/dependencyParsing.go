package dependencyParsing

import (
	"../processFiles"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

var Assets = []string{"btc", "eth", "bitcoin", "ethereum", "asset"}

type Token struct {
	Text struct {
		Content     string `json:"content"`
		BeginOffset int    `json:"beginOffset"`
	}
	PartOfSpeech struct {
		Tag string `json:"tag"`
	}
	DependencyEdge struct {
		HeadTokenIndex int    `json:"headTokenIndex"`
		Label          string `json:"label"`
	}
	Lemma string `json:"lemma"`
}

type Sentence struct {
	Text struct {
		Content     string `json:"content"`
		BeginOffset int    `json:"beginOffset"`
	}
}
type Tokenization struct {
	Sentences []Sentence `json:"sentences"`
	Tokens    []Token    `json:"tokens"`
}

type subString struct {
	subString           string
	BeginIdx            int
	EndIdx              int //end of idx is a last letter of subString
	SentenceNumber      int
	SentenceBeginOffset int
}

type PreBlobPostContent struct {
	Pre      string
	Blob     string
	Post     string
	Sentence string
	Label    string
}

type MatchAssetsInSentence struct {
	SentenceNumber int
	MatchedTokens  []Token
}

func (mais *MatchAssetsInSentence) isEmpty() bool {
	if len(mais.MatchedTokens) == 0 && mais.SentenceNumber == 0 {
		return true
	}
	return false
}

//Convert Json to data structure
func (data *Tokenization) GetDataStruct(body []uint8) {
	byte := []byte(body)
	if err := json.Unmarshal(byte, &data); err != nil {
		log.Error(err)
	}
}

//Check if word is in assets
func isWordInAssets(word string) bool {
	//var lowWord = strings.ToLower(word)
	//for _, asset := range Assets {
	//	if lowWord == asset {
	//		return true
	//	}
	//}
	if word == "ASSET" {
		return true
	}
	return false
}

//Iterate the whole array of words to find assets
func (data *Tokenization) MatchAssets() []Token {
	var assetsArr []Token
	for idx, token := range data.Tokens {
		fmt.Println("counter: ", idx)
		fmt.Println("token: ", token)
		if isWordInAssets(token.Text.Content) {
			assetsArr = append(assetsArr, token)
		}
	}
	return assetsArr
}

func (data *Tokenization) findSentenceNumber(tokenArr []Token) []MatchAssetsInSentence {
	var matchedAssets []MatchAssetsInSentence
	for idxSentence, sentence := range data.Sentences {
		var endIdxSentence = sentence.Text.BeginOffset + len(sentence.Text.Content)
		var assetsInSentence MatchAssetsInSentence
		for idxToken := 0; idxToken < len(tokenArr); idxToken++ {
			var tokenBeginIdx = tokenArr[idxToken].Text.BeginOffset
			if tokenBeginIdx >= sentence.Text.BeginOffset && tokenBeginIdx < endIdxSentence {
				assetsInSentence.SentenceNumber = idxSentence + 1
				assetsInSentence.MatchedTokens = append(assetsInSentence.MatchedTokens, tokenArr[idxToken])
			}
		}
		if !assetsInSentence.isEmpty() {
			matchedAssets = append(matchedAssets, assetsInSentence)
		}
	}
	return matchedAssets
}

//Find all dependencies for an asset
func (data *Tokenization) getDependencyStringForAsset(asset Token) subString {
	var tokensArr [] string
	var subStr subString
	var idx = 0

	for _, token := range data.Tokens {
		isHeadToken := token.DependencyEdge.HeadTokenIndex == asset.DependencyEdge.HeadTokenIndex
		isNotPunctuation := token.DependencyEdge.Label != "P"
		//var rootIdx int
		//if token.DependencyEdge.Label == "ROOT" {
		//	rootIdx = token.DependencyEdge.HeadTokenIndex
		//}
		//isRootIdx := token.DependencyEdge.HeadTokenIndex == rootIdx
		if isHeadToken && isNotPunctuation {
			if idx == 0 {
				subStr.BeginIdx = token.Text.BeginOffset
			}
			subStr.EndIdx = token.Text.BeginOffset + len(token.Text.Content) - 1
			if isWordInAssets(token.Text.Content) {
				tokensArr = append(tokensArr, "[ASSET]")
			} else {
				tokensArr = append(tokensArr, strings.ToLower(token.Lemma))
			}

			idx ++
		}
	}

	lengthTokensArr := len(tokensArr)

	if lengthTokensArr == 0 {
		return subStr
	}

	subStr.subString = strings.Join(tokensArr, " ")
	return subStr
}

//Find all dependencies for items in array of assets and create a substring from it
func (data *Tokenization) getDependenciesForAssets(assetsInSentencesDict []MatchAssetsInSentence) []subString {
	var subStringsArr []subString
	for _, matchedAssetsInSentence := range assetsInSentencesDict {
		for _, asset := range matchedAssetsInSentence.MatchedTokens {
			var dependencyTokens = data.getDependencyStringForAsset(asset)
			if dependencyTokens != (subString{}) {
				dependencyTokens.SentenceNumber = matchedAssetsInSentence.SentenceNumber
				dependencyTokens.SentenceBeginOffset = data.Sentences[matchedAssetsInSentence.SentenceNumber-1].Text.BeginOffset
				subStringsArr = append(subStringsArr, dependencyTokens)
			}
		}

	}
	//fmt.Println("dependencies", subStringsArr)
	return subStringsArr
}


func (data *Tokenization) getSentences() string {
	var sentences string
	for _, sentence := range data.Sentences {
		sentences += sentence.Text.Content
		sentences += " "
	}
	return sentences
}

func (data *Tokenization) processSentences(assetsSubStrings []subString, documents []processFiles.Document) []PreBlobPostContent {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Error(err)
	}
	var preBlobPostArr []PreBlobPostContent
	var sentences = data.getSentences()
	for idx, assetSubString := range assetsSubStrings {
		var PreBlobPost PreBlobPostContent

		var sentence = data.Sentences[assetSubString.SentenceNumber-1].Text.Content
		var endIdxSentence = data.Sentences[assetSubString.SentenceNumber-1].Text.BeginOffset + len(sentence) - 1
		PreBlobPost.Sentence = sentence
		PreBlobPost.Blob = reg.ReplaceAllString(assetSubString.subString, " ")
		PreBlobPost.Label = documents[assetSubString.SentenceNumber-1].Sentiment
		fmt.Println("Sentence: ", PreBlobPost.Sentence, "LABEL: ", PreBlobPost.Label)

		if assetSubString.BeginIdx > assetSubString.SentenceBeginOffset {
			PreBlobPost.Pre =strings.ToLower(reg.ReplaceAllString(sentences[assetSubString.SentenceBeginOffset:assetSubString.BeginIdx], " "))
		}
		if idx+1 != len(assetsSubStrings) {
			var nextAssetSubstr = assetsSubStrings[idx+1]
			if nextAssetSubstr.SentenceNumber-assetSubString.SentenceNumber != 0 {
				//fmt.Println("end idx: ", assetSubString.EndIdx+1)
				//fmt.Println("sentence end: ", endIdxSentence)
				//fmt.Println("data tokens: ", len(data.Tokens))
				if assetSubString.EndIdx+1 < endIdxSentence {
					PreBlobPost.Post = strings.ToLower(reg.ReplaceAllString(sentences[assetSubString.EndIdx+1 : endIdxSentence], " "))
				}

			} else  {
				//fmt.Println("sentences len: ", len(sentences))
				//fmt.Println("len of begin index: ", nextAssetSubstr.BeginIdx)
				//fmt.Println("IND: ", assetSubString.EndIdx+1)
				if assetSubString.EndIdx+1 < nextAssetSubstr.BeginIdx {
					PreBlobPost.Post = strings.ToLower(reg.ReplaceAllString(sentences[assetSubString.EndIdx+1 : nextAssetSubstr.BeginIdx], " "))
				}

			}
		} else {
			PreBlobPost.Post = strings.ToLower(reg.ReplaceAllString(sentences[assetSubString.EndIdx+1:]," "))
		}
		fmt.Println("PRE: ", PreBlobPost.Pre)
		fmt.Println("BLOB: ", PreBlobPost.Blob)
		fmt.Println("POST: ", PreBlobPost.Post)
		preBlobPostArr = append(preBlobPostArr, PreBlobPost)
	}
	return preBlobPostArr
}

func (data *Tokenization) TreeDependency(documents []processFiles.Document) []PreBlobPostContent {
	fmt.Println("data tokens: ", data.Tokens)
	var assetsDict = data.MatchAssets()
	var dependencyStringsForAssets = data.getDependenciesForAssets(data.findSentenceNumber(assetsDict))

	return data.processSentences(dependencyStringsForAssets, documents)
}
