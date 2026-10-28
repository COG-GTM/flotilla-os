import "@testing-library/jest-dom/vitest"

// Temporary: `helpers/testHelpers.ts` still calls `jest.fn()` at import time.
// Remove once GOL-257 (shell-data) ports it to `vi.fn()`.
Object.assign(globalThis, { jest: vi })
