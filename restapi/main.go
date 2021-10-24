package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"

	"github.com/seppo0010/juscaba/shared"
)

var es *elastic.Client

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

	router := gin.Default()
	router.GET("/expediente/:id", getExpedienteByID)

	router.Run("localhost:8080")
}

func doGetExpedienteByID(id string) (*shared.Ficha, error) {
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
func getExpedienteByID(c *gin.Context) {
	id := c.Param("id")

	expediente, err := doGetExpedienteByID(id)
	if err != nil {
		if elastic.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"message": "expediente not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "expediente not found"})
		}
		return
	}
	c.IndentedJSON(http.StatusOK, expediente)
}
