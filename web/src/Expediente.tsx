import React, { useState, useEffect } from 'react';
import Section from './Section';
import Actuacion, { ActuacionData } from './Actuacion';

import { MiniSearchConfig } from './ms'
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
}

export interface SearchData {
  term?: string
  minisearch: MiniSearch
}

export interface ExpedienteProps {
  data: ExpedienteData
  search: SearchData
}

export function ExpedienteLoader({ file, search }: ExpedienteURL) {
  const [data, setData] = useState<ExpedienteData>();
  const [error, setError] = useState<string>();

  const [minisearch, setMinisearch] = useState<MiniSearch>();

  useEffect(() => {
    setError(undefined)
    setData(undefined)
    fetch(`${process.env.PUBLIC_URL}/data/${file}.json`)
      .then(res => res.json())
      .then(setData)
      .catch(e => {
        setError(error => error + 'DATA: ' + e.toString());
      })
  }, [file])

  useEffect(() => {
    setError(undefined)
    fetch(`${process.env.PUBLIC_URL}/data/${file}-index.json`)
      .then(res => res.text())
      .then(json => {
        const ms = MiniSearch.loadJSON(json, MiniSearchConfig)
        setMinisearch(ms)
      })
      .catch(e => {
        setError(error => error + 'INDEX:' + e.toString());
      })
  }, [file])

  if (error) return (<h2>{error}</h2>)
  if (data && minisearch) return (<Expediente data={data} search={{ term: search, minisearch }} />)
  return (<p>loading ...</p>)
}

function Expediente({ data: { Actuaciones, caratula }, search }: ExpedienteProps) {
  const [actuaciones, setActuaciones] = useState<ActuacionData[]>()

  useEffect(() => {
    if (!search.term) {
      setActuaciones(Actuaciones)
      return
    }
    const res = search.minisearch.search(search.term, {
      prefix: term => term.length > 3,
      fuzzy: term => term.length > 3 ? 0.2 : false
    })

    const actsId: number[] = []
    const docsURL: string[] = []
    for (let i = 0; i < res.length; i++) {
      actsId.push(res[i].actId);
      docsURL.push(res[i].URL)
    }

    setActuaciones(Actuaciones
      .filter((act) => actsId.includes(act.actId))
      .map(({ documentos, ...act }) => ({
        ...act,
        documentos: documentos.filter((d) => docsURL.includes(d.URL))
      }))
    )
  }, [Actuaciones, search])

  return (
    <Section title={caratula}>
      {
        actuaciones && actuaciones.map((a) => (
          <Actuacion key={a.actId} {...a} />
        ))
      }
    </Section>
  )
}
export default Expediente
