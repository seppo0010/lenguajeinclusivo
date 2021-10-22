package main

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func getSha1(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func urlIsSaved(url string, minioClient *minio.Client, bucketName string) bool {
	_, err := minioClient.StatObject(context.Background(), bucketName, getSha1(url), minio.StatObjectOptions{})
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
		uploadInfo, err := minioClient.PutObject(context.Background(), bucketName, getSha1(url), res.Body, -1, minio.PutObjectOptions{ContentType: res.Header.Get("content-type")})
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

func initTaskQueue(name string) (*amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://queues/")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to connect to amqp")
		return nil, err
	}

	c, err := conn.Channel()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to create channel")
		return nil, err
	}
	err = c.ExchangeDeclare("tasks", "direct", true, false, false, false, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to declare exchange")
		return nil, err
	}
	_, err = c.QueueDeclare(name, true, false, false, false, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to declare queue")
		return nil, err
	}
	err = c.QueueBind(name, name, "tasks", false, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to bind queue")
		return nil, err
	}
	return c, nil
}

func enqueueIndex(url string) error {
	c, err := initTaskQueue("index")
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

func waitForURLs() (<-chan (string), error) {
	c, err := initTaskQueue("fetch")
	if err != nil {
		return nil, err
	}

	err = c.Qos(1, 0, false)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to set qos")
		return nil, err
	}
	tasks, err := c.Consume("fetch", "fetcher", false, false, false, false, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to create consumer")
		return nil, err
	}
	ch := make(chan string)
	go func(tasks <-chan amqp.Delivery) {
		for task := range tasks {
			ch <- string(task.Body)
			task.Ack(false)
		}
	}(tasks)
	return ch, nil
}

func readSecret(name string) (string, error) {
	body, err := ioutil.ReadFile(fmt.Sprintf("/run/secrets/%v", name))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"name":  name,
		}).Fatal("failed to read secret")
	}
	return strings.TrimSpace(string(body)), nil
}

func getMinioClient(bucketName string) (*minio.Client, error) {
	useSSL := false
	accessKeyID, err := readSecret("minio-user")
	if err != nil {
		return nil, err
	}
	secretAccessKey, err := readSecret("minio-password")
	if err != nil {
		return nil, err
	}
	minioClient, err := minio.New("minio:9000", &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("failed to connect to minio")
		return nil, err
	}
	found, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, err
	}
	if !found {
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: "us-east-1"})
		if err != nil {
			return nil, err
		}
	}
	return minioClient, nil
}

func main() {
	bucketName := "pdfs"
	minioClient, err := getMinioClient(bucketName)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Connect to minio")
	}
	urls, err := waitForURLs()
	if err != nil {
		log.Fatalln(err)
	}
	log.Info("waiting for urls")
	for url := range urls {
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
