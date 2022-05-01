import React from 'react';
import Document, { DocumentData } from './Document';
import { SearchData } from './Search';
import Section from './Section';
import Highlighter from './Highlighter';

export interface ActuacionData {
  actId: number
  anio: number
  cuij: string
  documentos: DocumentData[]
  titulo: string
  fechaFirma: number
}

export interface ActuacionProps {
  data: ActuacionData
  search: SearchData
}
function Actuacion({ data, search }: React.PropsWithChildren<ActuacionProps>) {
  return (
    <Section title={<Highlighter text={`${new Date(data.fechaFirma).toLocaleDateString('es-AR')} - ${data.titulo}`} term={search.term} />}>
      {data.documentos.map((d) => <Document key={d.URL} data={d} search={search} />)}
    </Section>
  )
}
export default Actuacion
