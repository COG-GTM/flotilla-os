import * as React from "react"
import { render, screen, userEvent } from "../../test-utils"
import { ListRequest, Props, ChildProps } from "../ListRequest"
import { RequestStatus } from "../Request"
import { SortOrder } from "../../types"

const DEFAULT_PROPS: Props<any, any> = {
  requestStatus: RequestStatus.NOT_READY,
  data: null,
  isLoading: false,
  error: null,
  query: {},
  request: (args: any) => {},
  setQuery: (query: object, shouldReplace?: boolean) => {},
  initialQuery: {},
  getRequestArgs: (query: object) => {},
  children: (props: ChildProps<any, any>) => <span />,
  receivedAt: new Date(),
}

// Renders buttons that call the child-prop update methods, plus the
// current page/sort values, so tests can drive ListRequest like a list view.
const renderControls = (props: ChildProps<any, any>) => (
  <div>
    <button onClick={() => props.updateFilter("foo", "bar")}>filter</button>
    <button onClick={() => props.updatePage(10)}>page</button>
    <button onClick={() => props.updateSort("a")}>sort a</button>
    <span>page: {props.currentPage}</span>
    <span>sort key: {props.currentSortKey}</span>
    <span>sort order: {props.currentSortOrder}</span>
  </div>
)

describe("ListRequest", () => {
  it("calls props.setQuery w/ props.initialQuery if props.query is empty on componentDidMount", () => {
    const request = vi.fn()
    const setQuery = vi.fn()
    const initialQuery = { foo: "bar" }

    render(
      <ListRequest
        {...DEFAULT_PROPS}
        request={request}
        initialQuery={initialQuery}
        query={{}}
        setQuery={setQuery}
      >
        {() => <span />}
      </ListRequest>
    )

    expect(setQuery).toHaveBeenCalledTimes(1)
    expect(setQuery).toHaveBeenCalledWith(initialQuery, true)
    expect(request).toHaveBeenCalledTimes(0)
  })

  it("calls props.request if props.query is not empty on componentDidMount", () => {
    const request = vi.fn()
    const setQuery = vi.fn()

    render(
      <ListRequest
        {...DEFAULT_PROPS}
        request={request}
        query={{ foo: "bar" }}
        setQuery={setQuery}
      >
        {() => <span />}
      </ListRequest>
    )

    expect(setQuery).toHaveBeenCalledTimes(0)
    expect(request).toHaveBeenCalledTimes(1)
  })

  it("calls props.request if prevProps.query and props.query are not equal on componentDidUpdate", () => {
    const request = vi.fn()

    const { rerender } = render(
      <ListRequest {...DEFAULT_PROPS} request={request} query={{ foo: "bar" }}>
        {() => <span />}
      </ListRequest>
    )

    // Should have been called once when the component mounts.
    expect(request).toHaveBeenCalledTimes(1)

    rerender(
      <ListRequest
        {...DEFAULT_PROPS}
        request={request}
        query={{ foo: "not-bar" }}
      >
        {() => <span />}
      </ListRequest>
    )
    expect(request).toHaveBeenCalledTimes(2)

    // An equal (but new) query object does not trigger another request.
    rerender(
      <ListRequest
        {...DEFAULT_PROPS}
        request={request}
        query={{ foo: "not-bar" }}
      >
        {() => <span />}
      </ListRequest>
    )
    expect(request).toHaveBeenCalledTimes(2)
  })

  it("calls props.request with the correct args", () => {
    const request = vi.fn()
    const getRequestArgs = vi.fn((q: object) => ({ args: q }))
    const query = { foo: "bar" }

    const { rerender } = render(
      <ListRequest
        {...DEFAULT_PROPS}
        request={request}
        getRequestArgs={getRequestArgs}
        query={query}
      >
        {() => <span />}
      </ListRequest>
    )

    expect(request).toHaveBeenCalledTimes(1)
    expect(getRequestArgs).toHaveBeenCalledWith(query)
    expect(request).toHaveBeenLastCalledWith({ args: query })

    const nextQuery = { foo: "baz" }
    rerender(
      <ListRequest
        {...DEFAULT_PROPS}
        request={request}
        getRequestArgs={getRequestArgs}
        query={nextQuery}
      >
        {() => <span />}
      </ListRequest>
    )
    expect(request).toHaveBeenCalledTimes(2)
    expect(request).toHaveBeenLastCalledWith({ args: nextQuery })
  })

  it("calls props.children with the correct args", async () => {
    const user = userEvent.setup()
    const setQuery = vi.fn()
    const query = { page: "3", sort_by: "b", order: SortOrder.DESC }

    render(
      <ListRequest {...DEFAULT_PROPS} query={query} setQuery={setQuery}>
        {renderControls}
      </ListRequest>
    )

    expect(screen.getByText("page: 3")).toBeInTheDocument()
    expect(screen.getByText("sort key: b")).toBeInTheDocument()
    expect(screen.getByText(`sort order: ${SortOrder.DESC}`)).toBeInTheDocument()

    // Test sort
    expect(setQuery).toHaveBeenCalledTimes(0)
    await user.click(screen.getByRole("button", { name: "sort a" }))
    expect(setQuery).toHaveBeenCalledTimes(1)
    expect(setQuery).toHaveBeenLastCalledWith({
      ...query,
      page: 1,
      sort_by: "a",
      order: SortOrder.ASC,
    })

    // Test filter
    await user.click(screen.getByRole("button", { name: "filter" }))
    expect(setQuery).toHaveBeenCalledTimes(2)
    expect(setQuery).toHaveBeenLastCalledWith({ ...query, page: 1, foo: "bar" })

    // Test page
    await user.click(screen.getByRole("button", { name: "page" }))
    expect(setQuery).toHaveBeenCalledTimes(3)
    expect(setQuery).toHaveBeenLastCalledWith({ ...query, page: 10 })
  })

  describe("query update methods", () => {
    const setQuery = vi.fn()
    let latestChildProps: ChildProps<any, any>

    const children = (props: ChildProps<any, any>) => {
      latestChildProps = props
      return renderControls(props)
    }

    const renderWithQuery = (query: object) =>
      render(
        <ListRequest {...DEFAULT_PROPS} setQuery={setQuery} query={query}>
          {children}
        </ListRequest>
      )

    afterEach(() => {
      setQuery.mockReset()
    })

    it("updateSort calls setQuery with the correct arguments", () => {
      // Note: we're manually re-rendering with the new query since we're
      // mocking setQuery and it won't actually update the query.
      const { rerender } = renderWithQuery({ a: 1 })
      const rerenderWithQuery = (query: object) =>
        rerender(
          <ListRequest {...DEFAULT_PROPS} setQuery={setQuery} query={query}>
            {children}
          </ListRequest>
        )

      expect(setQuery).toHaveBeenCalledTimes(0)
      latestChildProps.updateSort("x")
      expect(setQuery).toHaveBeenCalledTimes(1)
      expect(setQuery).toHaveBeenCalledWith({
        a: 1,
        page: 1,
        sort_by: "x",
        order: SortOrder.ASC,
      })
      rerenderWithQuery({ sort_by: "x", order: SortOrder.ASC })

      latestChildProps.updateSort("x")
      expect(setQuery).toHaveBeenCalledTimes(2)
      expect(setQuery).toHaveBeenCalledWith({
        page: 1,
        sort_by: "x",
        order: SortOrder.DESC,
      })
      rerenderWithQuery({ sort_by: "x", order: SortOrder.DESC })

      latestChildProps.updateSort("x")
      expect(setQuery).toHaveBeenCalledTimes(3)
      expect(setQuery).toHaveBeenLastCalledWith({
        page: 1,
        sort_by: "x",
        order: SortOrder.ASC,
      })
      rerenderWithQuery({ sort_by: "x", order: SortOrder.ASC })

      latestChildProps.updateSort("y")
      expect(setQuery).toHaveBeenCalledTimes(4)
      expect(setQuery).toHaveBeenLastCalledWith({
        page: 1,
        sort_by: "y",
        order: SortOrder.ASC,
      })
    })

    it("updatePage calls setQuery with the correct arguments", () => {
      renderWithQuery({ a: 1 })
      expect(setQuery).toHaveBeenCalledTimes(0)
      latestChildProps.updatePage(5000)
      expect(setQuery).toHaveBeenCalledTimes(1)
      expect(setQuery).toHaveBeenCalledWith({ a: 1, page: 5000 })
    })

    it("updateFilter calls setQuery with the correct arguments", () => {
      renderWithQuery({ a: 1 })
      expect(setQuery).toHaveBeenCalledTimes(0)
      latestChildProps.updateFilter("foo", "bar")
      expect(setQuery).toHaveBeenCalledTimes(1)
      expect(setQuery).toHaveBeenCalledWith({ a: 1, page: 1, foo: "bar" })
    })
  })
})
