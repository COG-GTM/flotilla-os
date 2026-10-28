import { OverlayToaster, Position, ToastProps } from "@blueprintjs/core"
import { createRoot } from "react-dom/client"

const toaster = OverlayToaster.createAsync(
  { position: Position.BOTTOM_RIGHT },
  {
    domRenderer: (toasterElement, containerElement) => {
      createRoot(containerElement).render(toasterElement)
    },
  }
)

const Toaster = {
  show: (props: ToastProps, key?: string): Promise<string> =>
    toaster.then(t => t.show(props, key)),
}

export default Toaster
