package shared

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func GetSha1(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func ReadSecret(name string) (string, error) {
	body, err := ioutil.ReadFile(fmt.Sprintf("/run/secrets/%v", name))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"name":  name,
		}).Fatal("failed to read secret")
	}
	return strings.TrimSpace(string(body)), nil
}

func InitTaskQueue(name string) (*amqp.Channel, error) {
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

func GetMinioClient(bucketName string) (*minio.Client, error) {
	useSSL := false
	accessKeyID, err := ReadSecret("minio-user")
	if err != nil {
		return nil, err
	}
	secretAccessKey, err := ReadSecret("minio-password")
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

func GetURLContent(url string, minioClient *minio.Client, bucketName string) (io.Reader, error) {
	r, err := minioClient.GetObject(context.Background(), bucketName, GetSha1(url), minio.StatObjectOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("cannot get object")
		return nil, err
	}
	return r, nil
}

func WaitForTasks(queue, consumer string) (<-chan []byte, error) {
	c, err := InitTaskQueue(queue)
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
	tasks, err := c.Consume(queue, consumer, false, false, false, false, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to create consumer")
		return nil, err
	}
	ch := make(chan []byte)
	go func(tasks <-chan amqp.Delivery) {
		for task := range tasks {
			ch <- task.Body
			task.Ack(false)
		}
	}(tasks)
	return ch, nil
}
