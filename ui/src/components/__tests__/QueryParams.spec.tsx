import * as React from "react"
import qs from "qs"
import { render } from "../../test-utils"
import ConnectedQueryParams from "../QueryParams"

describe("QueryParams", () => {
  it("provides a `query` and `setQuery` prop to it's children", () => {
    const children = vi.fn(() => <span />)
    const q = "?foo=bar&bar=baz&env=a|b&env=c|d"
    render(<ConnectedQueryParams>{children}</ConnectedQueryParams>, {
      initialEntries: [{ pathname: "foo", search: q }],
    })
    expect(children).toHaveBeenCalledWith({
      query: qs.parse(q.substr(1)),
      setQuery: expect.any(Function),
    })
  })
})
