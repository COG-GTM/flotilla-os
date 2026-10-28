import ConnectedRuns, {
  Runs as UnconnectedRuns,
  Props,
  initialQuery,
} from "../Runs"
import { RequestStatus } from "../Request"
import { SortOrder } from "../../types"
import api from "../../api"
import { createMockRunObject } from "../../helpers/testHelpers"
import { render, screen } from "../../test-utils"

vi.mock("../../api", () => ({
  default: {
    listRun: vi.fn(() =>
      Promise.resolve({ offset: 0, limit: 20, total: 0, history: [] })
    ),
  },
}))

describe("Runs", () => {
  describe("Connected", () => {
    it("requests runs with api.listRun on mount", () => {
      expect(api.listRun).toHaveBeenCalledTimes(0)

      render(<ConnectedRuns />)

      expect(api.listRun).toHaveBeenCalledTimes(1)
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
      currentSortKey: "started_at",
      currentSortOrder: SortOrder.DESC,
      query: initialQuery,
      receivedAt: new Date(),
    }

    it("renders a Spinner props.requestStatus is `NOT_READY`", () => {
      render(
        <UnconnectedRuns
          {...defaultProps}
          requestStatus={RequestStatus.NOT_READY}
        />
      )
      expect(screen.queryByText("No Error Message Provided")).toBeNull()
      expect(screen.queryByRole("table")).toBeNull()
      expect(screen.getByRole("progressbar")).toBeInTheDocument()
    })

    it("renders a Table props.requestStatus is `READY`", () => {
      render(
        <UnconnectedRuns
          {...defaultProps}
          requestStatus={RequestStatus.READY}
          data={{
            offset: 0,
            limit: 20,
            total: 3,
            history: [
              createMockRunObject({ run_id: "a" }),
              createMockRunObject({ run_id: "b" }),
              createMockRunObject({ run_id: "c" }),
            ],
          }}
        />
      )
      expect(screen.queryByText("No Error Message Provided")).toBeNull()
      expect(screen.queryByRole("progressbar")).toBeNull()
      expect(screen.getByRole("table")).toBeInTheDocument()
      for (const name of ["Status", "Started At", "Run ID", "Alias"]) {
        expect(screen.getByRole("columnheader", { name })).toBeInTheDocument()
      }
      for (const runID of ["a", "b", "c"]) {
        expect(screen.getByRole("link", { name: runID })).toHaveAttribute(
          "href",
          `/runs/${runID}`
        )
      }
    })

    it("renders an ErrorCallout props.requestStatus is `ERROR`", () => {
      render(
        <UnconnectedRuns
          {...defaultProps}
          requestStatus={RequestStatus.ERROR}
        />
      )
      expect(screen.getByText("No Error Message Provided")).toBeInTheDocument()
      expect(screen.queryByRole("table")).toBeNull()
      expect(screen.queryByRole("progressbar")).toBeNull()
    })
  })
})
