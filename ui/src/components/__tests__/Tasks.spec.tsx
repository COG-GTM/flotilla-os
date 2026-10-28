import React from "react"
import { render, screen } from "../../test-utils"
import ConnectedTasks, {
  Tasks as UnconnectedTasks,
  Props,
  initialQuery,
} from "../Tasks"
import { RequestStatus } from "../Request"
import { SortOrder } from "../../types"
import api from "../../api"
import { createMockTaskObject } from "../../helpers/testHelpers"

vi.hoisted(() => {
  process.env.REACT_APP_BASE_URL_DEV =
    process.env.REACT_APP_BASE_URL_DEV || "http://localhost"
})
vi.mock("../../helpers/FlotillaClient", async () => {
  const { default: createMockClient } = await import(
    "../../helpers/__mocks__/FlotillaClient"
  )
  // The shared mock's arrow-function implementation can't be called with `new`
  // under Vitest, so wrap it in a constructible function.
  return {
    default: function FlotillaClient() {
      return createMockClient()
    },
  }
})

describe("Tasks", () => {
  describe("Connected", () => {
    it("renders ListRequest and provides api.listTasks as the requestFn", () => {
      expect(api.listTasks).toHaveBeenCalledTimes(0)
      render(<ConnectedTasks />)
      expect(api.listTasks).toHaveBeenCalledTimes(1)
    })
  })

  describe("Unconnected", () => {
    const defaultProps: Props = {
      requestStatus: RequestStatus.NOT_READY,
      data: null,
      isLoading: false,
      error: null,
      updateSort: () => {},
      updatePage: () => {},
      updateFilter: () => {},
      currentPage: 1,
      currentSortKey: "alias",
      currentSortOrder: SortOrder.ASC,
      query: initialQuery,
      receivedAt: new Date(),
    }

    // ErrorCallout renders this text when no error code is provided.
    const errorCalloutText = "No Error Code Provided"

    it("renders a Spinner props.requestStatus is `NOT_READY`", () => {
      render(
        <UnconnectedTasks
          {...defaultProps}
          requestStatus={RequestStatus.NOT_READY}
        />
      )
      expect(screen.queryByText(errorCalloutText)).not.toBeInTheDocument()
      expect(screen.queryByRole("table")).not.toBeInTheDocument()
      expect(screen.getByRole("progressbar")).toBeInTheDocument()
    })

    it("renders a Table props.requestStatus is `READY`", () => {
      render(
        <UnconnectedTasks
          {...defaultProps}
          requestStatus={RequestStatus.READY}
          data={{
            offset: 0,
            limit: 20,
            total: 20,
            definitions: [
              createMockTaskObject({ definition_id: "a" }),
              createMockTaskObject({ definition_id: "b" }),
              createMockTaskObject({ definition_id: "c" }),
            ],
          }}
        />
      )
      expect(screen.queryByText(errorCalloutText)).not.toBeInTheDocument()
      expect(screen.queryByRole("progressbar")).not.toBeInTheDocument()
      expect(screen.getByRole("table")).toBeInTheDocument()
      ;["Alias", "Group Name", "Image", "Memory (MB)"].forEach(name => {
        expect(screen.getByRole("columnheader", { name })).toBeInTheDocument()
      })
    })

    it("renders an ErrorCallout props.requestStatus is `ERROR`", () => {
      render(
        <UnconnectedTasks
          {...defaultProps}
          requestStatus={RequestStatus.ERROR}
        />
      )
      expect(screen.getByText(errorCalloutText)).toBeInTheDocument()
      expect(screen.queryByRole("table")).not.toBeInTheDocument()
      expect(screen.queryByRole("progressbar")).not.toBeInTheDocument()
    })
  })
})
