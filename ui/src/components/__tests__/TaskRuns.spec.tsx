import ConnectedTaskRuns, {
  TaskRuns as UnconnectedTaskRuns,
  Props,
  initialQuery,
} from "../TaskRuns"
import { RequestStatus } from "../Request"
import { SortOrder } from "../../types"
import api from "../../api"
import { createMockRunObject } from "../../helpers/testHelpers"
import { render, screen } from "../../test-utils"

vi.mock("../../api", () => ({
  default: {
    listTaskRuns: vi.fn(() =>
      Promise.resolve({ offset: 0, limit: 20, total: 0, history: [] })
    ),
  },
}))

describe("TaskRuns", () => {
  describe("Connected", () => {
    it("requests the task's runs with api.listTaskRuns on mount", () => {
      const definitionID = "foo"
      expect(api.listTaskRuns).toHaveBeenCalledTimes(0)

      render(<ConnectedTaskRuns definitionID={definitionID} />)

      expect(api.listTaskRuns).toHaveBeenCalledTimes(1)
      expect(api.listTaskRuns).toHaveBeenCalledWith(
        expect.objectContaining({
          definitionID,
        })
      )
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

    it("renders a Spinner props.requestStatus is `NOT_READY`", () => {
      render(
        <UnconnectedTaskRuns
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
        <UnconnectedTaskRuns
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
      for (const name of [
        "Run ID",
        "Status",
        "Started At",
        "Finished At",
        "Cluster",
      ]) {
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
        <UnconnectedTaskRuns
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
