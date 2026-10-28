import * as React from "react"
import { render, screen } from "../../test-utils"
import Request, { ChildProps, RequestStatus } from "../Request"

const renderState = ({
  requestStatus,
  data,
  isLoading,
  error,
}: ChildProps<any, any>) => (
  <ul>
    <li>status: {requestStatus}</li>
    <li>data: {String(data)}</li>
    <li>loading: {String(isLoading)}</li>
    <li>error: {String(error)}</li>
  </ul>
)

describe("Request", () => {
  it("calls props.requestFn with props.initialRequestArgs when the component mounts", () => {
    const requestFn = vi.fn(() => Promise.resolve())
    render(
      <Request requestFn={requestFn} initialRequestArgs={{ foo: "bar" }}>
        {() => null}
      </Request>
    )
    expect(requestFn).toHaveBeenCalledTimes(1)
    expect(requestFn).toHaveBeenCalledWith({ foo: "bar" })
  })

  it("doesn't call props.requestFn when the component mounts if props.shouldRequestOnMount is false", () => {
    const requestFn = vi.fn(() => Promise.resolve())
    render(
      <Request
        requestFn={requestFn}
        initialRequestArgs={{ foo: "bar" }}
        shouldRequestOnMount={false}
      >
        {() => null}
      </Request>
    )
    expect(requestFn).toHaveBeenCalledTimes(0)
  })

  it("sets state correctly during the request method", async () => {
    const data = "data"
    const onSuccess = vi.fn()
    const success = render(
      <Request
        requestFn={() => Promise.resolve(data)}
        initialRequestArgs={{ foo: "bar" }}
        onSuccess={onSuccess}
      >
        {renderState}
      </Request>
    )
    expect(screen.getByText(`status: ${RequestStatus.NOT_READY}`)).toBeInTheDocument()
    expect(screen.getByText("data: null")).toBeInTheDocument()
    expect(screen.getByText("loading: true")).toBeInTheDocument()
    expect(screen.getByText("error: null")).toBeInTheDocument()
    expect(onSuccess).toHaveBeenCalledTimes(0)

    expect(
      await screen.findByText(`status: ${RequestStatus.READY}`)
    ).toBeInTheDocument()
    expect(screen.getByText(`data: ${data}`)).toBeInTheDocument()
    expect(screen.getByText("loading: false")).toBeInTheDocument()
    expect(screen.getByText("error: null")).toBeInTheDocument()
    expect(onSuccess).toHaveBeenCalledTimes(1)
    expect(onSuccess).toHaveBeenCalledWith(data)
    success.unmount()

    const onFailure = vi.fn()
    const err = "err"
    render(
      <Request
        requestFn={() => Promise.reject(err)}
        initialRequestArgs={{ foo: "bar" }}
        onFailure={onFailure}
      >
        {renderState}
      </Request>
    )
    expect(screen.getByText(`status: ${RequestStatus.NOT_READY}`)).toBeInTheDocument()
    expect(screen.getByText("data: null")).toBeInTheDocument()
    expect(screen.getByText("loading: true")).toBeInTheDocument()
    expect(screen.getByText("error: null")).toBeInTheDocument()
    expect(onFailure).toHaveBeenCalledTimes(0)

    expect(
      await screen.findByText(`status: ${RequestStatus.ERROR}`)
    ).toBeInTheDocument()
    expect(screen.getByText("data: null")).toBeInTheDocument()
    expect(screen.getByText("loading: false")).toBeInTheDocument()
    expect(screen.getByText(`error: ${err}`)).toBeInTheDocument()
    expect(onFailure).toHaveBeenCalledTimes(1)
    expect(onFailure).toHaveBeenCalledWith(err)
  })
})
