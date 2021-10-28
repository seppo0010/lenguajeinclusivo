package main

import (
	"encoding/json"
	"os"

	crawler "github.com/seppo0010/juscaba/crawler"
	extracttext "github.com/seppo0010/juscaba/extracttext"
	fetcher "github.com/seppo0010/juscaba/fetcher"
	shared "github.com/seppo0010/juscaba/shared"
	log "github.com/sirupsen/logrus"
)

func main() {
	jsonPath := os.Args[3]
	fm := &shared.FileManager{Directory: os.Args[2]}
	exp, err := crawler.GetExpediente(os.Args[1])
	if err != nil {
		os.Exit(1)
	}
	log.WithFields(log.Fields{
		"expediente":  exp,
		"actuaciones": len(exp.Actuaciones),
	}).Printf("finished")

	for _, act := range exp.Actuaciones {
		for _, doc := range act.Documentos {
			fetcher.Download(fm, doc.URL)
		}
	}
	for _, act := range exp.Actuaciones {
		for _, doc := range act.Documentos {
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
