package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
)

var FileNotSaved = errors.New("requested file has not been saved")
var InvalidSavedFile = errors.New("requested file has invalid metadata")

type SavedFile struct {
	SourceURL           string    `json:"sourceURL"`
	DestinationFilename string    `json:"destinationFilename"`
	FetchDate           time.Time `json:"fetchDate"`
}

func NewSavedFile(sourceURL, destinationFilename string) *SavedFile {
	return &SavedFile{
		SourceURL:           sourceURL,
		DestinationFilename: destinationFilename,
		FetchDate:           time.Now(),
	}
}

type FileManager struct {
	Directory string
}

func (s *FileManager) metadataPath(url string) string {
	return path.Join(s.Directory, GetSha1(url)+".json")
}

func (s *FileManager) destinationPath(sf *SavedFile) string {
	return path.Join(s.Directory, sf.DestinationFilename)
}

func (s *FileManager) DestinationURLforSourceURL(url string) (string, error) {
	sf, err := s.SavedFileForURL(url)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://srfp-documentos.odia.legal/%s", sf.DestinationFilename), nil
}

func (s *FileManager) IsSaved(url string) bool {
	_, err := os.Stat(s.metadataPath(url))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("file exists")
	}
	return err == nil
}

func (s *FileManager) SavedFileForURL(url string) (*SavedFile, error) {
	reader, err := os.Open(s.metadataPath(url))
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err.Error(),
		}).Info("error reading metadata file")
		return nil, FileNotSaved
	}
	defer reader.Close()

	var savedFile SavedFile
	err = json.NewDecoder(reader).Decode(&savedFile)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err.Error(),
		}).Warn("error decoding metadata file")
		return nil, InvalidSavedFile
	}
	return &savedFile, nil
}

func (s *FileManager) SaveSavedFile(sf *SavedFile, content []byte) error {
	contentWriter, err := os.Create(s.destinationPath(sf))
	if err != nil {
		log.WithFields(log.Fields{
			"savedFile": sf,
			"error":     err.Error(),
		}).Error("failed to create content file")
		return err
	}
	defer contentWriter.Close()
	_, err = contentWriter.Write(content)
	if err != nil {
		log.WithFields(log.Fields{
			"savedFile": sf,
			"error":     err.Error(),
		}).Error("failed to write content file")
		return err
	}

	metadataWriter, err := os.Create(s.metadataPath(sf.SourceURL))
	if err != nil {
		log.WithFields(log.Fields{
			"savedFile": sf,
			"error":     err.Error(),
		}).Error("failed to create metadata file")
		return err
	}
	defer metadataWriter.Close()
	err = json.NewEncoder(metadataWriter).Encode(sf)
	if err != nil {
		log.WithFields(log.Fields{
			"savedFile": sf,
			"error":     err.Error(),
		}).Error("failed to write metadata file")
		return err
	}
	return nil
}

func (s *FileManager) GetReader(url string) (io.Reader, error) {
	sf, err := s.SavedFileForURL(url)
	if err != nil {
		return nil, err
	}
	fp, err := os.Open(s.destinationPath(sf))
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err.Error(),
		}).Error("failed to read content file")
	}
	return fp, nil
}
