import React from "react"
import { render, screen, userEvent, waitFor } from "../../test-utils"
import Connected, { ClusterSelect } from "../ClusterSelect"
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

describe("ClusterSelect", () => {
  describe("Unconnected", () => {
    it("renders a Creatable component", async () => {
      const props = {
        options: [
          { label: "a", value: "a" },
          { label: "b", value: "b" },
          { label: "c", value: "c" },
        ],
        value: "a",
        onChange: vi.fn(),
      }
      render(<ClusterSelect {...props} isDisabled={false} />)

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
  })

  describe("Connected", () => {
    beforeEach(() => {
      vi.clearAllMocks()
    })

    it("calls api.listClusters", () => {
      expect(api.listClusters).toHaveBeenCalledTimes(0)
      render(<Connected value="" onChange={vi.fn()} isDisabled={false} />)
      expect(api.listClusters).toHaveBeenCalledTimes(1)
    })

    it("sends an empty array to the select if the server returns null", async () => {
      vi.mocked(api.listClusters).mockImplementationOnce(() =>
        Promise.resolve({
          offset: 0,
          limit: 10,
          clusters: (null as unknown) as string[], // the server can send null
          total: 0,
        })
      )
      render(<Connected value="" onChange={vi.fn()} isDisabled={false} />)
      await waitFor(() => {
        expect(api.listClusters).toHaveBeenCalledTimes(1)
      })

      await userEvent.click(screen.getByRole("combobox"))
      expect(screen.queryAllByRole("option")).toHaveLength(0)
      expect(screen.getByText("No options")).toBeInTheDocument()
    })
  })
})
