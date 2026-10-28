import * as React from "react"
import { Formik } from "formik"
import { render, screen } from "../../test-utils"
import {
  groupNameFieldSpec,
  imageFieldSpec,
  commandFieldSpec,
  memoryFieldSpec,
  tagsFieldSpec,
  envFieldSpec,
  cpuFieldSpec,
} from "../../helpers/taskFormHelpers"
import BaseTaskForm from "../BaseTaskForm"
import { UpdateTaskPayload } from "../../types"

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

describe("BaseTaskForm", () => {
  it("renders the correct fields", async () => {
    const initialValues: UpdateTaskPayload = {
      group_name: "my_group_name",
      image: "my_image",
      command: "my_command",
      memory: 1024,
      cpu: 512,
      tags: ["a", "b", "c"],
      env: [],
    }
    render(
      <Formik initialValues={initialValues} onSubmit={vi.fn()}>
        {({ values, setFieldValue, errors }) => (
          <BaseTaskForm
            values={values}
            setFieldValue={setFieldValue}
            errors={errors}
          />
        )}
      </Formik>
    )

    // Each of the six form groups renders its label and helper text.
    ;[
      groupNameFieldSpec,
      imageFieldSpec,
      commandFieldSpec,
      cpuFieldSpec,
      memoryFieldSpec,
      tagsFieldSpec,
    ].forEach(spec => {
      expect(screen.getByText(spec.label)).toBeInTheDocument()
      if (spec.description) {
        expect(screen.getByText(spec.description)).toBeInTheDocument()
      }
    })

    // Group name field (select) shows its initial value once groups load.
    expect(
      await screen.findByText(initialValues.group_name)
    ).toBeInTheDocument()

    // Image, command, cpu and memory fields are bound to their values.
    expect(screen.getByDisplayValue(initialValues.image)).toHaveAttribute(
      "name",
      imageFieldSpec.name
    )
    expect(screen.getByDisplayValue(initialValues.command)).toHaveAttribute(
      "name",
      commandFieldSpec.name
    )
    expect(screen.getByDisplayValue(String(initialValues.cpu))).toHaveAttribute(
      "name",
      cpuFieldSpec.name
    )
    expect(
      screen.getByDisplayValue(String(initialValues.memory))
    ).toHaveAttribute("name", memoryFieldSpec.name)

    // Tags field (multi select) shows each initial tag.
    initialValues.tags.forEach(tag => {
      expect(screen.getByText(tag)).toBeInTheDocument()
    })

    // The env field array is rendered.
    expect(screen.getByText(envFieldSpec.label)).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Add" })).toBeInTheDocument()

    // No field errors are rendered for valid initial values.
    expect(screen.queryByText(/required/i)).not.toBeInTheDocument()
  })
})
