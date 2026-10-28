# Flotilla UI

The Flotilla UI is a React application bundled along with the rest of Flotilla. If you are running the entire Flotilla stack locally, it is recommended to use docker-compose as documented in the main [README](https://github.com/stitchfix/flotilla-os#starting-the-service-locally). If you are interested in developing the UI itself, you can follow these steps:

## Stack

React 18, TypeScript 5, [Vite](https://vite.dev) for the dev server and build, [Vitest](https://vitest.dev) + [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/) for tests, Blueprint 5 for components. Node 22 or newer is required.

## Development

### Running Locally

```
git clone git@github.com:stitchfix/flotilla-os.git
cd flotilla-os/ui
npm ci
REACT_APP_BASE_URL=http://my-flotilla.com REACT_APP_BASE_URL_DEV=http://flotilla.staging.vertigo.stitchfix.com/api npm run dev
```

The dev server runs on http://localhost:3000. `REACT_APP_*` variables are read from the environment or from `.env` files and exposed to the app as `process.env.REACT_APP_*` (see `vite.config.ts`).

### Commands

| Command | What it does |
|---|---|
| `npm run dev` | Start the Vite dev server on port 3000 |
| `npm run typecheck` | Type-check with `tsc --noEmit` |
| `npm test` | Run the Vitest suite once (`npm run test:watch` to watch) |
| `npm run build` | Production build into `build/` |
| `npm run preview` | Serve the production build locally |

### Testing

Specs live in `__tests__/` folders next to the code and use React Testing Library. `src/test-utils.tsx` exports a `render` helper that wraps the component in a Redux `Provider` and a `MemoryRouter`.
