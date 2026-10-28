import * as React from "react"
import { Link, NavLink } from "react-router-dom"
import {
  ButtonGroup,
  Navbar,
  NavbarDivider,
  NavbarGroup,
  Alignment,
  Classes,
  Tag,
  Intent,
} from "@blueprintjs/core"
import SettingsButton from "./SettingsButton"

const Navigation: React.FunctionComponent = () => (
  <Navbar fixedToTop className={Classes.DARK}>
    <NavbarGroup align={Alignment.LEFT}>
      <Link to="/tasks" className={`${Classes.BUTTON} ${Classes.MINIMAL}`}>
        Flotilla
      </Link>
      <NavbarDivider />
      <ButtonGroup className={Classes.MINIMAL}>
        <NavLink
          to="/tasks"
          className={Classes.BUTTON}
          activeClassName={Classes.ACTIVE}
        >
          Tasks
        </NavLink>
        <NavLink
          to="/templates"
          className={Classes.BUTTON}
          activeClassName={Classes.ACTIVE}
        >
          <span>Templates</span>
          <Tag intent={Intent.DANGER}>New!</Tag>
        </NavLink>
        <NavLink
          to="/runs"
          className={Classes.BUTTON}
          activeClassName={Classes.ACTIVE}
        >
          Runs
        </NavLink>
      </ButtonGroup>
    </NavbarGroup>
    <NavbarGroup align={Alignment.RIGHT}>
      <ButtonGroup>
        <SettingsButton />
      </ButtonGroup>
    </NavbarGroup>
  </Navbar>
)

export default Navigation
