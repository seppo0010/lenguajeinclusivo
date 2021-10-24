import React, { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux'
import { fetchExpediente, selectExpediente } from './expedienteSlice'
import type { RootState } from './store'

function Expediente({id}: {id: string}) {
  const dispatch = useDispatch()
  const expediente = useSelector(selectExpediente)
  const expedienteStatus  = useSelector((state: RootState) => state.expediente.status)
  const [firmantesFilter, setFirmantesFilter] = useState('')

  const firmantes = expediente?.actuaciones
    .map((a) => a.firmantes)
    .filter((x, y, arr) => arr.indexOf(x) === y)
    .sort()

  useEffect(() => {
    if (expedienteStatus === 'idle') {
      dispatch(fetchExpediente({id}))
    }
  }, [expedienteStatus, dispatch, id])

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
            {act.titulo} - {act.firmantes} ({new Date(act.fechaFirma).toString()})
          </li>)}
        </ul>}
      </>}
    </div>
  );
}

export default Expediente;
