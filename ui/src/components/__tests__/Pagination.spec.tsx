import * as React from "react"
import { Classes } from "@blueprintjs/core"
import { render, screen, userEvent } from "../../test-utils"
import Pagination, { Props } from "../Pagination"

const DEFAULT_PROPS: Props = {
  updatePage: () => {},
  currentPage: 1,
  numItems: 100,
  pageSize: 20,
  isLoading: false,
}

const renderPagination = (props: Partial<Props> = {}) =>
  render(<Pagination {...DEFAULT_PROPS} {...props} />)

const getPrevButton = () => screen.getAllByRole("button")[0]
const getNextButton = () => screen.getAllByRole("button")[1]

describe("Pagination", () => {
  it("renders two buttons", () => {
    const { container } = renderPagination()
    expect(container.querySelectorAll(`.${Classes.BUTTON_GROUP}`)).toHaveLength(1)
    expect(screen.getAllByRole("button")).toHaveLength(2)
  })

  it("disables the previous button if on the first page", () => {
    renderPagination({ currentPage: 1 })
    expect(getPrevButton()).toBeDisabled()
  })

  it("disables the next button if on the last page", () => {
    const { rerender } = renderPagination({ numItems: 113, currentPage: 5 })
    expect(getNextButton()).toBeEnabled()

    rerender(<Pagination {...DEFAULT_PROPS} numItems={113} currentPage={6} />)
    expect(getNextButton()).toBeDisabled()
  })

  it("calls props.updatePage when the prev or next buttons are clicked", async () => {
    const user = userEvent.setup()
    const updatePage = vi.fn()

    const { rerender } = renderPagination({ updatePage, currentPage: 1 })
    expect(updatePage).toHaveBeenCalledTimes(0)
    await user.click(getNextButton())
    expect(updatePage).toHaveBeenCalledTimes(1)
    expect(updatePage).toHaveBeenCalledWith(2)

    rerender(
      <Pagination {...DEFAULT_PROPS} updatePage={updatePage} currentPage={2} />
    )
    await user.click(getPrevButton())
    expect(updatePage).toHaveBeenCalledTimes(2)
    expect(updatePage).toHaveBeenCalledWith(1)
  })
})
