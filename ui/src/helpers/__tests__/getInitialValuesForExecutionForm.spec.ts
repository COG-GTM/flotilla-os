import { getInitialValuesForTaskExecutionForm } from "../getInitialValuesForExecutionForm"
import getOwnerIdRunTagFromCookies from "../getOwnerIdRunTagFromCookies"
import { createMockTaskObject } from "../testHelpers"
import {
  LaunchRequestV2,
  ExecutionEngine,
  DefaultNodeLifecycle,
} from "../../types"

describe("getInitialValuesForTaskExecutionForm", () => {
  it("works correctly", () => {
    const td = createMockTaskObject()
    const expectedEks: LaunchRequestV2 = {
      cluster: process.env.REACT_APP_EKS_CLUSTER_NAME || "",
      cpu: td.cpu,
      memory: td.memory,
      env: td.env,
      engine: ExecutionEngine.EKS,
      command: td.command,
      node_lifecycle: DefaultNodeLifecycle,
      owner_id: getOwnerIdRunTagFromCookies(),
    }

    expect(getInitialValuesForTaskExecutionForm(td, null)).toEqual(expectedEks)
  })
})
