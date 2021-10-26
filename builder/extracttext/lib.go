package crawler

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

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

func getDocumentPlainText(p string) (string, error) {
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

func pdftohtml(p string) error {
	cmd := exec.Command("pdftohtml", p)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to pipe pdftohtml")
		return err
	}

	err = cmd.Start()
	if err != nil {
		stderrBytes, _ := ioutil.ReadAll(stderr)
		log.WithFields(log.Fields{
			"stderr": string(stderrBytes),
			"error":  err.Error(),
		}).Error("failed to run pdftohtml")
		return err
	}

	err = cmd.Wait()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to wait for pdftohtml")
		return err
	}
	return nil
}

func readImageText(filename string) (string, error) {
	cmd := exec.Command("tesseract", "-l", "spa", filename, "-")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to pipe tesseract")
		return "", err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to pipe tesseract")
		return "", err
	}

	err = cmd.Start()
	if err != nil {
		stderrBytes, _ := ioutil.ReadAll(stderr)
		log.WithFields(log.Fields{
			"stderr": string(stderrBytes),
			"error":  err.Error(),
		}).Error("failed to run tesseract")
		return "", err
	}

	stdoutBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to read tesseract's output")
		return "", err
	}
	err = cmd.Wait()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to wait for tesseract")
		return "", err
	}
	return string(stdoutBytes), nil
}

func getDocumentImagesText(dir, p string) (string, error) {
	return "", nil // FIXME: enable this
	log.WithFields(log.Fields{
		"directory": dir,
		"path":      p,
	}).Info("getting pdf images")
	err := pdftohtml(p)
	if err != nil {
		return "", err
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}
	text := ""
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".png") {
			log.WithFields(log.Fields{
				"image": filename,
			}).Info("reading image text")
			t, err := readImageText(path.Join(dir, filename))
			if err != nil {
				continue
			}
			text = fmt.Sprintf("%s\n%s", text, t)
		}
	}
	return text, nil
}

func GetDocumentText(r io.Reader) (string, error) {
	dir, p, err := writeToTempFile(r)
	defer os.RemoveAll(dir)
	if err != nil {
		return "", err
	}
	log.Info("getting pdf plain text")
	text, err := getDocumentPlainText(p)
	if err != nil {
		return "", err
	}
	log.Info("getting pdf images text")
	text2, err := getDocumentImagesText(dir, p)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n%s", text, text2), nil
}
