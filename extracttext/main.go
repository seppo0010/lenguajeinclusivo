package main

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/olivere/elastic/v7"
	"github.com/seppo0010/juscaba/shared"
	log "github.com/sirupsen/logrus"
)

const actuacionType = "actuacion"

func writeToTempFile(r io.Reader) (string, string, error) {
	dir, err := ioutil.TempDir("", "extracttext")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to create temp dir")
		return "", "", err
	}
	p := path.Join(dir, "content")
	f, err := os.Create(p)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to create temp file")
		return "", "", err
	}
	defer f.Close()
	_, err = f.ReadFrom(r)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to write to temp file")
		return "", "", err
	}
	return dir, p, nil
}
func getDocumentText(r io.Reader) (string, error) {
	_, p, err := writeToTempFile(r)
	if err != nil {
		return "", err
	}
	cmd := exec.Command("pdftotext", p, "-")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to pipe pdftotext")
		return "", err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to pipe pdftotext")
		return "", err
	}

	err = cmd.Start()
	if err != nil {
		stderrBytes, _ := ioutil.ReadAll(stderr)
		log.WithFields(log.Fields{
			"stderr": string(stderrBytes),
			"error":  err.Error(),
		}).Error("failed to run pdftotext")
		return "", err
	}

	stdoutBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to read pdftotext's output")
		return "", err
	}
	err = cmd.Wait()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to wait for pdftotext")
		return "", err
	}
	return string(stdoutBytes), nil
}
func documentHasText(es *elastic.Client, url string) (bool, error) {
	res, err := es.Search().
		Query(elastic.NewTermQuery("URL", url)).
		Do(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"url":   url,
		}).Error("failed to check for text")
		return false, err
	}
	hits := res.Hits.Hits
	if len(hits) == 0 {
		log.WithFields(log.Fields{
			"url": url,
		}).Warn("extracting text from unknown url")
		return true, nil
	}
	if len(hits) > 1 {
		log.WithFields(log.Fields{
			"url": url,
		}).Warn("more than one actuacion with the same URL found")
		return true, nil
	}
	var t struct {
		Text string `json:"text"`
	}
	json.Unmarshal(hits[0].Source, &t)
	return t.Text != "", nil
}

func updateActuacionWithText(es *elastic.Client, url, text string) error {
	_, err := es.UpdateByQuery(actuacionType).
		Query(elastic.NewTermQuery("URL", url)).
		Script(elastic.NewScript("ctx._source.text = params['t']").
			Params(map[string]interface{}{"t": text})).
		Do(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to update query")
		return err
	}
	return nil
}
func main() {
	bucketName := "pdfs"
	minioClient, err := shared.GetMinioClient(bucketName)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Connect to minio")
	}

	es, err := elastic.NewClient(
		elastic.SetURL("http://es:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to connect to elastic")
	}

	urls, err := shared.WaitForTasks("index", "indexer")
	if err != nil {
		log.Fatalln(err)
	}
	log.Info("waiting for urls")
	for urlBytes := range urls {
		url := string(urlBytes)
		log.WithFields(log.Fields{
			"url": url,
		}).Info("Received")
		hasText, _ := documentHasText(es, url)
		if hasText {
			log.WithFields(log.Fields{
				"url": url,
			}).Info("Skipping")
			continue
		}
		reader, err := shared.GetURLContent(url, minioClient, bucketName)
		if err != nil {
			continue
		}
		text, err := getDocumentText(reader)
		if err != nil {
			continue
		}
		err = updateActuacionWithText(es, url, text)
		if err != nil {
			continue
		}
		log.WithFields(log.Fields{
			"url": url,
		}).Info("updated")
	}
}
