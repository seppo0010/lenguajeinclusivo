package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		segments := strings.Split(path, "/")
		if len(segments) != 3 {
			http.NotFound(w, r)
			return
		}
		stat, err := minioClient.StatObject(context.Background(), bucketName, segments[2], minio.StatObjectOptions{})
		if err != nil {
			if err.Error() == "The specified key does not exist." {
				http.NotFound(w, r)
			} else {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
        w.Header().Add("content-size", fmt.Sprintf("%v", stat.Size))
		obj, err := minioClient.GetObject(context.Background(), bucketName, segments[2], minio.StatObjectOptions{})
		io.Copy(w, obj)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
