import ConnectedStopRunButton, { StopRunButton, Props } from "../StopRunButton"
import { RequestStatus } from "../Request"
import api from "../../api"
import { render, screen, userEvent, within } from "../../test-utils"

vi.mock("../../api", () => ({
  default: { stopRun: vi.fn(() => Promise.resolve({})) },
}))
vi.mock("../Toaster", () => ({ default: { show: vi.fn() } }))

const defaultProps: Props = {
  requestStatus: RequestStatus.NOT_READY,
  data: null,
  isLoading: false,
  error: null,
  request: vi.fn(),
  definitionID: "definitionID",
  runID: "runID",
  receivedAt: new Date(),
}

// Opens the confirmation dialog and clicks its "Stop" (submit) button.
async function confirmStop() {
  const user = userEvent.setup()
  await user.click(screen.getByRole("button", { name: "Stop" }))
  const dialog = await screen.findByRole("dialog")
  await user.click(within(dialog).getByRole("button", { name: "Stop" }))
}

describe("StopRunButton", () => {
  it("calls props.request with the correct args when the stop is confirmed", async () => {
    const r = vi.fn()
    render(<StopRunButton {...defaultProps} request={r} />)
    expect(r).toHaveBeenCalledTimes(0)

    await confirmStop()

    expect(r).toHaveBeenCalledTimes(1)
    expect(r).toHaveBeenCalledWith({
      definitionID: defaultProps.definitionID,
      runID: defaultProps.runID,
    })
  })

  it("provides api.stopRun as the requestFn", async () => {
    render(<ConnectedStopRunButton definitionID="id" runID="rid" />)
    expect(api.stopRun).toHaveBeenCalledTimes(0)

    await confirmStop()

    expect(api.stopRun).toHaveBeenCalledTimes(1)
    expect(api.stopRun).toHaveBeenCalledWith({
      definitionID: "id",
      runID: "rid",
    })
  })
})
