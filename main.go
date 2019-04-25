package main

import (
	"./bayes"
	"./dependencyParsing"
	"fmt"
	"math/rand"
	"strconv"
	log "github.com/sirupsen/logrus"
)

// parameters
var (
	testPercentage = 0.1
	threshold      = 0.0
)

var categories = []string{"1", "0", "-1"}

func main() {
	processedData := bayes.PrepareTrainDataset()
	fmt.Println(processedData)
	train, test := splitDataForTestingAndTrain(processedData)
	blobBayes := bayes.CreateClassifier(categories, threshold)
	preBayes := bayes.CreateClassifier(categories, threshold)
	postBayes := bayes.CreateClassifier(categories, threshold)

	//Train data
	for _, data := range train {
		blobBayes.TrainBayes(data.Label, data.Blob)
		preBayes.TrainBayes(data.Label, data.Pre)
		postBayes.TrainBayes(data.Label, data.Post)
	}

	blobBayes.DeleteStopWords(categories)
	preBayes.DeleteStopWords(categories)
	postBayes.DeleteStopWords(categories)

	blobBayes.Save("blobbayes.json")
	preBayes.Save("prebayes.json")
	postBayes.Save("postbayes.json")
	//test
	countPositive, countNegative, truePositive, trueNegative, falsePositive, falseNegative := 0, 0, 0, 0, 0, 0
	for _, data := range test {

		sentimentBlob := blobBayes.Classify(data.Blob)

		sentimentPre := preBayes.Classify(data.Pre)

		sentimentPost := postBayes.Classify(data.Post)

		//fmt.Println("----------------------------------------------")
		//fmt.Println("Sentence: ", data.Sentence)
		//fmt.Println("Real sentiment: ", data.Label)
		//fmt.Println("BLOB: ", data.Blob)
		//fmt.Println("PRE: ", data.Pre)
		//fmt.Println("POST: ", data.Post)

		iSentimentBlob, err := strconv.ParseInt(sentimentBlob, 10, 64)
		if err != nil {
			log.Error(err)
		}
		iSentimentPre, err := strconv.ParseInt(sentimentPre, 10, 64)
		if err != nil {
			log.Error(err)
		}
		iSentimentPost, err := strconv.ParseInt(sentimentPost, 10, 64)
		if err != nil {
			log.Error(err)
		}
		//sentiment
		sentiment := float64(iSentimentPre)*0.25 + float64(iSentimentBlob)*0.5 + float64(iSentimentPost)*0.25
		fmt.Println("sentiment: ", sentiment)
		if data.Label == "1" {
			countPositive++
			if sentiment > 0 {
				truePositive ++
			} else {

				falsePositive++
			}

		} else if data.Label == "-1" {
			countNegative++
			if sentiment < 0 {
				trueNegative++
			} else {
				falseNegative++
			}
		}

	}
	fmt.Printf("\ntruePositive on TRAIN dataset is %2.1f%% \nfalsePositive is %2.1f%% \ntrueNegative is %2.1f%% \nfalseNegative is %2.1f%% ", float64(truePositive)*100/float64(countPositive), float64(falsePositive)*100/float64(countPositive), float64(trueNegative)*100/float64(countNegative), float64(falseNegative)*100/float64(countNegative))

}

func splitDataForTestingAndTrain(data []dependencyParsing.PreBlobPostContent) ([]dependencyParsing.PreBlobPostContent, []dependencyParsing.PreBlobPostContent) {
	var train []dependencyParsing.PreBlobPostContent
	var test []dependencyParsing.PreBlobPostContent
	for _, line := range data {
		if rand.Float64() > testPercentage {
			train = append(train, line)
		} else {
			test = append(test, line)
		}
	}
	return train, test
}
