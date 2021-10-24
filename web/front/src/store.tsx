import { configureStore  } from '@reduxjs/toolkit'
import { combineReducers } from "redux"; 


import expediente from './expedienteSlice'

const reducer = combineReducers({
  expediente,
});

export const store = configureStore({
  reducer,
});

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
