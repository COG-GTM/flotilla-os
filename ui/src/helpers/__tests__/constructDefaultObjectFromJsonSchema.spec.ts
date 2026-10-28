import constructDefaultObjectFromJsonSchema from "../constructDefaultObjectFromJsonSchema"

describe("constructDefaultObjectFromJsonSchema", () => {
  it("returns an empty object when the schema has no properties", () => {
    expect(constructDefaultObjectFromJsonSchema({})).toEqual({})
    expect(constructDefaultObjectFromJsonSchema({ type: "object" })).toEqual({})
  })

  it("uses a type-specific fallback for properties without a default", () => {
    expect(
      constructDefaultObjectFromJsonSchema({
        type: "object",
        properties: {
          s: { type: "string" },
          n: { type: "number" },
          b: { type: "boolean" },
          a: { type: "array" },
          i: { type: "integer" },
        },
      })
    ).toEqual({ s: "", n: 0, b: false, a: [], i: null })
  })

  it("uses the property's default when one is set", () => {
    expect(
      constructDefaultObjectFromJsonSchema({
        properties: {
          s: { type: "string", default: "hello" },
          n: { type: "number", default: 42 },
          b: { type: "boolean", default: true },
          a: { type: "array", default: [1, 2] },
          i: { type: "integer", default: 7 },
        },
      })
    ).toEqual({ s: "hello", n: 42, b: true, a: [1, 2], i: 7 })
  })

  it("recurses into nested objects", () => {
    expect(
      constructDefaultObjectFromJsonSchema({
        properties: {
          outer: {
            type: "object",
            properties: {
              inner: { type: "string", default: "x" },
              deeper: { type: "object", properties: { n: { type: "number" } } },
            },
          },
          empty: { type: "object" },
        },
      })
    ).toEqual({ outer: { inner: "x", deeper: { n: 0 } }, empty: {} })
  })

  it("skips properties without a type", () => {
    expect(
      constructDefaultObjectFromJsonSchema({
        properties: { untyped: { default: "ignored" }, s: { type: "string" } },
      })
    ).toEqual({ s: "" })
  })
})
