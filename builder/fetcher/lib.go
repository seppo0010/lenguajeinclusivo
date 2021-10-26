package fetcher

import (
	"fmt"
	"net/http"
	"os"

	"github.com/seppo0010/juscaba/shared"
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

	f, err := os.Create(s.Path(url))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"url":   url,
		}).Error("Failed to write file")
		return err
	}
	defer f.Close()
	if res.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{
			"url":         url,
			"status code": res.StatusCode,
		}).Warn("Did not save, leaving file empty")
		return fmt.Errorf("did not save status %d", res.StatusCode)
	}

	f.ReadFrom(res.Body)
	log.WithFields(log.Fields{
		"url": url,
	}).Info("saved")

	return nil
}
