import {
  configureStore,
  Action,
  combineReducers,
  ThunkAction,
  Middleware,
} from "@reduxjs/toolkit"
import { createLogger } from "redux-logger"
import settings from "./settings"
import runView from "./runView"

const rootReducer = combineReducers({
  settings,
  runView,
})

const store = configureStore({
  reducer: rootReducer,
  middleware: getDefaultMiddleware => {
    // Only use redux-logger in non-production.
    if (process.env.NODE_ENV !== "production") {
      return getDefaultMiddleware().concat(
        createLogger({ collapsed: true, timestamp: false }) as Middleware
      )
    }
    return getDefaultMiddleware()
  },
})

export type RootState = ReturnType<typeof rootReducer>
export type AppDispatch = typeof store.dispatch
export type AppThunk = ThunkAction<void, RootState, unknown, Action<string>>
export default store
