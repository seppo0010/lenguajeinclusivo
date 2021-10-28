import React, { useState, useEffect } from 'react';
import Section from './Section';
import Actuacion, { ActuacionData } from './Actuacion';

export interface ExpedienteURL {
  id: string
  file: string
  selected?: boolean
}

export interface ExpedienteData {
  expId: number
  caratula: string
  Actuaciones: ActuacionData[]
}

export function ExpedienteLoader(props: ExpedienteURL) {
  const [data, setData] = useState<ExpedienteData>();
  const [error, setError] = useState<string>();

  useEffect(() => {
    setError(undefined)
    setData(undefined)
    fetch(`${process.env.PUBLIC_URL}/data/${props.file}.json`)
      .then(res => res.json())
      .then(json => {
        setData(json)
      })
      .catch(e => {
        setError(e.toString());
      })
  }, [props])

  if (error) return (<h2>{error}</h2>)
  if (data) return (<Expediente {...data} />)
  return (<p>loading ...</p>)
}

function Expediente(props: ExpedienteData) {
  return (
    <Section title={props.caratula}>
      {
        props.Actuaciones.map((a, i) => <Actuacion key={i} {...a} />)
      }
    </Section>
  )
}
export default Expediente
