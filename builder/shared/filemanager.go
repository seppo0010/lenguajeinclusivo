package shared

import (
	"errors"
	"io"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

type FileManager struct {
	Directory string
}

func (s *FileManager) Path(url string) string {
	return path.Join(s.Directory, GetSha1(url)) + ".pdf"
}

func (s *FileManager) IsSaved(url string) bool {
	_, err := os.Stat(s.Path(url))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("file exists")
		}
		return false
	}
	return true
}

func (s *FileManager) GetReader(url string) (io.Reader, error) {
	reader, err := os.Open(s.Path(url))
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
		}).Error("failed to open reader")
	}
	return reader, nil
}
