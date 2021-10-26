import React, { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux'
import { fetchExpediente, fetchExpedienteSearch, clearSearch } from './expedienteSlice'
import ActuacionView from './ActuacionView'
import type { RootState } from './store'

function Expediente({id}: {id: string}) {
  const dispatch = useDispatch()

  const [isSearch, setIsSearch] = useState(false)
  const whichExpediente = isSearch ? 'search' : 'base'
  const expediente = useSelector((state: RootState) => state.expediente[whichExpediente].expediente)
  const expedienteStatus  = useSelector((state: RootState) => state.expediente[whichExpediente].status)
  const [firmantesFilter, setFirmantesFilter] = useState('')
  const [searchCriteria, setSearchCriteria] = useState('')

  const firmantes = expediente?.actuaciones
    .map((a) => a.firmantes)
    .filter((x, y, arr) => arr.indexOf(x) === y)
    .sort()

  useEffect(() => {
    if (expedienteStatus === 'idle') {
      dispatch(fetchExpediente({id}))
    }
  }, [expedienteStatus, dispatch, id])

  const search = () => {
    dispatch(fetchExpedienteSearch({id, criteria: searchCriteria}))
    setIsSearch(true)
  }

  const doClearSearch = () => {
    dispatch(clearSearch())
    setIsSearch(false)
  }

  const actuaciones = expediente?.actuaciones.filter(
    (a) => a.firmantes === firmantesFilter || firmantesFilter === ''
  ) || []

  return (
    <div>
      {expediente &&
      <>
        <p>Expediente: {expediente?.ficha.caratula}</p>
        <p>
          <label>
            Buscar:
            <input type="text" value={searchCriteria} onChange={(e) => setSearchCriteria(e.target.value)} />
          </label>
          {isSearch ?
            <button onClick={() => doClearSearch()}>Ver todos</button> :
            <button onClick={() => search()}>Buscar</button>
          }
        </p>
        <p>
          <label>
            Firmantes:
            <select onChange={(e) => setFirmantesFilter(e.target.value)}>
              <option value="">Todos</option>
              {firmantes?.map((f) => (
                <option key={f} value={f}>
                  {f}
                </option>
              ))}
            </select>
          </label>
        </p>
        {actuaciones && <ul>
          {actuaciones.map((act) => <li key={act.actId}>
            <ActuacionView actuacion={act} />
          </li>)}
        </ul>}
      </>}
    </div>
  );
}


export default Expediente;
