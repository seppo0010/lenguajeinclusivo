import React, { useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux'
import { fetchExpediente, selectExpediente } from './expedienteSlice'
import type { RootState } from './store'

function Expediente({id}: {id: string}) {
  const dispatch = useDispatch()
  const expediente = useSelector(selectExpediente)
  const expedienteStatus  = useSelector((state: RootState) => state.expediente.status)
  console.log(expediente)

  useEffect(() => {
    if (expedienteStatus === 'idle') {
      dispatch(fetchExpediente({id}))
    }
  }, [expedienteStatus, dispatch, id])
  return (
    <div>
      <p>Expediente: {expediente?.ficha.caratula}</p>
    </div>
  );
}

export default Expediente;
