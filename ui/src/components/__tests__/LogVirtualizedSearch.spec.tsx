import { render, screen, userEvent } from "../../test-utils"
import LogVirtualizedSearch from "../LogVirtualizedSearch"

describe("LogVirtualizedSearch", () => {
  const defaultProps = {
    inputRef: null,
    cursorIndex: 0,
    totalMatches: 0,
    isSearchProcessing: false,
    searchQuery: "",
  }

  function setup(overrides: Partial<typeof defaultProps> = {}) {
    const handlers = {
      onChange: vi.fn(),
      onFocus: vi.fn(),
      onBlur: vi.fn(),
      onIncrement: vi.fn(),
      onDecrement: vi.fn(),
    }
    const utils = render(
      <LogVirtualizedSearch {...defaultProps} {...handlers} {...overrides} />
    )
    return { ...utils, ...handlers, user: userEvent.setup() }
  }

  it("renders the correct components", () => {
    const { container } = setup()
    expect(container.firstChild).toHaveClass(
      "flotilla-logs-virtualized-search-container"
    )
    expect(screen.getAllByRole("textbox")).toHaveLength(1)
    expect(screen.getByPlaceholderText("Search...")).toBeInTheDocument()
    expect(screen.getAllByRole("button")).toHaveLength(2)
  })

  it("handles input events", async () => {
    const { user, onFocus, onBlur, onChange } = setup()
    const input = screen.getByPlaceholderText("Search...")
    expect(onFocus).toHaveBeenCalledTimes(0)
    expect(onBlur).toHaveBeenCalledTimes(0)

    await user.click(input)
    expect(onFocus).toHaveBeenCalledTimes(1)
    expect(onBlur).toHaveBeenCalledTimes(0)

    await user.keyboard("x")
    expect(onChange).toHaveBeenCalledWith("x")

    await user.tab()
    expect(onFocus).toHaveBeenCalledTimes(1)
    expect(onBlur).toHaveBeenCalledTimes(1)
  })

  it("disables the buttons when there are no matches", () => {
    setup()
    screen.getAllByRole("button").forEach(button => {
      expect(button).toBeDisabled()
    })
  })

  it("handles button click events", async () => {
    const { user, onIncrement, onDecrement } = setup({
      cursorIndex: 5,
      totalMatches: 20,
    })
    expect(screen.getByText("6/20")).toBeInTheDocument()

    const [prev, next] = screen.getAllByRole("button")
    expect(onIncrement).toHaveBeenCalledTimes(0)
    expect(onDecrement).toHaveBeenCalledTimes(0)

    await user.click(prev)
    expect(onIncrement).toHaveBeenCalledTimes(0)
    expect(onDecrement).toHaveBeenCalledTimes(1)

    await user.click(next)
    expect(onIncrement).toHaveBeenCalledTimes(1)
    expect(onDecrement).toHaveBeenCalledTimes(1)
  })
})
