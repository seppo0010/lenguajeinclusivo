package main

import (
	"bufio"
	"encoding/json"
	"os"

	crawler "github.com/odia/juscaba/crawler"
	extracttext "github.com/odia/juscaba/extracttext"
	fetcher "github.com/odia/juscaba/fetcher"
	shared "github.com/odia/juscaba/shared"
	log "github.com/sirupsen/logrus"
)

func readBlacklist(path string) (map[string]bool, error) {
	res := map[string]bool{}
	file, err := os.Open(path)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"path":  path,
		}).Error("failed to read blacklist file")
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		res[scanner.Text()] = true
	}

	if err := scanner.Err(); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"path":  path,
		}).Error("failed to scan blacklist file")
		return nil, err
	}
	return res, nil
}

func main() {
	blacklistPath := os.Args[4]
	jsonPath := os.Args[3]
	fm := &shared.FileManager{Directory: os.Args[2]}
	exp, err := crawler.GetExpediente(os.Args[1])
	if err != nil {
		os.Exit(1)
	}
	blacklist, err := readBlacklist(blacklistPath)
	if err != nil {
		os.Exit(1)
	}
	log.WithFields(log.Fields{
		"expediente":  exp,
		"actuaciones": len(exp.Actuaciones),
	}).Printf("finished")

	for _, act := range exp.Actuaciones {
		for _, doc := range act.Documentos {
			skip, _ := blacklist[doc.URL]
			if skip {
				log.WithFields(log.Fields{
					"url": doc.URL,
				}).Info("skipping blacklisted URL")
				continue
			}
			fetcher.Download(fm, doc.URL)
			reader, err := fm.GetReader(doc.URL)
			if err != nil {
				continue
			}
			doc.Content, _ = extracttext.GetDocumentText(reader)
		}
	}
	fp, err := os.Create(jsonPath)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		})
		os.Exit(2)
	}
	defer fp.Close()
	json.NewEncoder(fp).Encode(exp)
}
