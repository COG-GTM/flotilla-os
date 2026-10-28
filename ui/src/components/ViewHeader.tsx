import * as React from "react"
import { Link } from "react-router-dom"
import { Breadcrumbs, BreadcrumbProps, Classes } from "@blueprintjs/core"

type Props = {
  breadcrumbs: BreadcrumbProps[]
  buttons?: React.ReactNode
  leftButton?: React.ReactNode
}

const ViewHeader: React.FunctionComponent<Props> = ({
  breadcrumbs,
  buttons,
  leftButton,
}) => (
  <div className="flotilla-view-header-container">
    <div style={{ display: "flex" }}>
      {leftButton && leftButton}
      <Breadcrumbs
        items={breadcrumbs}
        breadcrumbRenderer={(props: BreadcrumbProps) => (
          <Link to={props.href ? props.href : "/"}>{props.text}</Link>
        )}
        className={Classes.TEXT_LARGE}
      />
    </div>
    {buttons}
  </div>
)

export default ViewHeader
