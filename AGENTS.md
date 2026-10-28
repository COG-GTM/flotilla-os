# AGENTS.md

Flotilla is a Go service (repo root) with a React + TypeScript web UI in `ui/`.
This file holds the build and test rules for this repo and the target stack for the
**UI modernization**. The work plan itself is in Linear (see below).

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

## Work plan lives in Linear

The modernization plan is tracked in Linear (project "Stitch Fix demo: Flotilla UI modernization"): one coordinator epic plus one sub-issue per work package.
The issues define the steps, branch names, file ownership and definition of done. Read them with the Linear tools.

- If you're given the **epic**, you are the coordinator: follow its steps and start one managed Devin per sub-issue with the prompt `Work Linear issue <ID>.`
- If you're given a **sub-issue**, move it to In Progress, edit only the files it lists (comment on the issue if you need anything else), open your PR into the integration branch it names, link the PR on the issue, and move it to In Review.
- Status, PR links and the final rollup go on the Linear issues, so progress is visible there.

## Commands

UI after the migration (from `ui/`): `npm ci`, `npm run dev`, `npm run typecheck`, `npm test`, `npm run build`.
UI before the migration: `npm install`, `npm start`, `npm test`, `npm run build` (currently broken on a fresh install, see above).
Go service: see `README.md` (`docker-compose up -d`; needs AWS/EKS for real runs). CI (`.circleci/config.yml`) runs the Go tests only.
