import React, { useState, useEffect } from 'react';
import searchImg from './images/system-search-symbolic.svg';
import './App.css';
import Box from './Box';
import Section from './Section';
import Button from './Button';

import { ExpedienteURL, ExpedienteLoader } from './Expediente';

function App() {
  const [search, setSearch] = useState("");
  const [expedientes, setExpedientes] = useState<ExpedienteURL[]>([]);
  const [selected, setSelected] = useState<ExpedienteURL>();

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    event.preventDefault();
    setSearch(event.target.value)
  }

  useEffect(() => {
    (async () => {
      const res = await fetch(`${process.env.PUBLIC_URL}/data/expedientes.json`);
      const json = await res.json();
      setExpedientes(json)
    })();
  }, [])

  useEffect(() => {
    if (!expedientes.length) return;
    expedientes.forEach((expediente) => {
      // warm up cache
      fetch(`${process.env.PUBLIC_URL}/data/${expediente.file}.json`)
      fetch(`${process.env.PUBLIC_URL}/data/${expediente.file}-index.json`)
    })
  }, [expedientes])

  return (
    <div className="App">
      <Box id="title">
        <h1>Buscador de Actuaciones de O.D.I.A.</h1>
      </Box>
      <Section title="buscar">
        <input type="search" id="search-input" name="search"
          onChange={handleChange} value={search} placeholder="buscar..." />
        <img src={searchImg} alt="" />
      </Section>
      <Section title="expedientes">
        {(expedientes.length)
          ? expedientes.map(({ id }, i) =>
            <Button key={i}
              onClick={(e: React.MouseEvent<HTMLElement>) => setSelected(expedientes[i])}
              selected={selected === expedientes[i]}>
              {id}
            </Button>)
          : <p>loading...</p>}
        {selected !== undefined && <ExpedienteLoader {...selected} search={search} />}
      </Section>
    </div >
  );
}

export default App;
