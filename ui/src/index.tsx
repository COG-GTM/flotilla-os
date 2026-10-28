import "normalize.css/normalize.css"
import "@blueprintjs/icons/lib/css/blueprint-icons.css"
import "@blueprintjs/core/lib/css/blueprint.css"
import React from "react"
import { createRoot } from "react-dom/client"
import { Provider } from "react-redux"
import "./index.css"
import App from "./components/App"
import store from "./state/store"
import { init } from "./state/settings"

store.dispatch(init())

createRoot(document.getElementById("root") as HTMLElement).render(
  <Provider store={store}>
    <App />
  </Provider>
)
