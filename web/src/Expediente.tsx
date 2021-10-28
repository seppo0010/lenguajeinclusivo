import React, { useState, useEffect } from 'react';
import Section from './Section';
import Actuacion, { ActuacionData } from './Actuacion';

import MiniSearch from 'minisearch'

export interface ExpedienteURL {
  id: string
  file: string
  selected?: boolean
  search?: string
}

export interface ExpedienteData {
  expId: number
  caratula: string
  Actuaciones: ActuacionData[]
  search?: string
  minisearch: MiniSearch
}

export function ExpedienteLoader(props: ExpedienteURL) {
  const [data, setData] = useState<ExpedienteData>();
  const [error, setError] = useState<string>();
  const [errorIndex, setErrorIndex] = useState<string>();

  const [minisearch, setMinisearch] = useState<MiniSearch | undefined>();

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

  useEffect(() => {
    setErrorIndex(undefined)
    fetch(`${process.env.PUBLIC_URL}/data/${props.file}-index.json`)
      .then(res => res.text())
      .then(json => {
        // FIXME: duplicates create-index.ts
        const ms = MiniSearch.loadJSON(json, {
          fields: ['content', 'URL'],
          storeFields: ['content', 'URL', 'ExpId'],
          idField: 'numeroDeExpediente'
        })
        setMinisearch(ms)
      })
      .catch(e => {
        setErrorIndex(e.toString());
      })
  }, [props])

  if (error || errorIndex) return (<h2>{error || errorIndex}</h2>)
  if (data && minisearch) return (<Expediente {...data} search={props.search} minisearch={minisearch}/>)
  return (<p>loading ...</p>)
}

function Expediente(props: ExpedienteData) {
  const [actuaciones, setActuaciones] = useState<ActuacionData[]>([])

  useEffect(() => {
    if (!props.search) {
      setActuaciones(props.Actuaciones)
      return
    }
    const res = props.minisearch.search(props.search)
    const actsId = res.map((r) => r.actId)
    const docsURL = res.map((r) => r.URL)
    setActuaciones(props.Actuaciones
      .filter((act) => actsId.includes(act.actId))
      .map((act) => ({
        ...act,
        documentos: act.documentos.filter((d) => docsURL.includes(d.URL))
      }))
    )
  }, [props])

  return (
    <Section title={props.caratula}>
      {
        actuaciones.map((a, i) => (
          <Actuacion key={i} {...a} />)
        )
      }
    </Section>
  )
}
export default Expediente
