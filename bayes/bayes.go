package bayes

import (
	"../dependencyParsing"
	"../nGram"
	"../processFiles"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"sort"
)

const thresholdStopWords = 0.9
const minThreshold = 0.001

type sorted struct {
	category    string
	probability float64
}

type SimpleBayes struct {
	Ngrams                   map[string]map[string]int `json:"ngrams"`
	TotalNgrams              int                       `json:"totalngrams"`
	CategoriesInDocuments    map[string]int            `json:"categories"`
	TotalDocuments           int                       `json:"totaldoc"`
	NumberNgramForCategories map[string]int            `json:"categoriesNgram"`
	Threshold                float64                   `json:"threshold"`
}

// Saves weights of words as json into filename
func (c *SimpleBayes) Save(filename string) error {
	e, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, e, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Loads weights from json in filename
func (c *SimpleBayes) Load(filename string) error {

	x, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(x, c)
	if err != nil {
		return err
	}

	return nil
}
func CreateClassifier(categories []string, threshold float64) (SimpleBayes) {
	c := SimpleBayes{
		Ngrams:                   make(map[string]map[string]int),
		TotalNgrams:              0,
		CategoriesInDocuments:    make(map[string]int),
		TotalDocuments:           0,
		NumberNgramForCategories: make(map[string]int),
		Threshold:                0,
	}

	for _, category := range categories {
		c.Ngrams[category] = make(map[string]int)
		c.CategoriesInDocuments[category] = 0
		c.NumberNgramForCategories[category] = 0
	}
	return c
}

func (c *SimpleBayes) countNgrams(document string) map[string]int {
	nGrams := ngrams.GetNgrams(document)
	var nGramsCount = make(map[string]int)
	for _, ngram := range nGrams {
		nGramsCount[ngram]++
	}
	return nGramsCount
}

func PrepareTrainDataset() []dependencyParsing.PreBlobPostContent {
	sentences, documents := processFiles.OpenCSV()
	var structForPostRequest = []byte(`{"document": {"content": "` + sentences + `", "type": "PLAIN_TEXT"}, "encodingType": "UTF8"}`)
	var body = processFiles.SendPostRequest(structForPostRequest)
	var data dependencyParsing.Tokenization
	data.GetDataStruct(body)

	if len(data.Sentences) != len(documents) {
		fmt.Println(len(data.Sentences))
		fmt.Println(len(documents))
		logrus.Error("Parse of sentences is unsuccessful")
		return nil
	}
	processedData := data.TreeDependency(documents)
	//for _, data := range processedData {
	//
	//	cleanPre := ngrams.GetNgrams(data.Pre)
	//	cleanBlob := ngrams.GetNgrams(data.Blob)
	//	cleanPost := ngrams.GetNgrams(data.Post)
	//
	//}

	return processedData

}
func (c *SimpleBayes) TrainBayes(label string, document string) {
	for ngram, count := range c.countNgrams(document) {
		//fmt.Println(c.Ngrams[label][ngram])
		c.Ngrams[label][ngram] += count
		c.NumberNgramForCategories[label] += count
		c.TotalNgrams += count
	}
	c.CategoriesInDocuments[label]++
	c.TotalDocuments++
}

func (c *SimpleBayes) probabilityForNgram(label string, ngram string) float64 {
	return math.Max(float64(c.Ngrams[label][ngram]), 1) / float64(c.NumberNgramForCategories[label])
}

func (c *SimpleBayes) deleteNgramFrom(label string, count int, ngram string) {
	c.TotalNgrams -= count
	c.NumberNgramForCategories[label] -= count
	delete(c.Ngrams[label], ngram)
}

func (c *SimpleBayes) DeleteStopWords(labels []string) {

	var pNgramsForCategoriesDict = make([]map[string]float64, len(labels))
	for idx, label := range labels {
		pNgramsForCategoriesDict[idx] = make(map[string]float64)
		for ngram, count := range c.Ngrams[label] {
			pForNgram := c.probabilityForNgram(label, ngram)
			if pForNgram < minThreshold {
				c.deleteNgramFrom(label, count, ngram)
			} else {
				pNgramsForCategoriesDict[idx][ngram] = pForNgram
			}

		}
	}
	if len(labels) != 3 {
		return
	}

	for ngram, prb := range pNgramsForCategoriesDict[0] {
		prb2, ok := pNgramsForCategoriesDict[2][ngram]
		if ok && (prb/prb2 > thresholdStopWords) {
			for _, deleteFromLabel := range labels {
				count := c.Ngrams[deleteFromLabel][ngram]
				c.deleteNgramFrom(deleteFromLabel, count, ngram)
			}
		}
	}

}

func (c *SimpleBayes) pNgramCategory(category string, ngram string) float64 {
	return float64(c.Ngrams[category][ngram]) / float64(c.NumberNgramForCategories[category])
}

func (c *SimpleBayes) pDocumentCategory(cateogry string, document string) float64 {
	var p = 1.0
	for ngram := range c.countNgrams(document) {
		p = p * c.pNgramCategory(cateogry, ngram)
	}
	return p
}

func (c *SimpleBayes) pCategory(category string) float64 {
	return float64(c.CategoriesInDocuments[category]) / float64(c.TotalDocuments)
}

func (c *SimpleBayes) pCategoryDocument(category string, document string) float64 {
	return c.pDocumentCategory(category, document) * c.pCategory(category)
}

func (c *SimpleBayes) Probabilities(document string) map[string]float64 {
	p := make(map[string]float64)
	for category := range c.Ngrams {
		p[category] = c.pCategoryDocument(category, document)
	}
	return p
}

func (c *SimpleBayes) Classify(document string) string {

	prob := c.Probabilities(document)
	var sp []sorted
	for c, p := range prob {
		if c != "0" {
			sp = append(sp, sorted{c, p})
		}
	}
	sort.Slice(sp, func(i, j int) bool {
		return sp[i].probability > sp[j].probability
	})
	var category string

	if sp[0].probability/sp[1].probability > c.Threshold {
		category = sp[0].category
	} else {
		category = "0"
	}
	return category

}
