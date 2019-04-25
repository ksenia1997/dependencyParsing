package ngrams

import (
	"regexp"
	"strings"
)

const Ngram = 2

func GetNgrams(sentence string) []string {
	space := regexp.MustCompile("\\s+")
	sentence = space.ReplaceAllString(sentence, " ")

	var nGramsForSentence []string
	words := strings.Split(sentence, " ")

	if words[0] == "" {
		words = words[1:]
	}
	for index := 0; index < len(words)-Ngram+1; index++ {
		nGramsForSentence = append(nGramsForSentence, strings.Join(words[index:index+Ngram], " "))
	}

	return nGramsForSentence
}
