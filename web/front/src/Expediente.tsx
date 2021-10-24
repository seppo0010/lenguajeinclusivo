import React, { useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux'
import { fetchExpediente, selectExpediente } from './expedienteSlice'
import type { RootState } from './store'

function Expediente({id}: {id: string}) {
  const dispatch = useDispatch()
  const expediente = useSelector(selectExpediente)
  const expedienteStatus  = useSelector((state: RootState) => state.expediente.status)

  useEffect(() => {
    if (expedienteStatus === 'idle') {
      dispatch(fetchExpediente({id}))
    }
  }, [expedienteStatus, dispatch, id])
  return (
    <div>
      <p>Expediente: {expediente?.ficha.caratula}</p>
      {expediente?.actuaciones && <ul>
        {expediente?.actuaciones.map((act) => <li key={act.actId}>
          {act.titulo} - {act.firmantes} ({new Date(act.fechaFirma).toString()})
        </li>)}
      </ul>}
    </div>
  );
}

export default Expediente;
