import { createMemoryHistory } from "history"
import { Route, Router } from "react-router-dom"
import { Classes } from "@blueprintjs/core"
import { render, screen, userEvent, waitFor, within } from "../../test-utils"
import TemplateRunForm from "../TemplateRunForm"
import { TemplateContext, TemplateCtx } from "../Template"
import { RequestStatus } from "../Request"
import { ExecutionEngine, Template } from "../../types"
import api from "../../api"

vi.mock("../../api", () => ({
  default: {
    runTemplate: vi.fn(),
    listClusters: vi.fn(() => Promise.resolve({ clusters: [] })),
  },
}))
vi.mock("../Toaster", () => ({ default: { show: vi.fn() } }))

const template: Template = {
  template_id: "tpl-1",
  template_name: "spark-job",
  version: 1,
  schema: {
    type: "object",
    properties: { name: { type: "string", title: "Job Name" } },
  },
  command_template: "",
  defaults: null,
  avatar_uri: null,
  image: "image",
  memory: 1024,
  gpu: 0,
  cpu: 512,
  env: [],
  privileged: false,
  adaptive_resource_allocation: false,
  container_name: "container",
  tags: [],
}

const ctx = (overrides: Partial<TemplateCtx> = {}): TemplateCtx => ({
  data: template,
  requestStatus: RequestStatus.READY,
  isLoading: false,
  error: null,
  receivedAt: new Date(),
  request: vi.fn(),
  basePath: "/templates/:templateID",
  templateID: "tpl-1",
  ...overrides,
})

const renderForm = (value: TemplateCtx) => {
  const history = createMemoryHistory({
    initialEntries: ["/templates/tpl-1/run"],
  })
  const utils = render(
    <Router history={history}>
      <TemplateContext.Provider value={value}>
        <Route component={TemplateRunForm} />
      </TemplateContext.Provider>
    </Router>
  )
  return { history, ...utils }
}

describe("TemplateRunForm", () => {
  beforeEach(() => {
    vi.mocked(api.runTemplate).mockReset()
  })

  it("renders a spinner until the template is loaded", () => {
    renderForm(ctx({ data: null, requestStatus: RequestStatus.NOT_READY }))
    expect(screen.getByRole("progressbar")).toBeInTheDocument()
  })

  it("renders the Formik fields and the JSON schema fields", () => {
    const { container } = renderForm(ctx())
    expect(
      container.querySelector('input[name="owner_id"]')
    ).toBeInTheDocument()
    expect(screen.getByRole("radio", { name: "EKS" })).toBeChecked()
    expect(screen.getByText("Job Name")).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Submit" })).toBeInTheDocument()
  })

  it("submits the Formik values merged with the JSON schema form data", async () => {
    vi.mocked(api.runTemplate).mockResolvedValue({ run_id: "run-1" } as never)
    const user = userEvent.setup()
    const { container, history } = renderForm(ctx())

    await user.type(
      container.querySelector('input[name="owner_id"]') as HTMLInputElement,
      "alice"
    )
    const group = screen
      .getByText("Job Name")
      .closest(`.${Classes.FORM_GROUP}`) as HTMLElement
    await user.type(within(group).getByRole("textbox"), "nightly")
    await user.click(screen.getByRole("button", { name: "Submit" }))

    await waitFor(() => expect(api.runTemplate).toHaveBeenCalledTimes(1))
    expect(api.runTemplate).toHaveBeenCalledWith({
      templateID: "tpl-1",
      data: expect.objectContaining({
        owner_id: "alice",
        cpu: 512,
        memory: 512,
        engine: ExecutionEngine.EKS,
        template_payload: { name: "nightly" },
      }),
    })
    await waitFor(() => expect(history.location.pathname).toBe("/runs/run-1"))
  })
})
