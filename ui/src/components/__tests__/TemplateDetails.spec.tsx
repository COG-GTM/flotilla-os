import { render, screen } from "../../test-utils"
import { Template } from "../Template"
import { TemplateCtx } from "../Template"
import { RequestStatus } from "../Request"
import { Template as TemplateShape } from "../../types"
import { createMockRunObject } from "../../helpers/testHelpers"
import api from "../../api"

vi.mock("../../api", () => ({
  default: {
    listTemplateHistoryByTemplateID: vi.fn(),
  },
}))

const template: TemplateShape = {
  template_id: "tpl-1",
  template_name: "spark-job",
  version: 2,
  schema: {},
  command_template: "",
  defaults: null,
  avatar_uri: "https://example.com/a.png",
  image: "my-image:latest",
  memory: 2048,
  gpu: 1,
  cpu: 1024,
  env: [{ name: "FOO", value: "bar" }],
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

describe("Template / TemplateDetails", () => {
  beforeEach(() => {
    vi.mocked(api.listTemplateHistoryByTemplateID).mockResolvedValue({
      offset: 0,
      limit: 20,
      total: 1,
      history: [createMockRunObject({ run_id: "run-abc" })],
    })
  })

  it("renders a spinner while the template loads", () => {
    render(
      <Template
        {...ctx({ data: null, requestStatus: RequestStatus.NOT_READY })}
      />,
      { initialEntries: ["/templates/tpl-1"] }
    )
    expect(screen.getAllByRole("progressbar").length).toBeGreaterThan(0)
  })

  it("renders the template's attributes, env and Run link", async () => {
    render(<Template {...ctx()} />, { initialEntries: ["/templates/tpl-1"] })

    expect(screen.getByText("my-image:latest")).toBeInTheDocument()
    expect(screen.getByText("Memory (MB)")).toBeInTheDocument()
    expect(screen.getByText("FOO")).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Run" })).toHaveAttribute(
      "href",
      "/templates/tpl-1/execute"
    )
  })

  it("loads and renders the template's run history", async () => {
    render(<Template {...ctx()} />, { initialEntries: ["/templates/tpl-1"] })

    expect(
      await screen.findByRole("link", { name: "run-abc" })
    ).toHaveAttribute("href", "/runs/run-abc")
    expect(api.listTemplateHistoryByTemplateID).toHaveBeenCalledWith(
      expect.objectContaining({ templateID: "tpl-1" })
    )
  })

  it("renders the error callout when the template fails to load", () => {
    render(
      <Template
        {...ctx({
          data: null,
          requestStatus: RequestStatus.ERROR,
          error: { code: "404", message: "missing" } as TemplateCtx["error"],
        })}
      />,
      { initialEntries: ["/templates/tpl-1"] }
    )
    expect(screen.getByText("missing")).toBeInTheDocument()
  })
})
