import { render } from "../../test-utils"
import { LogProcessor } from "../LogProcessor"

vi.mock("../../workers/index")

describe("LogProcessor", () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it("calls processLogs upon mounting and if logs/width changes", () => {
    const processLogs = vi.spyOn(LogProcessor.prototype, "processLogs")
    const { rerender } = render(
      <LogProcessor logs="abc" width={100} height={100} hasRunFinished={false} />
    )
    expect(processLogs).toHaveBeenCalledTimes(1)

    rerender(
      <LogProcessor
        logs="abcdefg"
        width={100}
        height={100}
        hasRunFinished={false}
      />
    )
    expect(processLogs).toHaveBeenCalledTimes(2)

    rerender(
      <LogProcessor
        logs="abcdefg"
        width={200}
        height={100}
        hasRunFinished={false}
      />
    )
    expect(processLogs).toHaveBeenCalledTimes(3)

    // Height changes alone do not re-process logs.
    rerender(
      <LogProcessor
        logs="abcdefg"
        width={200}
        height={300}
        hasRunFinished={false}
      />
    )
    expect(processLogs).toHaveBeenCalledTimes(3)
  })
})
