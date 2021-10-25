package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
	"github.com/minio/minio-go/v7"

	"github.com/seppo0010/juscaba/shared"
)

type miniDocumento struct {
	Nombre string `json:"nombre"`
	URL    string
}

type actuacionWithDocumentos struct {
	*shared.Actuacion
	Documentos []*miniDocumento `json:"documentos"`
}

var es *elastic.Client
var minioClient *minio.Client
const bucketName = "pdfs"

func main() {
	var err error
	es, err = elastic.NewClient(
		elastic.SetURL("http://es:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to connect to elastic")
	}

	minioClient, err = shared.GetMinioClient(bucketName)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.GET("/api/expediente/:id", getExpedienteByID)
	router.GET("/download/:hash", download)

	router.Run("localhost:8080")
}

func getFichaByID(id string) (*shared.Ficha, error) {
	obj, err := es.Get().
		Index(shared.FichaType).
		Type(shared.FichaType).
		Id(shared.FichaID(id)).
		Do(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"id":  id,
			"err": err.Error(),
		}).Info("failed to get expediente")
		return nil, err
	}
	var f *shared.Ficha
	err = json.Unmarshal(obj.Source, &f)
	if err != nil {
		log.WithFields(log.Fields{
			"id":  id,
			"obj": obj,
			"err": err.Error(),
		}).Debug("failed to decode json")
		return nil, err
	}
	return f, nil
}

func getExpedienteActuaciones(ficha *shared.Ficha) ([]*shared.Actuacion, error) {
	res, err := es.Search(shared.ActuacionType).
		Query(elastic.NewTermQuery("numeroDeExpediente", ficha.NumeroDeExpediente("/"))).
        From(0).Size(10000).
		Do(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err.Error(),
			"expediente": ficha.NumeroDeExpediente("/"),
		}).Error("failed to get actuaciones")
		return nil, err
	}

	hits := res.Hits.Hits
	actuaciones := make([]*shared.Actuacion, len(hits))
	for i, h := range hits {
		var actuacion *shared.Actuacion
		err = json.Unmarshal(h.Source, &actuacion)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err.Error(),
				"expediente": ficha.NumeroDeExpediente("/"),
			}).Error("failed to extract hit data")
			return nil, err
		}
		actuaciones[i] = actuacion
	}
	return actuaciones, nil
}

func getActuacionesWithDocumentos(ficha *shared.Ficha, actuaciones []*shared.Actuacion) ([]*actuacionWithDocumentos, error) {
	res, err := es.Search(shared.DocumentType).
		Query(elastic.NewTermQuery("numeroDeExpediente", ficha.NumeroDeExpediente("/"))).
        From(0).Size(10000).
		Do(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err.Error(),
			"expediente": ficha.NumeroDeExpediente("/"),
		}).Error("failed to get actuaciones")
		return nil, err
	}

	hits := res.Hits.Hits
	documentsByActuacionID := map[string][]*shared.Documento{}
	for _, h := range hits {
		var documento *shared.Documento
		err = json.Unmarshal(h.Source, &documento)
		if err != nil {
			log.WithFields(log.Fields{
				"hit":        fmt.Sprintf("%#v", h),
				"error":      err.Error(),
				"json":       string(h.Source),
				"expediente": ficha.NumeroDeExpediente("/"),
			}).Error("failed to extract hit data")
			return nil, err
		}
		documentsByActuacionID[documento.ActuacionID] = append(
			documentsByActuacionID[documento.ActuacionID],
			documento,
		)
	}

	awd := make([]*actuacionWithDocumentos, len(actuaciones))
	for i, a := range actuaciones {
		md := make([]*miniDocumento, len(documentsByActuacionID[a.Id()]))
		for j, m := range documentsByActuacionID[a.Id()] {
			md[j] = &miniDocumento{
				Nombre: m.Nombre,
				URL:    m.GetURL(),
			}
		}
		awd[i] = &actuacionWithDocumentos{
			Actuacion:  a,
			Documentos: md,
		}
	}
	return awd, nil
}

func getExpedienteByID(c *gin.Context) {
	id := c.Param("id")

	ficha, err := getFichaByID(id)
	if err != nil {
		if elastic.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"message": "ficha not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "error getting ficha"})
		}
		return
	}
	actuaciones, err := getExpedienteActuaciones(ficha)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error getting actuaciones"})
		return
	}

	awd, err := getActuacionesWithDocumentos(ficha, actuaciones)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error getting actuaciones' documents"})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"ficha":       ficha,
		"actuaciones": awd,
	})
}

func download(c *gin.Context) {
	hash := c.Param("hash")
	stat, err := minioClient.StatObject(context.Background(), bucketName, hash, minio.StatObjectOptions{})
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			c.JSON(http.StatusNotFound, gin.H{"message": "document not found"})
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}
	r, err := minioClient.GetObject(context.Background(), bucketName, hash, minio.StatObjectOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"hash": hash,
		}).Error("failed to get object")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
	}
	mimeType := "application/json" // FIXME: fetch api returns wrong type, inference?
	c.DataFromReader(http.StatusOK, stat.Size, mimeType, r, map[string]string{})
}
