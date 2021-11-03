package fetcher

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/odia/juscaba/shared"
	log "github.com/sirupsen/logrus"
)

func Download(s *shared.FileManager, url string) error {
	if s.IsSaved(url) {
		log.WithFields(log.Fields{
			"url": url,
		}).Printf("skipping url")
		return nil
	}
	res, err := http.Get(url)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"url":   url,
		}).Warn("Failed to get url")
		return err
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"url":   url,
		}).Error("Failed to read url data")
		return err
	}

	h := sha1.New()
	_, err = h.Write(content)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"url":   url,
		}).Error("Failed to create hash")
		return err
	}

	savedFile := shared.NewSavedFile(url, fmt.Sprintf("%x.pdf", h.Sum(nil)))
	return s.SaveSavedFile(savedFile, content)
}
