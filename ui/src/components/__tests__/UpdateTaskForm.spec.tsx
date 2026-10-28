import * as React from "react"
import { Route } from "react-router-dom"
import { render, screen, userEvent, waitFor } from "../../test-utils"
import UpdateTaskForm from "../UpdateTaskForm"
import api from "../../api"
import { createMockTaskObject } from "../../helpers/testHelpers"
import { RequestStatus } from "../Request"
import { TaskContext, TaskCtx } from "../Task"

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

describe("UpdateTaskForm", () => {
  const DEFINITION_ID = "my_def_id"
  const task = createMockTaskObject({ definition_id: DEFINITION_ID })
  const mockTaskCtx: TaskCtx = {
    data: task,
    requestStatus: RequestStatus.READY,
    isLoading: false,
    error: null,
    request: vi.fn(),
    basePath: "",
    definitionID: DEFINITION_ID,
    receivedAt: new Date(),
  }
  let pathname = ""

  beforeEach(() => {
    vi.clearAllMocks()
    render(
      <TaskContext.Provider value={mockTaskCtx}>
        <Route
          path="/"
          render={routeProps => {
            pathname = routeProps.location.pathname
            return (
              <UpdateTaskForm {...routeProps} definitionID={DEFINITION_ID} />
            )
          }}
        />
      </TaskContext.Provider>,
      { initialEntries: [`/tasks/${DEFINITION_ID}/update`] }
    )
  })

  it("renders the correct components", () => {
    // The request is only made on submit, not on mount.
    expect(api.updateTask).toHaveBeenCalledTimes(0)
    // BaseTaskForm fields are prefilled from the task context.
    expect(screen.getByDisplayValue(task.image)).toHaveAttribute(
      "name",
      "image"
    )
    expect(screen.getByDisplayValue(task.command)).toHaveAttribute(
      "name",
      "command"
    )
    // Alias cannot be updated.
    expect(screen.queryByDisplayValue(task.alias)).not.toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Submit" })).toHaveAttribute(
      "id",
      "submitButton"
    )
  })

  it("calls api.updateTask when submitted", async () => {
    expect(api.updateTask).toHaveBeenCalledTimes(0)
    expect(mockTaskCtx.request).toHaveBeenCalledTimes(0)

    await userEvent.click(screen.getByRole("button", { name: "Submit" }))

    await waitFor(() => {
      expect(api.updateTask).toHaveBeenCalledTimes(1)
    })
    expect(api.updateTask).toHaveBeenCalledWith({
      definitionID: DEFINITION_ID,
      data: {
        env: task.env,
        image: task.image,
        group_name: task.group_name,
        memory: task.memory,
        cpu: task.cpu,
        command: task.command,
        tags: task.tags,
      },
    })

    // On success, the task is re-requested and the user returns to it.
    await waitFor(() => {
      expect(mockTaskCtx.request).toHaveBeenCalledTimes(1)
    })
    expect(mockTaskCtx.request).toHaveBeenCalledWith({
      definitionID: DEFINITION_ID,
    })
    expect(pathname).toEqual(`/tasks/${DEFINITION_ID}`)
  })
})
