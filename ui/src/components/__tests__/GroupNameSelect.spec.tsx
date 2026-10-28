import React from "react"
import { render, screen, userEvent } from "../../test-utils"
import Connected, { GroupNameSelect } from "../GroupNameSelect"
import api from "../../api"

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

describe("GroupNameSelect", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it("renders a Select component", async () => {
    const props = {
      options: [
        { label: "a", value: "a" },
        { label: "b", value: "b" },
        { label: "c", value: "c" },
      ],
      value: "a",
      onChange: vi.fn(),
    }
    render(<GroupNameSelect {...props} isDisabled={false} />)

    // The select renders props.value as the selected option.
    const combobox = screen.getByRole("combobox")
    expect(screen.getByText("a")).toBeInTheDocument()

    // Opening the menu lists props.options.
    await userEvent.click(combobox)
    expect(screen.getAllByRole("option").map(o => o.textContent)).toEqual([
      "a",
      "b",
      "c",
    ])

    // Selecting an option calls props.onChange with its value.
    expect(props.onChange).toHaveBeenCalledTimes(0)
    await userEvent.click(screen.getByRole("option", { name: "b" }))
    expect(props.onChange).toHaveBeenCalledTimes(1)
    expect(props.onChange).toHaveBeenCalledWith("b")
  })

  it("calls api.listGroups", async () => {
    expect(api.listGroups).toHaveBeenCalledTimes(0)
    render(<Connected value="" onChange={vi.fn()} isDisabled={false} />)
    expect(api.listGroups).toHaveBeenCalledTimes(1)
    // Once the groups load, the select replaces the spinner.
    expect(await screen.findByRole("combobox")).toBeInTheDocument()
  })
})
