import React from 'react';
import Document, { DocumentData } from './Document';
import Section from './Section';

export interface ActuacionData {
  actId: number
  anio: number
  cuij: string
  documentos: DocumentData[]
  titulo: string
}

function Actuacion(props: React.PropsWithChildren<ActuacionData>) {
  return (
    <Section title={props.titulo}>
      {props.documentos.map((d) => <Document key={d.URL} {...d} />)}
    </Section>
  )
}
export default Actuacion
