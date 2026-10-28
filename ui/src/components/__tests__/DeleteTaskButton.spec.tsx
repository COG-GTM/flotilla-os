import * as React from "react"
import { render, screen, userEvent, waitFor, within } from "../../test-utils"
import ConnectedDeleteTaskButton, {
  DeleteTaskButton,
  Props,
} from "../DeleteTaskButton"
import { RequestStatus } from "../Request"
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

const defaultProps: Props = {
  requestStatus: RequestStatus.NOT_READY,
  data: null,
  isLoading: false,
  error: null,
  request: vi.fn(),
  definitionID: "definitionID",
  receivedAt: new Date(),
}

const confirmDelete = async () => {
  await userEvent.click(screen.getByRole("button", { name: "Delete" }))
  const dialog = await screen.findByRole("dialog")
  await userEvent.click(within(dialog).getByRole("button", { name: "Delete" }))
}

describe("DeleteTaskButton", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it("calls props.request with the correct args when the delete is confirmed", async () => {
    const r = vi.fn()
    render(<DeleteTaskButton {...defaultProps} request={r} />)
    expect(r).toHaveBeenCalledTimes(0)
    await confirmDelete()
    expect(r).toHaveBeenCalledTimes(1)
    expect(r).toHaveBeenCalledWith({ definitionID: defaultProps.definitionID })
  })

  it("provides api.deleteTask as the requestFn", async () => {
    render(<ConnectedDeleteTaskButton definitionID="id" />)
    expect(api.deleteTask).toHaveBeenCalledTimes(0)
    await confirmDelete()
    await waitFor(() => {
      expect(api.deleteTask).toHaveBeenCalledTimes(1)
    })
    expect(api.deleteTask).toHaveBeenCalledWith({ definitionID: "id" })
  })
})
