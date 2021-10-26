package shared

import (
	"fmt"
)

const FichaType = "ficha"
const ActuacionType = "actuacion"
const DocumentType = "document"
const RegularAttachment = 0
const ActuacionesNotificadasAttachment = 1
const AdjuntosAttachment = 1
const CedulaAttachment = 2

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

func FichaID(expedienteID string) string {
	return fmt.Sprintf("ficha %v", expedienteID)
}

func (ficha *Ficha) NumeroDeExpediente(separator string) string {
	return fmt.Sprintf("%d%s%d", ficha.Numero, separator, ficha.Anio)
}

func (ficha *Ficha) Id() string {
	return FichaID(ficha.NumeroDeExpediente("-"))
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
	Content          []*Actuacion            `json:"content"`
}

type Actuacion struct {
	EsCedula               int          `json:"esCedula"`
	Codigo                 string       `json:"codigo"`
	ActuacionesNotificadas string       `json:"actuacionesNotificadas"`
	Numero                 int          `json:"numero"`
	FechaFirma             int          `json:"fechaFirma"`
	Firmantes              string       `json:"firmantes"`
	ActId                  int          `json:"actId"`
	Titulo                 string       `json:"titulo"`
	FechaNotificacion      int          `json:"fechaNotificacion"`
	PoseeAdjunto           int          `json:"poseeAdjunto"`
	CUIJ                   string       `json:"cuij"`
	Anio                   int          `json:"anio"`
	Documentos             []*Documento `json:"documentos"`
}

func (actuacion *Actuacion) Id() string {
	return fmt.Sprintf("actuacion %d", actuacion.ActId)
}

type ActuacionWithExpediente struct {
	Actuacion
	NumeroDeExpediente string `json:"numeroDeExpediente"`
}

type ActuacionesPagePageable struct {
	PageNumber int `json:"pageNumber"`
	PageSize   int `json:"pageSize"`
	Offset     int `json:"offset"`
}

type Expediente struct {
	*Ficha
	Actuaciones []*Actuacion
}

type Documento struct {
	URL                string
	ActuacionID        string `json:"actuacionId"`
	NumeroDeExpediente string `json:"numeroDeExpediente"`
	Type               int    `json:"type"`
	Nombre             string `json:"nombre"`
	Content            string `json:"content"`
}

func (d *Documento) GetURL() string {
	return fmt.Sprintf("/download/%v", GetSha1(d.URL))
}
