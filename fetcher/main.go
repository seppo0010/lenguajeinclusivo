package main

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/streadway/amqp"
)

func getSha1(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func start(minioClient *minio.Client, bucketName string) error {
	found, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		return err
	}
	if !found {
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: "us-east-1"})
		if err != nil {
			return err
		}
	}
	return nil
}

func urlIsSaved(url string, minioClient *minio.Client, bucketName string) bool {
	_, err := minioClient.StatObject(context.Background(), bucketName, getSha1(url), minio.StatObjectOptions{})
	return err == nil
}

func saveURLToBucket(url string, minioClient *minio.Client, bucketName string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		uploadInfo, err := minioClient.PutObject(context.Background(), bucketName, getSha1(url), res.Body, -1, minio.PutObjectOptions{ContentType: res.Header.Get("content-type")})
		if err != nil {
			return err
		}
		log.Printf("uploaded %s %#v\n", url, uploadInfo)
	} else {
		log.Printf("did not save %s because status code was %d", url, res.StatusCode)
	}

	return nil
}

func waitForURLs() (<-chan (string), error) {
	conn, err := amqp.Dial("amqp://queues/")
	if err != nil {
		return nil, err
	}

	c, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	err = c.ExchangeDeclare("tasks", "direct", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	_, err = c.QueueDeclare("fetch", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	err = c.QueueBind("fetch", "fetch", "tasks", false, nil)
	if err != nil {
		return nil, err
	}

	err = c.Qos(1, 0, false)
	if err != nil {
		return nil, err
	}
	tasks, err := c.Consume("fetch", "fetcher", false, false, false, false, nil)
	if err != nil {
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

func main() {
	endpoint := "minio:9000"
	accessKeyID := "00c54c8a5e6a0e3eb04801d0f1b04425"
	secretAccessKey := "994381299ca8033f9f4786332bcd692072daec65"
	useSSL := false
	bucketName := "pdfs"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}
	err = start(minioClient, bucketName)
	if err != nil {
		log.Fatalln(err)
	}
	urls, err := waitForURLs()
	if err != nil {
		log.Fatalln(err)
	}
    log.Printf("waiting for urls")
	for url := range urls {
		log.Printf("received url: %s", url)
		exists := urlIsSaved(url, minioClient, bucketName)
		if exists {
			log.Printf("url %s already exists", url)
			continue
		}
		err = saveURLToBucket(url, minioClient, bucketName)
		if err != nil {
			log.Println(err)
		}
	}
}
