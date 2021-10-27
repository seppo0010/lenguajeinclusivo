import React, { useState, useEffect } from 'react';
import searchImg from './images/system-search-symbolic.svg';
import './App.css';
import Box from './Box';
import Section from './Section';
import Button from './Button';

import Expediente, { ExpedienteData } from './Expediente';

interface ExpedienteURL {
  id: string
  file: string
  selected: boolean
}

interface ExpDataMap {
  [id: string]: ExpedienteData
}

function App() {
  const [search, setSearch] = useState("");
  const [expedientes, setExpedientes] = useState<ExpedienteURL[]>([]);
  const [data, setData] = useState<ExpDataMap>();
  const [selected, setSelected] = useState<ExpedienteData>();

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    event.preventDefault();
    setSearch(event.target.value)
  }

  const handleSelect = (e: ExpedienteData) => (event: React.MouseEvent<HTMLLinkElement>) => {
    event.preventDefault();
    setSelected(e)
  }

  useEffect(() => {
    fetch(`${process.env.PUBLIC_URL}/data/expedientes.json`)
      .then(resp => resp.json())
      .then(setExpedientes)
  }, [])

  useEffect(() => {
    if (!expedientes.length) return;
    for (let i = 0; i < expedientes.length; i++) {
      fetch(`${process.env.PUBLIC_URL}/data/${expedientes[i].file}.json`)
        .then(resp => resp.json())
        .then(d => setData({ ...data, [d.ExpId]: d }))
    }
  }, [expedientes])

  return (
    <div className="App">
      <Box id="title">
        <h1>Buscador de Actuaciones de O.D.I.A.</h1>
      </Box>
      <Section title="expedientes">
        {(expedientes.length && data)
          ? expedientes.map(({ id }, i) =>
            <Button key={i}
              onClick={(e: React.MouseEvent<HTMLElement>) => setSelected(data[id])}
              selected={selected === data[i]}>
              {id}
            </Button>)
          : <p>loading...</p>}
        {selected !== undefined && <Expediente {...selected} />}
      </Section>
      <div title="buscar">
        <input type="search" id="search-input" name="search"
          onChange={handleChange} value={search} placeholder="buscar..." />
        <img src={searchImg} />
        {search}
      </div>
    </div >
  );
}

export default App;
