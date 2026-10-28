import { Colors } from "@blueprintjs/core"
import { render, screen, userEvent } from "../../test-utils"
import { LogVirtualized, Props } from "../LogVirtualized"

const defaultProps: Props = {
  width: 100,
  height: 100,
  logs: ["a", "b", "c", "d"],
  shouldAutoscroll: true,
  dispatch: vi.fn(),
  hasRunFinished: false,
  hasLogs: true,
  isLogRequestIntervalActive: true,
  settings: {
    USE_OPTIMIZED_LOG_RENDERER: true,
    SHOULD_OVERRIDE_CMD_F_IN_RUN_VIEW: true,
  },
}

describe("LogVirtualized", () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it("scrolls to the most recent line upon mounting", () => {
    const scrollTo = vi.spyOn(LogVirtualized.prototype, "scrollTo")
    expect(scrollTo).toHaveBeenCalledTimes(0)

    // Mount LogVirtualized with shouldAutoscroll === true.
    const { unmount } = render(<LogVirtualized {...defaultProps} />)
    expect(scrollTo).toHaveBeenCalledTimes(1)
    expect(scrollTo).toHaveBeenLastCalledWith(defaultProps.logs.length, "end")
    unmount()

    // Mount LogVirtualized with shouldAutoscroll === false.
    render(<LogVirtualized {...defaultProps} shouldAutoscroll={false} />)
    expect(scrollTo).toHaveBeenCalledTimes(1)
  })

  it("calls this.handleCursorChange if state.searchCursor is updated", async () => {
    const user = userEvent.setup()
    const handleCursorChange = vi.spyOn(
      LogVirtualized.prototype,
      "handleCursorChange"
    )
    render(<LogVirtualized {...defaultProps} logs={["ab", "ac", "ad"]} />)
    expect(handleCursorChange).toHaveBeenCalledTimes(0)

    // Searching sets the query and resets the cursor.
    await user.type(screen.getByPlaceholderText("Search..."), "a")
    expect(handleCursorChange).toHaveBeenCalledTimes(1)

    // Moving to the next match updates the cursor.
    const [, next] = screen.getAllByRole("button")
    await user.click(next)
    expect(handleCursorChange).toHaveBeenCalledTimes(2)
  })

  it("scrolls to the most recent line if the number of lines is different", () => {
    const scrollTo = vi.spyOn(LogVirtualized.prototype, "scrollTo")
    const { rerender } = render(<LogVirtualized {...defaultProps} />)
    expect(scrollTo).toHaveBeenCalledTimes(1)

    const logs = ["a", "b", "c", "d", "e", "f"]
    rerender(<LogVirtualized {...defaultProps} logs={logs} />)
    expect(scrollTo).toHaveBeenCalledTimes(2)
    expect(scrollTo).toHaveBeenLastCalledWith(logs.length, "end")
  })

  it("handles search correctly", async () => {
    const user = userEvent.setup()
    const logs = ["one two three", "four five six", "seven eight nine"]
    render(<LogVirtualized {...defaultProps} logs={logs} />)
    const input = screen.getByPlaceholderText("Search...")

    // No matches before searching.
    expect(screen.queryByText(/^\d+\/\d+$/)).not.toBeInTheDocument()

    // "s" matches lines 1 and 2; the cursor starts on the first match.
    await user.type(input, "s")
    expect(screen.getByText("1/2")).toBeInTheDocument()
    expect(screen.getByText("four five six").closest("pre")).toHaveStyle({
      color: Colors.GOLD5,
    })

    // "seven" only matches line 2.
    await user.clear(input)
    await user.type(input, "seven")
    expect(screen.getByText("1/1")).toBeInTheDocument()
    expect(screen.getByText("seven eight nine").closest("pre")).toHaveStyle({
      color: Colors.GOLD5,
    })
    expect(
      screen.getByText("four five six").closest("pre")
    ).not.toHaveStyle({ color: Colors.GOLD5 })
  })

  it("handles cursor changes correctly", async () => {
    const user = userEvent.setup()
    const scrollTo = vi.spyOn(LogVirtualized.prototype, "scrollTo")
    const logs = ["x0", "x1", "x2", "x3"]
    render(<LogVirtualized {...defaultProps} logs={logs} />)

    // Every line matches "x"; the cursor starts on line 0.
    await user.type(screen.getByPlaceholderText("Search..."), "x")
    expect(screen.getByText("1/4")).toBeInTheDocument()
    expect(scrollTo).toHaveBeenLastCalledWith(0, "center")

    const [prev, next] = screen.getAllByRole("button")

    await user.click(next)
    expect(screen.getByText("2/4")).toBeInTheDocument()
    expect(scrollTo).toHaveBeenLastCalledWith(1, "center")

    await user.click(next)
    expect(screen.getByText("3/4")).toBeInTheDocument()
    expect(scrollTo).toHaveBeenLastCalledWith(2, "center")

    // Decrementing from the first match wraps to the last one.
    await user.click(prev)
    await user.click(prev)
    await user.click(prev)
    expect(screen.getByText("4/4")).toBeInTheDocument()
    expect(scrollTo).toHaveBeenLastCalledWith(3, "center")
  })
})
