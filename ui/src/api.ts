import FlotillaClient from "./helpers/FlotillaClient"

const err =
  "Base URL undefined. If you are running this in development, please set the `REACT_APP_BASE_URL_DEV` environment variable. If you are running this in production, please set the `REACT_APP_BASE_URL` environment variable."

let baseURL: string | undefined = undefined

switch (process.env.NODE_ENV) {
  case "production":
    baseURL = process.env.REACT_APP_BASE_URL
    break
  case "development":
  case "test":
  default:
    baseURL = process.env.REACT_APP_BASE_URL_DEV
    break
}

if (baseURL === undefined) {
  throw new Error(err)
}

const token = process.env.REACT_APP_API_TOKEN

const client = new FlotillaClient({
  baseURL,
  headers: token ? { Authorization: `Bearer ${token}` } : {},
})

export default client
