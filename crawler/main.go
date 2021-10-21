package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/streadway/amqp"
)

var fichaType = "ficha"
var actuacionType = "actuación"

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

type ActuacionWithExpediente struct {
	Actuacion
	NumeroDeExpediente string
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

	ficha := Ficha{}
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
	size := 20
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

func (t *LoggerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	lr := req.Context().Value(LogRequest("log"))
	if lr != nil && lr.(bool) == true {
		requestDump, err := httputil.DumpRequest(req, true)
		if err == nil {
			log.Printf("Request: %s", string(requestDump))
		}
	}

	return t.transport.RoundTrip(req)
}

func insertExpediente(exp *Expediente) error {
	var err error
	var wg sync.WaitGroup
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			"http://es:9200",
		},
		Transport: &LoggerTransport{transport: http.DefaultTransport},
	})
	if err != nil {
		return err
	}

	wg.Add(1)
	go func(ficha *Ficha) {
		log.Printf("saving ficha")
		defer log.Printf("finished saving ficha")
		defer wg.Done()
		r, w := io.Pipe()
		enc := json.NewEncoder(w)
		go func() {
			defer w.Close()
			enc.Encode(ficha)
		}()
		res, innerErr := esapi.IndexRequest{
			Index:      fichaType,
			DocumentID: fmt.Sprintf("%d-%d", ficha.Numero, ficha.Anio),
			Body:       r,
			Refresh:    "true",
			Pretty:     true,
			Human:      true,
		}.Do(context.WithValue(context.Background(), LogRequest("log"), false), es)
		if innerErr != nil {
			err = innerErr
			return
		}
		defer res.Body.Close()

		if res.IsError() {
			err = fmt.Errorf("%s", res.Status())
			return
		}
	}(exp.Ficha)

	for i, actuacion := range exp.Actuaciones {
		wg.Add(1)
		go func(i int, actuacion Actuacion) {
			log.Printf("saving actuacion %d", i)
			defer log.Printf("finished saving actuacion %d", i)
			defer wg.Done()
			r, w := io.Pipe()
			enc := json.NewEncoder(w)
			go func() {
				defer w.Close()
				enc.Encode(ActuacionWithExpediente{
					Actuacion:          actuacion,
					NumeroDeExpediente: fmt.Sprintf("%d/%d", exp.Ficha.Numero, exp.Ficha.Anio),
				})
			}()
			res, innerErr := esapi.IndexRequest{
				Index:      actuacionType,
				DocumentID: fmt.Sprintf("%d", actuacion.ActId),
				Body:       r,
				Refresh:    "true",
				Pretty:     true,
				Human:      true,
			}.Do(context.WithValue(context.Background(), LogRequest("log"), false), es)
			if innerErr != nil {
				err = innerErr
				return
			}
			defer res.Body.Close()

			if res.IsError() {
				err = fmt.Errorf("%s", res.Status())
				return
			}
		}(i, actuacion)
	}
	wg.Wait()

	return err
}

func waitForExpediente() (<-chan (string), error) {
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
	_, err = c.QueueDeclare("crawl", true, false, false, false, nil)
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
