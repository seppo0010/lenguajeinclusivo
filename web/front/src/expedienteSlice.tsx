import { createSlice, createAsyncThunk, nanoid } from '@reduxjs/toolkit'
import type { RootState } from './store'

export const fetchExpediente = createAsyncThunk('expediente/get', async ({id}: {id: string}) => {
  const req = await fetch('/api/expediente/' + encodeURIComponent(id.replace('/', '-')))
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
  firmantes: string
}

export interface State {
  expediente: Expediente | null
  status: 'idle' | 'succeeded' | 'loading' | 'failed'
  error: undefined | string,
}

const initialState: State = {
  expediente: null,
  status: 'idle',
  error: undefined,
}

const expedienteSlice = createSlice({
  name: 'expediente',
  initialState,
  reducers: {
    expedienteFetched: {
      reducer(state, action) {
        state.expediente = action.payload
      },
      prepare(data: any) { return { id: nanoid(), payload: data } as any },
    },
  },
  extraReducers: (builder) => {
    builder.addCase(fetchExpediente.pending, (state, action) => {
      state.status = 'loading'
    })
    builder.addCase(fetchExpediente.fulfilled, (state, action) => {
      state.status = 'succeeded'
      state.expediente = action.payload
    })
    builder.addCase(fetchExpediente.rejected, (state, action) => {
      state.status = 'failed'
      state.error = action.error.message
    })
  },
})

export const selectExpediente = (state: RootState) => state.expediente.expediente

export default expedienteSlice.reducer
