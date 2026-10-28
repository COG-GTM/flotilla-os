import { createMemoryHistory } from "history"
import { Route, Router } from "react-router-dom"
import { render, screen, userEvent, waitFor } from "../../test-utils"
import TemplateExecutionForm from "../TemplateExecutionForm"
import { TemplateContext, TemplateCtx } from "../Template"
import { RequestStatus } from "../Request"
import { ExecutionEngine, Run, Template } from "../../types"
import api from "../../api"

vi.mock("../../api", () => ({
  default: {
    runTemplate: vi.fn(),
    listClusters: vi.fn(() => Promise.resolve({ clusters: ["c1"] })),
  },
}))
vi.mock("../Toaster", () => ({ default: { show: vi.fn() } }))

const template: Template = {
  template_id: "tpl-1",
  template_name: "spark-job",
  version: 1,
  schema: {
    type: "object",
    properties: { name: { type: "string", default: "default-name" } },
  },
  command_template: "",
  defaults: null,
  avatar_uri: null,
  image: "image",
  memory: 2048,
  gpu: 0,
  cpu: 1024,
  env: [{ name: "FOO", value: "bar" }],
  privileged: false,
  adaptive_resource_allocation: false,
  container_name: "container",
  tags: [],
}

const ctx: TemplateCtx = {
  data: template,
  requestStatus: RequestStatus.READY,
  isLoading: false,
  error: null,
  receivedAt: new Date(),
  request: vi.fn(),
  basePath: "/templates/:templateID",
  templateID: "tpl-1",
}

const renderForm = (state?: Partial<Run>, value: TemplateCtx = ctx) => {
  const history = createMemoryHistory()
  history.push("/templates/tpl-1/execute", state)
  const utils = render(
    <Router history={history}>
      <TemplateContext.Provider value={value}>
        <Route component={TemplateExecutionForm} />
      </TemplateContext.Provider>
    </Router>
  )
  return { history, ...utils }
}

const input = (container: HTMLElement, name: string) =>
  container.querySelector(`input[name="${name}"]`) as HTMLInputElement

describe("TemplateExecutionForm", () => {
  const prevCluster = process.env.REACT_APP_EKS_CLUSTER_NAME

  beforeEach(() => {
    process.env.REACT_APP_EKS_CLUSTER_NAME = "eks-cluster"
    vi.mocked(api.runTemplate).mockReset()
  })

  afterAll(() => {
    process.env.REACT_APP_EKS_CLUSTER_NAME = prevCluster
  })

  it("renders the error callout when the template fails to load", () => {
    renderForm(undefined, {
      ...ctx,
      data: null,
      requestStatus: RequestStatus.ERROR,
      error: { code: "404", message: "not found" } as TemplateCtx["error"],
    })
    expect(screen.getByText("not found")).toBeInTheDocument()
  })

  it("prefills values from the template when there is no previous run", () => {
    const { container } = renderForm()
    expect(input(container, "cpu")).toHaveValue(1024)
    expect(input(container, "memory")).toHaveValue(2048)
    expect(screen.getByRole("radio", { name: "EKS" })).toBeChecked()
    expect(screen.getByDisplayValue("FOO")).toBeInTheDocument()
    expect(screen.getByDisplayValue("bar")).toBeInTheDocument()
  })

  it("prefills values from the run passed in location state", () => {
    const { container } = renderForm({
      cpu: 2048,
      memory: 4096,
      engine: ExecutionEngine.ECS,
      cluster: "ecs-cluster",
      env: [{ name: "RUN_ENV", value: "x" }],
    })
    expect(input(container, "cpu")).toHaveValue(2048)
    expect(input(container, "memory")).toHaveValue(4096)
    expect(screen.getByRole("radio", { name: "ECS" })).toBeChecked()
    expect(screen.getByText("Cluster")).toBeInTheDocument()
    expect(screen.getByDisplayValue("RUN_ENV")).toBeInTheDocument()
  })

  it("submits the form values with the default template payload", async () => {
    vi.mocked(api.runTemplate).mockResolvedValue({ run_id: "run-9" } as never)
    const user = userEvent.setup()
    const { history } = renderForm()

    const submit = screen.getByRole("button", { name: "Submit" })
    await waitFor(() => expect(submit).toBeEnabled())
    await user.click(submit)

    await waitFor(() => expect(api.runTemplate).toHaveBeenCalledTimes(1))
    expect(api.runTemplate).toHaveBeenCalledWith({
      templateID: "tpl-1",
      data: expect.objectContaining({
        cpu: 1024,
        memory: 2048,
        engine: ExecutionEngine.EKS,
        cluster: "eks-cluster",
        env: [{ name: "FOO", value: "bar" }],
        template_payload: { name: "default-name" },
      }),
    })
    await waitFor(() => expect(history.location.pathname).toBe("/runs/run-9"))
  })
})
