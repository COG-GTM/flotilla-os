import React from "react"
import { render, RenderOptions } from "@testing-library/react"
import { Provider } from "react-redux"
import { MemoryRouter, MemoryRouterProps } from "react-router-dom"
import { configureStore, combineReducers } from "@reduxjs/toolkit"
import settings from "./state/settings"
import runView from "./state/runView"

const rootReducer = combineReducers({ settings, runView })

type TestRootState = ReturnType<typeof rootReducer>

export function createTestStore(preloadedState?: Partial<TestRootState>) {
  return configureStore({ reducer: rootReducer, preloadedState })
}

type ProviderOptions = Omit<RenderOptions, "wrapper"> & {
  preloadedState?: Partial<TestRootState>
  store?: ReturnType<typeof createTestStore>
  initialEntries?: MemoryRouterProps["initialEntries"]
}

// Renders `ui` inside a Redux `Provider` (fresh store unless one is passed)
// and a `MemoryRouter` (starting at `initialEntries`, default "/").
export function renderWithProviders(
  ui: React.ReactElement,
  {
    preloadedState,
    store = createTestStore(preloadedState),
    initialEntries = ["/"],
    ...options
  }: ProviderOptions = {}
) {
  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <Provider store={store}>
        <MemoryRouter initialEntries={initialEntries}>{children}</MemoryRouter>
      </Provider>
    )
  }
  return { store, ...render(ui, { wrapper: Wrapper, ...options }) }
}

export * from "@testing-library/react"
export { default as userEvent } from "@testing-library/user-event"
export { renderWithProviders as render }
