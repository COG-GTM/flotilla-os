import { AxiosError } from "axios"
import { render, screen } from "../../test-utils"
import { Templates, Props } from "../Templates"
import { RequestStatus } from "../Request"
import { ListTemplateResponse, SortOrder, Template } from "../../types"

vi.mock("../../api", () => ({ default: {} }))

const template = (overrides: Partial<Template>): Template => ({
  template_id: "tpl-1",
  template_name: "spark-job",
  version: 1,
  schema: {},
  command_template: "",
  defaults: null,
  avatar_uri: "https://example.com/a.png",
  image: "image",
  memory: 1024,
  gpu: 0,
  cpu: 512,
  env: [],
  privileged: false,
  adaptive_resource_allocation: false,
  container_name: "container",
  tags: [],
  ...overrides,
})

const baseProps = (overrides: Partial<Props> = {}): Props => ({
  data: null,
  requestStatus: RequestStatus.NOT_READY,
  isLoading: false,
  error: null,
  receivedAt: null,
  updateSort: vi.fn(),
  updatePage: vi.fn(),
  updateFilter: vi.fn(),
  currentPage: 1,
  currentSortKey: "template_name",
  currentSortOrder: SortOrder.ASC,
  query: {},
  ...overrides,
})

describe("Templates", () => {
  it("renders a spinner while loading", () => {
    render(<Templates {...baseProps()} />)
    expect(screen.getByRole("progressbar")).toBeInTheDocument()
  })

  it("renders a link per template when data is ready", () => {
    const data: ListTemplateResponse = {
      offset: 0,
      limit: 20,
      total: 2,
      templates: [
        template({
          template_id: "tpl-1",
          template_name: "spark-job",
          version: 1,
        }),
        template({ template_id: "tpl-2", template_name: "etl", version: 3 }),
      ],
    }
    render(
      <Templates {...baseProps({ requestStatus: RequestStatus.READY, data })} />
    )
    expect(screen.getByRole("link", { name: /spark-job v1/ })).toHaveAttribute(
      "href",
      "/templates/tpl-1"
    )
    expect(screen.getByRole("link", { name: /etl v3/ })).toHaveAttribute(
      "href",
      "/templates/tpl-2"
    )
    expect(screen.getAllByAltText("template-logo")).toHaveLength(2)
  })

  it("renders a callout when ready without data", () => {
    render(<Templates {...baseProps({ requestStatus: RequestStatus.READY })} />)
    expect(screen.getByText(/No templates found!/)).toBeInTheDocument()
  })

  it("renders the error when the request fails", () => {
    const error = { code: "500", message: "boom" } as AxiosError
    render(
      <Templates
        {...baseProps({ requestStatus: RequestStatus.ERROR, error })}
      />
    )
    expect(screen.getByText("boom")).toBeInTheDocument()
  })
})
