import React from 'react';
import Section from './Section';
import Actuacion, { ActuacionData } from './Actuacion';

export interface ExpedienteData {
  expId: number
  caratula: string
  Actuaciones: ActuacionData[]
}

function Expediente(props: ExpedienteData) {
  console.error('render', props)
  return (
    <Section title={props.caratula}>
      {
        props.Actuaciones.map((a, i) => <Actuacion key={i} {...a} />)
      }
    </Section>
  )
}
export default Expediente
