# AGENTS.md

Flotilla is a Go service (repo root) with a React + TypeScript web UI in `ui/`.
This file tells Devin how to work in this repo, and in particular how to run the
**UI modernization** as a coordinated, parallel effort.

## UI today (`ui/`)

- React 16.8, TypeScript 3.4.5, Blueprint 3.15, `react-scripts` (Create React App), Enzyme 3.9 + Jest.
- About 107 `.ts`/`.tsx` files (~10.3k lines) and 22 Enzyme spec files in `__tests__/` folders.
- No lockfile is committed. On a fresh `npm install` today, `npm run build` fails
  (TypeScript 3.4 cannot parse current `@types/*` files) and all 22 test suites fail to start
  (Enzyme's `cheerio` dependency needs `TextDecoder`). This is dependency drift, not a product bug.

## UI modernization target

| Area | From | To |
|---|---|---|
| React / ReactDOM | 16.8 | 18.3 (`createRoot`) |
| TypeScript | 3.4.5 | ~5.9, `"moduleResolution": "Bundler"`, `"jsx": "react-jsx"` |
| Build / dev server | `react-scripts` | Vite (`vite.config.ts`, `index.html` at `ui/` root, `build/` outDir, port 3000) |
| Tests | Jest + Enzyme | Vitest (jsdom, globals) + React Testing Library + `@testing-library/user-event` + `@testing-library/jest-dom` |
| Blueprint | 3.15 | `@blueprintjs/core` / `@blueprintjs/icons` 5.x (do not use 6.x; `@blueprintjs/datetime` is unused, remove it) |
| react-select | 2.4 | 5.x (`react-select/lib/Creatable` -> `react-select/creatable`, `react-select/lib/types` -> `react-select`) |
| formik | 1.5 | 2.x |
| react-redux | 7 | 8.x |
| react-json-view | 1.x | `@microlink/react-json-view` |
| react-router-dom | 5.1 | 5.3.4 (stay on v5; v6 is out of scope) |

Rules for the migration:
- Keep behavior identical. No redesigns, no new features, no unrelated refactors.
- Keep `process.env.REACT_APP_*` / `process.env.NODE_ENV` reads working; the foundation exposes them through `define` in `vite.config.ts`, so feature code does not change for env vars.
- Port every Enzyme spec to React Testing Library. Keep each original test's intent; query by role/label/text, not by component type. Don't delete a test unless the behavior it covers no longer exists, and say so in the PR.
- No `any` casts or `@ts-ignore` added just to silence the upgrade. If one is unavoidable, add a one-line reason.

## How to run it in parallel (coordinator + managed Devins)

When asked to modernize/upgrade the UI, act as the **coordinator**:

1. **Foundation (coordinator only).** Branch `devin/ui-modernization` off `master`. Own and change only:
   `ui/package.json` (+ commit `ui/package-lock.json`), `ui/tsconfig.json`, `ui/vite.config.ts`, `ui/index.html`,
   `ui/src/index.tsx`, `ui/src/setupTests.ts` (replaces `setupTests.js`), `ui/src/vite-env.d.ts` (replaces `react-app-env.d.ts`),
   `ui/src/test-utils.tsx` (an RTL `render` helper that wraps Redux `Provider` + `MemoryRouter`), `ui/Dockerfile`, `ui/README.md`.
   Done when `npm ci` and `npx vite build` succeed. Type errors and failing specs in feature files are expected at this point.
   Record the baseline: total `tsc` error count and per-package counts. Push the branch.
2. **Fan out.** Start one managed Devin per work package below, all in parallel, each branching off
   `devin/ui-modernization` and opening its PR **into `devin/ui-modernization`** (never into `master`).
   Tag the sessions `flotilla-ui-modernization`.
3. **Merge and integrate.** As child PRs arrive, review them against the rules above and merge them into
   `devin/ui-modernization`. Resolve any conflicts yourself.
4. **Verify the whole UI.** `npm ci && npm run typecheck && npm test && npm run build` all pass on the integration branch. Boot `npm run dev` and screenshot the Tasks, Templates and Runs pages; without a backend, API error callouts are acceptable, but there must be no render crash or console error from React/Blueprint.
5. **Open one PR** from `devin/ui-modernization` into `master` with the rollup below. Do not merge it.

### Work packages (file ownership)

A child edits only its own files (and their specs). If it needs a change in a file it doesn't own, it asks the coordinator instead.
Any `ui/src` file not listed belongs to **shell-data**.

| Package | Files under `ui/src` |
|---|---|
| **tasks-forms** | `components/`: BaseTaskForm, CreateTaskForm, UpdateTaskForm, TaskExecutionForm, EnvFieldArray, EnvList, EnvQueryFilter, FieldError, GroupNameSelect, TagsSelect, ClusterSelect, GenericMultiSelect, NodeLifecycleSelect, Task, TaskDetails, Tasks, DeleteTaskButton; `helpers/`: selectHelpers, taskFormHelpers, runFormHelpers, getInitialValuesForExecutionForm |
| **templates** | `components/`: Template, TemplateDetails, TemplateExecutionForm, TemplateHistoryTable, TemplateRunForm, Templates; `helpers/`: constructDefaultObjectFromJsonSchema |
| **runs** | `components/`: Run, Runs, RunAttributes, RunDebugAttributes, RunEvents, RunSidebar, RunStatusSelect, RunTag, StopRunButton, TaskRuns, CloudtrailRecords, Duration, EngineTag, ResourceUsageValue, ISO8601AttributeValue, ARASwitch; `helpers/`: calculateDuration, getEnhancedRunStatus, getOwnerIdRunTagFromCookies |
| **logs** | `components/`: Log, LogProcessor, LogRequesterCloudWatchLogs, LogRequesterS3, LogVirtualized, LogVirtualizedRow, LogVirtualizedSearch, AutoscrollSwitch; `workers/` |
| **shell-data** | `components/`: App, Navigation, ViewHeader, SettingsButton, Toaster, Toggler, Table, SortableTh, Pagination, QueryParams, Request, ListRequest, ListFiltersDropdown, Attribute, ErrorCallout; `api.ts`, `constants.ts`, `types.ts`, `localstorage.ts`, `state/`, `helpers/FlotillaClient.ts`, `helpers/__mocks__/`, `helpers/testHelpers.ts`, `helpers/pageToOffsetLimit.ts` |

A spec belongs to the package that owns the file it tests (e.g. `components/__tests__/Run.spec.tsx` -> runs).

### Child definition of done

- `npx tsc --noEmit -p ui` reports **zero errors in the package's files** (errors in other packages' files are expected while siblings are still working; the coordinator checks the full count after merging).
- `npx vitest run <the package's spec files>` passes from `ui/`.
- PR into `devin/ui-modernization` whose description lists: files changed, tsc errors in the package before -> after, specs ported, any behavior questions.

### Rollup (coordinator's final PR description and report)

| Package | Session | PR | tsc errors before -> after | Specs ported / passing | ACUs | Wall time |
|---|---|---|---|---|---|---|

Plus totals, the integration-branch commands and their results, the screenshots, and anything that needs a human decision.

## Commands

UI after the migration (from `ui/`): `npm ci`, `npm run dev`, `npm run typecheck`, `npm test`, `npm run build`.
UI before the migration: `npm install`, `npm start`, `npm test`, `npm run build` (currently broken on a fresh install, see above).
Go service: see `README.md` (`docker-compose up -d`; needs AWS/EKS for real runs). CI (`.circleci/config.yml`) runs the Go tests only.
