import * as React from "react"
import { Route } from "react-router-dom"
import { render, screen, userEvent, waitFor } from "../../test-utils"
import CreateTaskForm, { ConnectedProps } from "../CreateTaskForm"
import api from "../../api"
import { CreateTaskPayload } from "../../types"

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

describe("CreateTaskForm", () => {
  const initialValues: CreateTaskPayload = {
    env: [{ name: "foo", value: "bar" }],
    image: "my_image",
    group_name: "my_group",
    alias: "my_alias",
    memory: 1024,
    command: "my_command",
    tags: ["a", "b"],
    cpu: 512,
  }
  const onSuccess = vi.fn()
  let pathname = ""

  beforeEach(() => {
    vi.clearAllMocks()
    render(
      <Route
        path="/"
        render={(routeProps: Omit<ConnectedProps, "initialValues">) => {
          pathname = routeProps.location.pathname
          return (
            <CreateTaskForm
              {...routeProps}
              initialValues={initialValues}
              onSuccess={onSuccess}
            />
          )
        }}
      />,
      { initialEntries: ["/tasks/create"] }
    )
  })

  it("renders the correct components", () => {
    // The request is only made on submit, not on mount.
    expect(api.createTask).toHaveBeenCalledTimes(0)
    // Alias field plus the BaseTaskForm fields, bound to initial values.
    expect(screen.getByDisplayValue(initialValues.alias)).toHaveAttribute(
      "name",
      "alias"
    )
    expect(screen.getByDisplayValue(initialValues.image)).toHaveAttribute(
      "name",
      "image"
    )
    expect(screen.getByRole("button", { name: "Submit" })).toHaveAttribute(
      "id",
      "submitButton"
    )
  })

  it("calls api.createTask when submitted", async () => {
    expect(api.createTask).toHaveBeenCalledTimes(0)
    expect(onSuccess).toHaveBeenCalledTimes(0)
    expect(pathname).toEqual("/tasks/create")

    await userEvent.click(screen.getByRole("button", { name: "Submit" }))

    await waitFor(() => {
      expect(api.createTask).toHaveBeenCalledTimes(1)
    })
    expect(api.createTask).toHaveBeenCalledWith({ data: initialValues })

    // On success, onSuccess is invoked and the user is sent to the new task.
    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledTimes(1)
    })
    expect(pathname).toEqual("/tasks/my_definition_id")
  })
})
