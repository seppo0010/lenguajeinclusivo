package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/olivere/elastic"
	"github.com/streadway/amqp"
)

const index = "index"
const fichaType = "ficha"
const actuacionType = "actuacion"

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

type FichaRadicaciones struct {
	SecretariaPrimeraInstancia string `json:"secretariaPrimeraInstancia"`
	OrganismoSegundaInstancia  string `json:"organismoSegundaInstancia"`
	SecretariaSegundaInstancia string `json:"secretariaSegundaInstancia"`
	OrganismoPrimeraInstancia  string `json:"organismoPrimeraInstancia"`
}

type FichaObjetosJuicio struct {
	ObjetoJuicio string `json:"objetoJuicio"`
	Categoria    string `json:"categoria"`
	EsPrincipal  int    `json:"esPrincipal"`
	Materia      string `json:"materia"`
}

type FichaUbicacion struct {
	Organismo   string `json:"organismo"`
	Dependencia string `json:"dependencia"`
}
type Ficha struct {
	ExpId            int
	Radicaciones     FichaRadicaciones    `json:"radicaciones"`
	Numero           int                  `json:"numero"`
	Anio             int                  `json:"anio"`
	Sufijo           int                  `json:"sufijo"`
	ObjetosJuicio    []FichaObjetosJuicio `json:"objetosJuicio"`
	Ubicacion        FichaUbicacion       `json:"ubicacion"`
	FechaInicio      int                  `json:"fechaInicio"`
	UltimoMovimiento int                  `json:"ultimoMovimiento"`
	TieneSentencia   int                  `json:"tieneSentencia"`
	EsPrivado        int                  `json:"esPrivado"`
	TipoExpediente   string               `json:"tipoExpediente"`
	CUIJ             string               `json:"cuij"`
	Caratula         string               `json:"caratula"`
	Monto            float64              `json:"monto"`
	Etiquetas        string               `json:"etiquetas"`
}

func (ficha *Ficha) Id() string {
	return fmt.Sprintf("ficha %d-%d", ficha.Numero, ficha.Anio)
}

type ActuacionesPage struct {
	TotalPages       int                     `json:"totalPages"`
	TotalElements    int                     `json:"totalElements"`
	NumberOfElements int                     `json:"numberOfElements"`
	Last             bool                    `json:"last"`
	First            bool                    `json:"first"`
	Size             int                     `json:"size"`
	Number           int                     `json:"number"`
	Pageable         ActuacionesPagePageable `json:"pageable"`
	Content          []Actuacion             `json:"content"`
}

type Actuacion struct {
	EsCedula               int    `json:"esCedula"`
	Codigo                 string `json:"codigo"`
	ActuacionesNotificadas string `json:"actuacionesNotificadas"`
	Numero                 int    `json:"numero"`
	FechaFirma             int    `json:"fechaFirma"`
	Firmantes              string `json:"firmantes"`
	ActId                  int    `json:"actId"`
	Titulo                 string `json:"titulo"`
	FechaNotificacion      int    `json:"fechaNotificacion"`
	PoseeAdjunto           int    `json:"poseeAdjunto"`
	CUIJ                   string `json:"cuij"`
	Anio                   int    `json:"anio"`
}

func (actuacion *Actuacion) Id() string {
	return fmt.Sprintf("actuacion %d", actuacion.ActId)
}

type ActuacionWithExpediente struct {
	Actuacion
	NumeroDeExpediente string
	URL                string
}

type ActuacionesPagePageable struct {
	PageNumber int `json:"pageNumber"`
	PageSize   int `json:"pageSize"`
	Offset     int `json:"offset"`
}

type Expediente struct {
	*Ficha
	Actuaciones []Actuacion
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

	resp, err := http.PostForm("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/lista", url.Values{
		"info": {string(info)},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	sr := SearchResult{}
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		return nil, err
	}
	res := make([]int, len(sr.Content))
	for i, s := range sr.Content {
		res[i] = s.ExpId
	}
	return res, nil
}

func getFicha(candidate int) (*Ficha, error) {
	resp, err := http.Get(fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/ficha?expId=%d", candidate))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ficha := Ficha{ExpId: candidate}
	err = json.NewDecoder(resp.Body).Decode(&ficha)
	if err != nil {
		return nil, err
	}
	return &ficha, nil
}

func searchExpediente(criteria string) (*Expediente, error) {
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
			log.Printf("expediente found!")
			actuaciones, err := getActuaciones(candidate)
			if err != nil {
				return nil, err
			}
			return &Expediente{
				Ficha:       ficha,
				Actuaciones: actuaciones,
			}, nil
		}
	}
	return nil, fmt.Errorf("cannot find ficha for criteria: %s", criteria)
}

func getActuacionesPage(expId int, pagenum int) (*ActuacionesPage, error) {
	log.Printf("getting actuaciones page %d", pagenum)
	size := 100
	res, err := http.Get(fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones?filtro=%%7B%%22cedulas%%22%%3Atrue%%2C%%22escritos%%22%%3Atrue%%2C%%22despachos%%22%%3Atrue%%2C%%22notas%%22%%3Atrue%%2C%%22expId%%22%%3A%d%%2C%%22accesoMinisterios%%22%%3Afalse%%7D&page=%d&size=%d", expId, pagenum, size))
	if err != nil {
		return nil, err
	}
	page := ActuacionesPage{}
	err = json.NewDecoder(res.Body).Decode(&page)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func getActuaciones(expId int) ([]Actuacion, error) {
	actuaciones := make([]Actuacion, 0, 1)
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

type LoggerTransport struct {
	transport http.RoundTripper
}

type LogRequest string

func initTaskQueue(name string) (*amqp.Channel, error) {
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
	_, err = c.QueueDeclare(name, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func insertExpediente(exp *Expediente) error {
	var err error
	var wg sync.WaitGroup
	es, err := elastic.NewClient(elastic.SetURL("http://es:9200"))
	if err != nil {
		return err
	}

	c, err := initTaskQueue("fetch")
	if err != nil {
		return err
	}

	log.Printf("saving ficha")
	put, err := es.Index().
		Index(fichaType).
		Type(fichaType).
		Id(exp.Ficha.Id()).
		BodyJson(exp.Ficha).
		Do(context.Background())
	if err != nil {
		return err
	}
	log.Printf("Indexed ficha %s to index %s\n", put.Id, put.Index)

	for i, actuacion := range exp.Actuaciones {
		wg.Add(1)
		go func(i int, actuacion Actuacion) {
			log.Printf("saving actuacion %d", i)
			url := fmt.Sprintf(
				"https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/pdf?datos=%%7B%%22actId%%22:%d,%%22expId%%22:%d,%%22esNota%%22:false,%%22cedulaId%%22:null,%%22ministerios%%22:false%%7D",
				actuacion.ActId,
				exp.Ficha.ExpId,
			)
			put, innerErr := es.Index().
				Index(actuacionType).
				Type(actuacionType).
				Id(actuacion.Id()).
				BodyJson(ActuacionWithExpediente{
					Actuacion:          actuacion,
					NumeroDeExpediente: fmt.Sprintf("%d/%d", exp.Ficha.Numero, exp.Ficha.Anio),
					URL:                url,
				}).
				Do(context.Background())
			defer wg.Done()
			if innerErr != nil {
				err = innerErr
				return
			}
			log.Printf("Indexed actuacion %s to index %s\n", put.Id, put.Index)
			err = c.Publish(
				"tasks",
				"fetch",
				false,
				false,
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(url),
				})
		}(i, actuacion)
	}
	wg.Wait()

	return err
}

func waitForExpediente() (<-chan (string), error) {
	c, err := initTaskQueue("crawl")
	if err != nil {
		return nil, err
	}

	err = c.QueueBind("crawl", "crawl", "tasks", false, nil)
	if err != nil {
		return nil, err
	}

	err = c.Qos(1, 0, false)
	if err != nil {
		return nil, err
	}
	tasks, err := c.Consume("crawl", "crawler", false, false, false, false, nil)
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
	expedientes, err := waitForExpediente()
	if err != nil {
		log.Fatalf("%s\n", err)
		return
	}
	log.Print("waiting for expedientes")
	for numexpediente := range expedientes {
		log.Printf("received expediente %s", numexpediente)
		exp, err := searchExpediente(numexpediente)
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		err = insertExpediente(exp)
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		log.Printf("finished with expediente %s", numexpediente)
	}
}
