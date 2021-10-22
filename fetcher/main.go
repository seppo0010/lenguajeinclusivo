package main

import (
	"context"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/seppo0010/juscaba/shared"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func urlIsSaved(url string, minioClient *minio.Client, bucketName string) bool {
	_, err := minioClient.StatObject(context.Background(), bucketName, shared.GetSha1(url), minio.StatObjectOptions{})
	return err == nil
}

func saveURLToBucket(url string, minioClient *minio.Client, bucketName string) error {
	res, err := http.Get(url)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"url":   url,
		}).Warn("Failed to get url")
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		uploadInfo, err := minioClient.PutObject(context.Background(), bucketName, shared.GetSha1(url), res.Body, -1, minio.PutObjectOptions{ContentType: res.Header.Get("content-type")})
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
				"url":   url,
			}).Error("Failed to create object")
			return err
		}
		log.WithFields(log.Fields{
			"url":        url,
			"uploadInfo": uploadInfo,
		}).Info("Uploaded")
	} else {
		log.Printf("did not save %s because status code was %d", url, res.StatusCode)
		log.WithFields(log.Fields{
			"url":         url,
			"status code": res.StatusCode,
		}).Warn("Did not save")
	}

	return nil
}

func enqueueIndex(url string) error {
	c, err := shared.InitTaskQueue("index")
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"url": url,
	}).Info("Enqueuing")
	err = c.Publish(
		"tasks",
		"index",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(url),
		})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to publish to queue")
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
	urls, err := shared.WaitForTasks("fetch", "fetcher")
	if err != nil {
		log.Fatalln(err)
	}
	log.Info("waiting for urls")
	for urlBytes := range urls {
		url := string(urlBytes)
		log.WithFields(log.Fields{
			"url": url,
		}).Info("Received")
		exists := urlIsSaved(url, minioClient, bucketName)
		if !exists {
			err = saveURLToBucket(url, minioClient, bucketName)
			if err != nil {
				continue
			}
		}
		err = enqueueIndex(url)
		if err != nil {
			continue
		}
	}
}
