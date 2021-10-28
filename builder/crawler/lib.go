package crawlern

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/odia/juscaba/shared"
	log "github.com/sirupsen/logrus"
)

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

func GetExpediente(criteria string) (*shared.Expediente, error) {
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

			actuaciones, err := getActuaciones(ficha)
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

func getActuaciones(ficha *shared.Ficha) ([]*shared.Actuacion, error) {
	actuaciones := make([]*shared.Actuacion, 0, 1)
	pagenum := 0
	for {
		page, err := getActuacionesPage(ficha.ExpId, pagenum)
		if err != nil {
			return nil, err
		}
		if len(page.Content) == 0 {
			break
		}
		actuaciones = append(actuaciones, page.Content...)
		pagenum++
	}
	for _, act := range actuaciones {
		act.Documentos, _ = fetchDocumentos(ficha, act)
	}
	return actuaciones, nil
}

func GetAdjuntosCedula(url string, ficha *shared.Ficha, actuacion *shared.Actuacion) ([]*shared.Documento, error) {
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
		return nil, err
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
		return nil, err
	}
	documentos := make([]*shared.Documento, len(adjuntos))
	for i, adjunto := range adjuntos {
		url := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/cedulas/adjuntoPdf?filter=%%7B%%22aacId%%22:%v,%%22expId%%22:%v,%%22ministerios%%22:false%%7D",
			adjunto["adjuntoId"],
			ficha.ExpId,
		)
		documentos[i] = &shared.Documento{
			URL:                url,
			ActuacionID:        actuacion.Id(),
			NumeroDeExpediente: fmt.Sprintf("%d/%d", ficha.Numero, ficha.Anio),
			Type:               shared.CedulaAttachment,
			Nombre:             adjunto["adjuntoNombre"].(string),
		}
	}
	return documentos, nil
}
func GetAdjuntosNoCedula(url string, ficha *shared.Ficha, actuacion *shared.Actuacion) ([]*shared.Documento, error) {
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
		return nil, err
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
		return nil, err
	}
	documentos := make([]*shared.Documento, len(adjuntos["adjuntos"]))
	for i, adjunto := range adjuntos["adjuntos"] {
		url := fmt.Sprintf("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/adjuntoPdf?filter=%%7B%%22aacId%%22:%v,%%22expId%%22:%v,%%22ministerios%%22:false%%7D",
			adjunto["adjId"],
			ficha.ExpId,
		)
		documentos[i] = &shared.Documento{
			URL:                url,
			ActuacionID:        actuacion.Id(),
			NumeroDeExpediente: fmt.Sprintf("%d/%d", ficha.Numero, ficha.Anio),
			Type:               shared.AdjuntosAttachment,
			Nombre:             adjunto["titulo"].(string),
		}
	}
	return documentos, nil
}

func GetAdjuntos(url string, ficha *shared.Ficha, actuacion *shared.Actuacion) ([]*shared.Documento, error) {
	if actuacion.EsCedula == 1 {
		return GetAdjuntosCedula(url, ficha, actuacion)
	} else {
		return GetAdjuntosNoCedula(url, ficha, actuacion)
	}
}

func fetchDocumentos(ficha *shared.Ficha, actuacion *shared.Actuacion) ([]*shared.Documento, error) {
	documentos := make([]*shared.Documento, 0)
	url := fmt.Sprintf(
		"https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/pdf?datos=%%7B%%22actId%%22:%d,%%22expId%%22:%d,%%22esNota%%22:false,%%22cedulaId%%22:null,%%22ministerios%%22:false%%7D",
		actuacion.ActId,
		ficha.ExpId,
	)
	documentos = append(documentos, &shared.Documento{
		URL:                url,
		ActuacionID:        actuacion.Id(),
		NumeroDeExpediente: fmt.Sprintf("%d/%d", ficha.Numero, ficha.Anio),
		Type:               shared.RegularAttachment,
		Nombre:             "",
	})
	if actuacion.ActuacionesNotificadas != "" {

		url := fmt.Sprintf(
			"https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/actuaciones/pdf?datos=%%7B%%22actId%%22:%%22%v%%22,%%22expId%%22:%v,%%22esNota%%22:false,%%22cedulaId%%22:%v,%%22ministerios%%22:false%%7D",
			actuacion.ActuacionesNotificadas,
			ficha.ExpId,
			actuacion.ActId,
		)
		documentos = append(documentos, &shared.Documento{
			URL:                url,
			ActuacionID:        actuacion.Id(),
			NumeroDeExpediente: fmt.Sprintf("%d/%d", ficha.Numero, ficha.Anio),
			Type:               shared.ActuacionesNotificadasAttachment,
			Nombre:             "",
		})
	}
	if actuacion.PoseeAdjunto > 0 {
		adjuntos, _ := GetAdjuntos(url, ficha, actuacion)
		documentos = append(documentos, adjuntos...)
	}

	return documentos, nil
}
