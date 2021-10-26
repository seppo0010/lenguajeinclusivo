import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'

export const fetchExpediente = createAsyncThunk('expediente/get', async ({id}: {id: string}) => {
  const req = await fetch('/api/expediente/' + encodeURIComponent(id.replace('/', '-')))
  const res = await req.json()
  return res
})

export const fetchExpedienteSearch = createAsyncThunk('expediente/search', async ({id, criteria}: {id: string, criteria: string}) => {
  let formData = new FormData()
  formData.append("criteria", criteria)
  const req = await fetch('/api/expediente/search/' + encodeURIComponent(id.replace('/', '-')), {
    method: 'POST',
    body: formData,
  })
  const res = await req.json()
  return res
})

export declare interface Expediente {
  ficha: Ficha
  actuaciones: Actuacion[]
}

export declare interface Ficha {
  caratula: string
}

export declare interface Actuacion {
  actId: number
  titulo: string
  firmantes: string
  fechaFirma: number
  documentos: Documento[]
}

export interface FetchState {
  expediente: Expediente | null
  status: 'idle' | 'succeeded' | 'loading' | 'failed'
  error: undefined | string
}
export interface State {
  base: FetchState
  search: FetchState
}

export interface Documento {
  URL: string
  nombre: string
}

const initialState: State = {
  base: {
    expediente: null,
    status: 'idle',
    error: undefined,
  },
  search: {
    expediente: null,
    status: 'idle',
    error: undefined,
  },
}

const expedienteSlice = createSlice({
  name: 'expediente',
  initialState,
  reducers: {
    clearSearch(state) {
      state.search.status = 'idle'
      state.search.expediente = null
      state.search.error = undefined
    }
  },
  extraReducers: (builder) => {
    builder.addCase(fetchExpediente.pending, (state, action) => {
      state.base.status = 'loading'
    })
    builder.addCase(fetchExpediente.fulfilled, (state, action) => {
      state.base.status = 'succeeded'
      state.base.expediente = action.payload
      state.base.expediente?.actuaciones.sort((e1, e2) => - e1.fechaFirma + e2.fechaFirma)
    })
    builder.addCase(fetchExpediente.rejected, (state, action) => {
      state.base.status = 'failed'
      state.base.error = action.error.message
    })
    builder.addCase(fetchExpedienteSearch.pending, (state, action) => {
      state.search.status = 'loading'
      state.search.expediente = null
      state.search.error = undefined
    })
    builder.addCase(fetchExpedienteSearch.fulfilled, (state, action) => {
      state.search.status = 'succeeded'
      state.search.expediente = action.payload
      state.search.expediente?.actuaciones.sort((e1, e2) => - e1.fechaFirma + e2.fechaFirma)
    })
    builder.addCase(fetchExpedienteSearch.rejected, (state, action) => {
      state.search.status = 'failed'
      state.search.error = action.error.message
    })
  },
})

export const { clearSearch } = expedienteSlice.actions
export default expedienteSlice.reducer
