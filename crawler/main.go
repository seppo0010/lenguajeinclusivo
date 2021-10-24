package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/olivere/elastic"
	"github.com/seppo0010/juscaba/shared"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const actuacionMapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
		"properties":{
			"codigo":{
				"type":"keyword"
			},
			"actuacionesNotificadas":{
				"type":"keyword"
			},
			"firmantes":{
				"type":"keyword"
			},
			"titulo":{
				"type":"keyword"
			},
			"cuij":{
				"type":"keyword"
			},
			"numeroDeExpediente":{
				"type":"keyword"
			}
		}
	}
}`

const documentMapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
		"properties":{
			"URL":{
				"type":"keyword"
			}
		}
	}
}`

type SearchFormFilter struct {
	Identificador string `json:"identificador"`
}
type SearchForm struct {
	Filter       string `json:"filter"`
	TipoBusqueda string `json:"tipoBusqueda"`
	Page         int    `json:"page"`
	Size         int    `json:"size"`
}

type SearchResultContent struct {
	ExpId int `json:"expId"`
}

type SearchResult struct {
	Content []SearchResultContent `json:"content"`
}

func getExpedienteCandidates(criteria string) ([]int, error) {
	filter, _ := json.Marshal(SearchFormFilter{
		Identificador: criteria,
	})
	info, _ := json.Marshal(SearchForm{
		Filter:       string(filter),
		TipoBusqueda: "CAU",
		Page:         0,
		Size:         10,
	})

	u := "https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/lista"
	resp, err := http.PostForm(u, url.Values{
		"info": {string(info)},
	})
	if err != nil {
		log.WithFields(log.Fields{
			"expediente": criteria,
			"url":        u,
		}).Warn("Failed to get expediente")
		return nil, err
	}
	defer resp.Body.Close()

	sr := SearchResult{}
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		log.WithFields(log.Fields{
			"expediente": criteria,
			"url":        u,
			"httpStatus": resp.StatusCode,
		}).Warn("Failed to decode json")
		return nil, err
	}
	res := make([]int, len(sr.Content))
	for i, s := range sr.Content {
		res[i] = s.ExpId
	}
	return res, nil
}

func getFicha(candidate int) (*shared.Ficha, error) {
	u := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/ficha?expId=%d", candidate)
	resp, err := http.Get(u)
	if err != nil {
		log.WithFields(log.Fields{
			"expId": candidate,
			"url":   u,
		}).Warn("Failed to get ficha")
		return nil, err
	}
	defer resp.Body.Close()

	ficha := shared.Ficha{ExpId: candidate}
	err = json.NewDecoder(resp.Body).Decode(&ficha)
	if err != nil {
		log.WithFields(log.Fields{
			"expId":      candidate,
			"url":        u,
			"httpStatus": resp.StatusCode,
		}).Warn("Failed to decode json")
		return nil, err
	}
	return &ficha, nil
}

func searchExpediente(criteria string) (*shared.Expediente, error) {
	candidates, err := getExpedienteCandidates(criteria)
	if err != nil {
		return nil, err
	}

	for _, candidate := range candidates {
		ficha, err := getFicha(candidate)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(criteria, fmt.Sprintf("%d/%d", ficha.Numero, ficha.Anio)) {
			log.WithFields(log.Fields{
				"expediente": criteria,
			}).Info("Expediente found!")

			actuaciones, err := getActuaciones(candidate)
			if err != nil {
				return nil, err
			}
			return &shared.Expediente{
				Ficha:       ficha,
				Actuaciones: actuaciones,
			}, nil
		}
	}
	log.WithFields(log.Fields{
		"expediente": criteria,
	}).Info("cannot find expediente")
	return nil, fmt.Errorf("cannot find ficha for criteria: %s", criteria)
}

func getActuacionesPage(expId int, pagenum int) (*shared.ActuacionesPage, error) {
	log.WithFields(log.Fields{
		"page": pagenum,
	}).Info("getting actuaciones")
	size := 100
	u := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones?filtro=%%7B%%22cedulas%%22%%3Atrue%%2C%%22escritos%%22%%3Atrue%%2C%%22despachos%%22%%3Atrue%%2C%%22notas%%22%%3Atrue%%2C%%22expId%%22%%3A%d%%2C%%22accesoMinisterios%%22%%3Afalse%%7D&page=%d&size=%d",
		expId,
		pagenum,
		size,
	)
	res, err := http.Get(u)
	if err != nil {
		log.WithFields(log.Fields{
			"expId":   expId,
			"pagenum": pagenum,
			"url":     u,
		}).Warn("Failed to get actuaciones")
		return nil, err
	}
	page := shared.ActuacionesPage{}
	err = json.NewDecoder(res.Body).Decode(&page)
	if err != nil {
		log.WithFields(log.Fields{
			"expId":      expId,
			"pagenum":    pagenum,
			"url":        u,
			"httpStatus": res.StatusCode,
		}).Warn("Failed to decode json")
		return nil, err
	}
	return &page, nil
}

func getActuaciones(expId int) ([]shared.Actuacion, error) {
	actuaciones := make([]shared.Actuacion, 0, 1)
	pagenum := 0
	for {
		page, err := getActuacionesPage(expId, pagenum)
		if err != nil {
			return nil, err
		}
		if len(page.Content) == 0 {
			break
		}
		actuaciones = append(actuaciones, page.Content...)
		pagenum++
	}
	return actuaciones, nil
}

func insertAdjuntosCedula(es *elastic.Client, c *amqp.Channel, url string, ficha *shared.Ficha, actuacion *shared.Actuacion) {
	u := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/cedulas/adjuntos?filter=%%7B%%22cedulaCuij%%22:%%22%v%%22,%%22expId%%22:%v,%%22ministerios%%22:false%%7D",
		actuacion.CUIJ,
		ficha.ExpId,
	)
	resp, err := http.Get(u)
	if err != nil {
		log.WithFields(log.Fields{
			"actId": actuacion.ActId,
			"url":   u,
		}).Warn("Failed to get adjuntos")
		return
	}
	defer resp.Body.Close()

	adjuntos := []map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&adjuntos)
	if err != nil {
		log.WithFields(log.Fields{
			"actId":      actuacion.ActId,
			"url":        u,
			"httpStatus": resp.StatusCode,
		}).Warn("Failed to decode json")
		return
	}
	for _, adjunto := range adjuntos {
		url := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/cedulas/adjuntoPdf?filter=%%7B%%22aacId%%22:%v,%%22expId%%22:%v,%%22ministerios%%22:false%%7D",
			adjunto["adjuntoId"],
			ficha.ExpId,
		)
		insertDocument(es, c, url, ficha, actuacion, shared.CedulaAttachment, adjunto["adjuntoNombre"].(string))
	}
}
func insertAdjuntosNoCedula(es *elastic.Client, c *amqp.Channel, url string, ficha *shared.Ficha, actuacion *shared.Actuacion) {
	u := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/adjuntos?actId=%v&expId=%v&accesoMinisterios=false",
		actuacion.ActId,
		ficha.ExpId,
	)
	resp, err := http.Get(u)
	if err != nil {
		log.WithFields(log.Fields{
			"actId": actuacion.ActId,
			"url":   u,
		}).Warn("Failed to get adjuntos")
		return
	}
	defer resp.Body.Close()

	adjuntos := map[string][]map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&adjuntos)
	if err != nil {
		log.WithFields(log.Fields{
			"httpStatus": resp.StatusCode,
			"actId":      actuacion.ActId,
			"url":        u,
		}).Warn("Failed to decode json")
		return
	}
	for _, adjunto := range adjuntos["adjuntos"] {
		url := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/adjuntoPdf?filter=%%7B%%22aacId%%22:%v,%%22expId%%22:%v,%%22ministerios%%22:false%%7D",
			adjunto["adjId"],
			ficha.ExpId,
		)
		insertDocument(es, c, url, ficha, actuacion, shared.AdjuntosAttachment, adjunto["titulo"].(string))
	}
}

func insertAdjuntos(es *elastic.Client, c *amqp.Channel, url string, ficha *shared.Ficha, actuacion *shared.Actuacion) {
	if actuacion.EsCedula == 1 {
		insertAdjuntosCedula(es, c, url, ficha, actuacion)
	} else {
		insertAdjuntosNoCedula(es, c, url, ficha, actuacion)
	}
}

func insertDocument(es *elastic.Client, c *amqp.Channel, url string, ficha *shared.Ficha, actuacion *shared.Actuacion, typ int, nombre string) {
	_, err := es.Index().
		Index(shared.DocumentType).
		Type("_doc").
		OpType("create").
		Id(url).
		BodyJson(map[string]interface{}{
			"URL":                url,
			"actuacionId":        actuacion.Id(),
			"numeroDeExpediente": fmt.Sprintf("%d/%d", ficha.Numero, ficha.Anio),
			"type":               typ,
			"nombre":             nombre,
		}).
		Do(context.Background())

	if err != nil && !elastic.IsConflict(err) {
		log.WithFields(log.Fields{
			"ficha":     ficha.Id(),
			"actuacion": actuacion.ActId,
		}).Error(err.Error())
	}
	err = c.Publish(
		"tasks",
		"fetch",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(url),
		})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to publish fetch task")
	}
}

func insertExpediente(es *elastic.Client, exp *shared.Expediente) error {
	var err error
	var wg sync.WaitGroup
	c, err := shared.InitTaskQueue("fetch")
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"ficha": exp.Ficha.Id(),
	}).Info("saving ficha")
	_, err = es.Index().
		Index(shared.FichaType).
		Type(shared.FichaType).
		Id(exp.Ficha.Id()).
		BodyJson(exp.Ficha).
		Do(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"ficha": exp.Ficha.Id(),
		}).Error(err.Error())
		return err
	}

	for i, actuacion := range exp.Actuaciones {
		wg.Add(1)
		go func(i int, actuacion shared.Actuacion) {
			defer wg.Done()
			log.WithFields(log.Fields{
				"ficha":     exp.Ficha.Id(),
				"actuacion": actuacion.ActId,
			}).Info("saving actuacion")
			url := fmt.Sprintf(
				"https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/pdf?datos=%%7B%%22actId%%22:%d,%%22expId%%22:%d,%%22esNota%%22:false,%%22cedulaId%%22:null,%%22ministerios%%22:false%%7D",
				actuacion.ActId,
				exp.Ficha.ExpId,
			)
			_, innerErr := es.Index().
				Index(shared.ActuacionType).
				Type("_doc").
				OpType("create").
				Id(actuacion.Id()).
				BodyJson(shared.ActuacionWithExpediente{
					Actuacion:          actuacion,
					NumeroDeExpediente: fmt.Sprintf("%d/%d", exp.Ficha.Numero, exp.Ficha.Anio),
				}).
				Do(context.Background())
			if innerErr != nil && !elastic.IsConflict(innerErr) {
				log.WithFields(log.Fields{
					"ficha":     exp.Ficha.Id(),
					"actuacion": actuacion.ActId,
				}).Error(innerErr.Error())
				return
			}
			insertDocument(es, c, url, exp.Ficha, &actuacion, shared.RegularAttachment, "")
			if actuacion.ActuacionesNotificadas != "" {

				insertDocument(es, c, fmt.Sprintf(
					"https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/pdf?datos=%%7B%%22actId%%22:%%22%v%%22,%%22expId%%22:%v,%%22esNota%%22:false,%%22cedulaId%%22:%v,%%22ministerios%%22:false%%7D",
					actuacion.ActuacionesNotificadas,
					exp.ExpId,
					actuacion.ActId,
				), exp.Ficha, &actuacion, shared.ActuacionesNotificadasAttachment, "")
			}
			if actuacion.PoseeAdjunto > 0 {
				insertAdjuntos(es, c, url, exp.Ficha, &actuacion)
			}

		}(i, actuacion)
	}
	wg.Wait()

	return err
}

func waitForExpediente() (<-chan (string), error) {
	c, err := shared.InitTaskQueue("crawl")
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
	tasks, err := c.Consume("crawl", "crawler", false, false, false, false, nil)
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

func installMappings(es *elastic.Client) error {
	exists, err := es.IndexExists(shared.ActuacionType).Do(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to check for index")
		return err
	}
	if !exists {
		createIndex, err := es.CreateIndex(shared.ActuacionType).BodyString(actuacionMapping).Do(context.Background())
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to create index")
			return err
		}
		if !createIndex.Acknowledged {
			return fmt.Errorf("did not ack index creation")
		}
	}

	exists, err = es.IndexExists(shared.DocumentType).Do(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to check for index")
		return err
	}
	if !exists {
		createIndex, err := es.CreateIndex(shared.DocumentType).BodyString(documentMapping).Do(context.Background())
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to create index")
			return err
		}
		if !createIndex.Acknowledged {
			return fmt.Errorf("did not ack index creation")
		}
	}
	return nil
}

func main() {
	es, err := elastic.NewClient(
		elastic.SetURL("http://es:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to connect to elastic")
	}
	err = installMappings(es)
	if err != nil {
		log.Fatal(err.Error())
	}

	expedientes, err := waitForExpediente()
	if err != nil {
		log.Fatalf("%s\n", err)
		return
	}
	log.Info("waiting for expedientes")
	for numexpediente := range expedientes {
		log.Printf("received expediente %s", numexpediente)
		log.WithFields(log.Fields{
			"expediente": numexpediente,
		}).Info("Received expediente")
		exp, err := searchExpediente(numexpediente)
		if err != nil {
			continue
		}
		err = insertExpediente(es, exp)
		if err != nil {
			continue
		}
		log.WithFields(log.Fields{
			"expediente": numexpediente,
		}).Info("Finished expediente")
	}
}
