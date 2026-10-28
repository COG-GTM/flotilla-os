import * as React from "react"
import { Formik } from "formik"
import { render, screen, userEvent } from "../../test-utils"
import { EnvFieldArray } from "../EnvFieldArray"
import { Env } from "../../types"

describe("EnvFieldArray", () => {
  const values: Env[] = [
    { name: "a", value: "b" },
    { name: "c", value: "d" },
    { name: "e", value: "f" },
  ]
  const push = vi.fn()
  const remove = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    render(
      <Formik initialValues={{ env: values }} onSubmit={vi.fn()}>
        {() => (
          <EnvFieldArray
            values={values}
            push={push}
            remove={remove}
            errors={{}}
          />
        )}
      </Formik>
    )
  })

  const getRemoveButtons = () => {
    const addButton = screen.getByRole("button", { name: "Add" })
    return screen.getAllByRole("button").filter(b => b !== addButton)
  }

  it("renders props.values", () => {
    // Each item renders a name input, a value input and a remove button.
    expect(screen.getAllByRole("textbox")).toHaveLength(values.length * 2)
    values.forEach(env => {
      expect(screen.getByDisplayValue(env.name)).toBeInTheDocument()
      expect(screen.getByDisplayValue(env.value)).toBeInTheDocument()
    })
    expect(getRemoveButtons()).toHaveLength(values.length)
  })

  it("calls props.remove with the index of the item when clicked", async () => {
    const index = 1
    expect(remove).toHaveBeenCalledTimes(0)
    await userEvent.click(getRemoveButtons()[index])
    expect(remove).toHaveBeenCalledTimes(1)
    expect(remove).toHaveBeenCalledWith(index)
  })

  it("calls props.push with an empty env struct when the add button is clicked", async () => {
    expect(push).toHaveBeenCalledTimes(0)
    await userEvent.click(screen.getByRole("button", { name: "Add" }))
    expect(push).toHaveBeenCalledTimes(1)
    expect(push).toHaveBeenCalledWith({ name: "", value: "" })
  })
})
