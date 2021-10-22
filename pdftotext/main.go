package main

import (
	"context"
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
	dir, err := ioutil.TempDir("", "pdftotext")
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

func updateActuacionWithText(url, text string) error {
	es, err := elastic.NewClient(
		elastic.SetURL("http://es:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to connect to elastic")
		return err
	}

	q, err := es.UpdateByQuery(actuacionType).
		Query(elastic.NewTermQuery("URL", url)).
		Script(elastic.NewScript("ctx._source.text = params['t']").
			Params(map[string]interface{}{"t": text})).
		Do(context.Background())
	log.WithFields(log.Fields{"q": q}).Infof("%#v", q)
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
		reader, err := shared.GetURLContent(url, minioClient, bucketName)
		if err != nil {
			continue
		}
		text, err := getDocumentText(reader)
		if err != nil {
			continue
		}
		err = updateActuacionWithText(url, text)
		if err != nil {
			continue
		}
		log.WithFields(log.Fields{
			"url": url,
		}).Info("updated")
	}
}
