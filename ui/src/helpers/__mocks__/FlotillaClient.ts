import {
  CreateTaskPayload,
  ListClustersResponse,
  ListGroupsResponse,
  ListRunParams,
  ListRunResponse,
  ListTagsResponse,
  ListTaskParams,
  ListTaskResponse,
  ListTaskRunsParams,
  ListTaskRunsResponse,
  Run,
  RunLog,
  RunStatus,
  LaunchRequestV2,
  Task,
  UpdateTaskPayload,
  ExecutionEngine,
  NodeLifecycle,
} from "../../types"
import { createMockRunObject, createMockTaskObject } from "../testHelpers"

// `api.ts` throws without a base URL; default it for specs using this mock.
if (process.env.REACT_APP_BASE_URL_DEV === undefined) {
  process.env.REACT_APP_BASE_URL_DEV = "http://localhost"
}

const getTask = vi.fn(
  ({ definitionID }: { definitionID: string }): Promise<Task> =>
    new Promise<Task>(resolve => {
      resolve(createMockTaskObject({ definition_id: definitionID }))
    })
)

const getTaskByAlias = vi.fn(
  ({ alias }: { alias: string }): Promise<Task> =>
    new Promise<Task>(resolve => {
      resolve(createMockTaskObject({ alias }))
    })
)

const listTaskRuns = vi.fn(
  ({
    definitionID,
    params,
  }: {
    definitionID: string
    params: ListTaskRunsParams
  }): Promise<ListTaskRunsResponse> =>
    new Promise<ListTaskRunsResponse>(resolve => {
      resolve({
        offset: params.offset,
        limit: params.limit,
        sort_by: params.sort_by,
        order: params.order,
        total: 0,
        history: [], // @TODO
        env_filters: {},
        cluster_name: params.cluster_name,
        status: params.status,
      })
    })
)

const listTasks = vi.fn(
  ({ params }: { params: ListTaskParams }): Promise<ListTaskResponse> =>
    new Promise<ListTaskResponse>(resolve => {
      resolve({
        offset: params.offset,
        limit: params.limit,
        sort_by: params.sort_by,
        order: params.order,
        total: 0,
        definitions: [], // @TODO
        alias: params.alias,
        group_name: params.group_name,
        image: params.image,
      })
    })
)

const createTask = vi.fn(
  ({ data }: { data: CreateTaskPayload }): Promise<Task> =>
    new Promise<Task>(resolve => {
      resolve(createMockTaskObject(data))
    })
)

const updateTask = vi.fn(
  ({
    definitionID,
    data,
  }: {
    definitionID: string
    data: UpdateTaskPayload
  }): Promise<Task> =>
    new Promise<Task>(resolve => {
      resolve(createMockTaskObject({ ...data, definition_id: definitionID }))
    })
)

const deleteTask = vi.fn(
  ({ definitionID }: { definitionID: string }): Promise<any> =>
    new Promise<any>(resolve => {
      resolve(true)
    })
)

const runTask = vi.fn(
  ({
    definitionID,
    data,
  }: {
    definitionID: string
    data: LaunchRequestV2
  }): Promise<Run> =>
    new Promise<Run>(resolve => {
      resolve(
        createMockRunObject({
          instance: {
            dns_name: "dns_name",
            instance_id: "instance_id",
          },
          task_arn: "task_arn",
          run_id: "run_id",
          definition_id: definitionID,
          alias: "alias",
          image: "image",
          cluster: data.cluster,
          status: RunStatus.PENDING,
          group_name: "group_name",
          env: data.env || [],
          cpu: 1,
          memory: 1024,
          command: "echo 'hi'",
          queued_at: "2019-10-24T05:21:51",
          engine: ExecutionEngine.ECS,
          node_lifecycle: NodeLifecycle.SPOT,
          max_cpu_used: 1,
          max_memory_used: 1,
          pod_name: "pod",
          cloudtrail_notifications: { Records: [] },
        })
      )
    })
)

const listRun = vi.fn(
  ({ params }: { params: ListRunParams }): Promise<ListRunResponse> =>
    new Promise<ListRunResponse>(resolve => {
      resolve({
        offset: params.offset,
        limit: params.limit,
        sort_by: params.sort_by,
        order: params.order,
        total: 0,
        history: [],
        env_filters: params.env,
        cluster_name: params.cluster_name,
        alias: params.alias,
        status: params.status,
      })
    })
)

const getRun = vi.fn(
  ({ runID }: { runID: string }): Promise<Run> =>
    new Promise<Run>(resolve => {
      resolve(
        createMockRunObject({
          instance: {
            dns_name: "dns_name",
            instance_id: "instance_id",
          },
          task_arn: "task_arn",
          run_id: runID,
          definition_id: "definition_id",
          alias: "alias",
          image: "image",
          cluster: "cluster",
          status: RunStatus.PENDING,
          group_name: "group_name",
          env: [],
          cpu: 1,
          memory: 1024,
          command: "echo 'hi'",
          queued_at: "2019-10-24T05:21:51",
          engine: ExecutionEngine.ECS,
          node_lifecycle: NodeLifecycle.SPOT,
          max_cpu_used: 1,
          max_memory_used: 1,
          pod_name: "pod",
          cloudtrail_notifications: { Records: [] },
        })
      )
    })
)

const getRunLog = vi.fn(
  ({
    runID,
    lastSeen = "",
  }: {
    runID: string
    lastSeen?: string
  }): Promise<RunLog> =>
    new Promise<RunLog>(resolve => {
      resolve({
        log: "",
        last_seen: lastSeen,
      })
    })
)

const stopRun = vi.fn(
  ({
    definitionID,
    runID,
  }: {
    definitionID: string
    runID: string
  }): Promise<any> =>
    new Promise<any>(resolve => {
      resolve(true)
    })
)

export const listClusters = vi.fn(
  (): Promise<ListClustersResponse> =>
    new Promise<ListClustersResponse>(resolve => {
      resolve({
        offset: 0,
        limit: 20,
        total: 0,
        clusters: ["a", "b", "c"],
      })
    })
)

const listGroups = vi.fn(
  (): Promise<ListGroupsResponse> =>
    new Promise<ListGroupsResponse>(resolve => {
      resolve({
        offset: 0,
        limit: 20,
        total: 0,
        groups: ["a", "b", "c"],
      })
    })
)

const listTags = vi.fn(
  (): Promise<ListTagsResponse> =>
    new Promise<ListTagsResponse>(resolve => {
      resolve({
        offset: 0,
        limit: 20,
        total: 0,
        tags: ["a", "b", "c"],
      })
    })
)

// `function` (not an arrow) so the mock can be constructed with `new`.
export default vi.fn(function FlotillaClient() {
  return {
    getTask,
    getTaskByAlias,
    listTaskRuns,
    listTasks,
    createTask,
    updateTask,
    deleteTask,
    runTask,
    listRun,
    getRun,
    getRunLog,
    stopRun,
    listClusters,
    listGroups,
    listTags,
  }
})
