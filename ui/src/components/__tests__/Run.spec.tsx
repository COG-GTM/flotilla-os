import { Run, Props } from "../Run"
import { RunStatus, ExecutionEngine, NodeLifecycle } from "../../types"
import { RequestStatus } from "../Request"
import { createMockRunObject } from "../../helpers/testHelpers"
import { render } from "../../test-utils"

// Child views (logs, pod events) fetch on mount; keep those requests pending
// so this spec only exercises the run polling interval.
vi.mock("../../api", () => {
  const pending = () => new Promise<never>(() => {})
  return {
    default: {
      getRunLog: vi.fn(pending),
      getRunLogRaw: vi.fn(pending),
      listRunEvents: vi.fn(pending),
    },
  }
})
vi.mock("../../workers/index")

// jsdom has no ResizeObserver; the log viewer measures itself with one.
vi.stubGlobal(
  "ResizeObserver",
  class {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
)

const MockRun = createMockRunObject({
  instance: {
    dns_name: "dns_name",
    instance_id: "instance_id",
  },
  task_arn: "task_arn",
  run_id: "run_id",
  definition_id: "definition_id",
  alias: "alias",
  image: "image",
  cluster: "cluster",
  exit_code: 0,
  status: RunStatus.RUNNING,
  started_at: "2019-10-24T05:21:51",
  finished_at: "2019-10-25T06:21:51",
  group_name: "group_name",
  env: [],
  engine: ExecutionEngine.EKS,
  node_lifecycle: NodeLifecycle.ON_DEMAND,
  max_cpu_used: 0,
  max_memory_used: 0,
  pod_name: "",
  cpu: 100,
  memory: 100,
  queued_at: "2019-10-24T04:21:51",
})

const defaultProps: Props = {
  requestStatus: RequestStatus.READY,
  data: MockRun,
  isLoading: false,
  error: null,
  runID: MockRun.run_id,
  request: vi.fn(),
  query: {},
  setQuery: vi.fn(),
  receivedAt: new Date(),
  shouldAutoscroll: true,
  hasLogs: false,
  isLogRequestIntervalActive: false,
  dispatch: vi.fn(),
}

describe("Run", () => {
  let setRequestInterval: ReturnType<typeof vi.spyOn>
  let clearRequestInterval: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    setRequestInterval = vi
      .spyOn(Run.prototype, "setRequestInterval")
      .mockImplementation(() => {})
    clearRequestInterval = vi
      .spyOn(Run.prototype, "clearRequestInterval")
      .mockImplementation(() => {})
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  /**
   * If the run is in a non-stopped state, the component should start an
   * interval to continuously fetch the run.
   */
  it("sets a request interval if the run isn't stopped on componentDidMount", () => {
    expect(setRequestInterval).toHaveBeenCalledTimes(0)

    // Mount a stopped run.
    render(
      <Run
        {...defaultProps}
        data={{
          ...MockRun,
          status: RunStatus.STOPPED,
        }}
      />
    )
    expect(setRequestInterval).toHaveBeenCalledTimes(0)

    // Mount a running one.
    render(<Run {...defaultProps} />)
    expect(setRequestInterval).toHaveBeenCalledTimes(1)
  })

  it("sets the request interval if props.requestStatus changes from NOT_READY to READY and the run is not stopped.", () => {
    const notReadyProps: Props = {
      ...defaultProps,
      requestStatus: RequestStatus.NOT_READY,
      data: null,
      runID: "a",
    }

    // Request has not completed.
    const stopped = render(<Run {...notReadyProps} />)
    expect(setRequestInterval).toHaveBeenCalledTimes(0)

    // Set requestStatus to READY.
    stopped.rerender(
      <Run
        {...notReadyProps}
        requestStatus={RequestStatus.READY}
        data={{ ...MockRun, status: RunStatus.STOPPED }}
      />
    )
    expect(setRequestInterval).toHaveBeenCalledTimes(0)

    // Request has not completed.
    const running = render(<Run {...notReadyProps} />)
    expect(setRequestInterval).toHaveBeenCalledTimes(0)

    // Set requestStatus to READY.
    running.rerender(
      <Run
        {...notReadyProps}
        requestStatus={RequestStatus.READY}
        data={{ ...MockRun, status: RunStatus.RUNNING }}
      />
    )
    expect(setRequestInterval).toHaveBeenCalledTimes(1)
  })

  it("clears the request interval if the run transitions into a stopped state on componentDidUpdate", () => {
    const props: Props = { ...defaultProps, runID: "a" }
    const { rerender } = render(<Run {...props} />)
    expect(clearRequestInterval).toHaveBeenCalledTimes(0)
    expect(setRequestInterval).toHaveBeenCalledTimes(1)

    // Set the state to stopped
    rerender(
      <Run {...props} data={{ ...MockRun, status: RunStatus.STOPPED }} />
    )

    expect(clearRequestInterval).toHaveBeenCalledTimes(1)
  })
})
