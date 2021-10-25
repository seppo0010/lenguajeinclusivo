import React from 'react';
import { Actuacion } from './expedienteSlice'

function ActuacionView({actuacion}: {actuacion: Actuacion}) {
  return (
    <span>
      {actuacion.titulo} - {actuacion.firmantes} ({new Date(actuacion.fechaFirma).toString()})
      {actuacion.documentos && (
        <ul>
          {actuacion.documentos.map((d) => (
            <li key={d.URL}>
              <a target="_blank" rel="noreferrer" href={d.URL}>{d.nombre || 'Documento sin título'}</a>
            </li>
          ))}
        </ul>
      )}
    </span>
  )
}
export default ActuacionView
