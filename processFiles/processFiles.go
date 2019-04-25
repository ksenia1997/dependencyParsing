package processFiles

import (
	"bytes"
	"encoding/csv"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)
type Document struct {
	Sentiment string
	Text      string
}
//Open data source, parse labels and sentences
func OpenCSV() (string, []Document) {
	var labels []string
	var sentences string
	var documents []Document
	rFile, err := os.Open("final.csv")
	if err != nil {
		log.Error(err)
		return "", nil
	}
	defer rFile.Close()
	reader := csv.NewReader(rFile)

	counter := 0
	for {

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		sentencePart := strings.Join(strings.Fields(record[0]), " ") + ". "
		sentences += sentencePart
		counter++
		labels = append(labels, record[1])
		documents = append(documents, Document{strings.TrimSpace(record[1]), sentencePart})
	}

	return sentences, documents
}

//get json file with dependency trees
func SendPostRequest(jsonStr []byte) []uint8 {
	url := "https://language.googleapis.com/v1/documents:analyzeSyntax?fields=language%2Csentences%2Ctokens&key=AIzaSyAMPV13zzIjcv3jpYYA4xKAHL2MJIU9bPs"
	//fmt.Println("URL:>", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return nil
	}
	return body

}
