package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"os"

	crawler "github.com/odia/juscaba/crawler"
	extracttext "github.com/odia/juscaba/extracttext"
	fetcher "github.com/odia/juscaba/fetcher"
	shared "github.com/odia/juscaba/shared"
	log "github.com/sirupsen/logrus"
)

type arguments struct {
	blacklist   map[string]bool
	jsonPath    string
	fm          *shared.FileManager
	exp         *shared.Expediente
	parseImages bool
}

func parseArguments() (*arguments, error) {
	var pdfsPath, expId, blacklistPath string
	var err error
	args := arguments{}
	flag.StringVar(&blacklistPath, "blacklist", "", "path to file with urls not to download")
	flag.StringVar(&args.jsonPath, "json", "", "json destination path")
	flag.StringVar(&pdfsPath, "pdfs", "", "pdfs destination path")
	flag.StringVar(&expId, "expediente", "", "expediente identifier (e.g.: \"182908/2020-0\")")
	flag.BoolVar(&args.parseImages, "images", true, "apply ocr")
	flag.Parse()

	log.WithFields(log.Fields{
		"blacklist":   blacklistPath,
		"json":        args.jsonPath,
		"pdfs":        pdfsPath,
		"expediente":  expId,
		"parseImages": args.parseImages,
	}).Print("arguments")

	args.fm = &shared.FileManager{Directory: pdfsPath}
	args.exp, err = crawler.GetExpediente(expId)
	if err != nil {
		return nil, err
	}

	args.blacklist, err = readBlacklist(blacklistPath)
	if err != nil {
		return nil, err
	}
	return &args, nil
}

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
	args, err := parseArguments()
	if err != nil {
		os.Exit(1)
	}
	log.WithFields(log.Fields{
		"expediente":  args.exp,
		"actuaciones": len(args.exp.Actuaciones),
	}).Printf("finished")

	for _, act := range args.exp.Actuaciones {
		for _, doc := range act.Documentos {
			skip, _ := args.blacklist[doc.URL]
			if skip {
				log.WithFields(log.Fields{
					"url": doc.URL,
				}).Info("skipping blacklisted URL")
				continue
			}
			fetcher.Download(args.fm, doc.URL)
			reader, err := args.fm.GetReader(doc.URL)
			if err != nil {
				continue
			}
			doc.Content, _ = extracttext.GetDocumentText(reader, args.parseImages)
		}
	}
	fp, err := os.Create(args.jsonPath)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		})
		os.Exit(2)
	}
	defer fp.Close()
	json.NewEncoder(fp).Encode(args.exp)
}
