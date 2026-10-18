const { formatCents, splitCents } = require("../src/money");

describe("formatCents", () => {
  test.each([
    [0, "0.00 EUR"],
    [5, "0.05 EUR"],
    [100, "1.00 EUR"],
    [123456, "1234.56 EUR"],
    [-250, "-2.50 EUR"],
  ])("%i -> %s", (cents, expected) => {
    expect(formatCents(cents)).toBe(expected);
  });

  test("other currency", () => {
    expect(formatCents(1999, "USD")).toBe("19.99 USD");
  });
});

describe("splitCents", () => {
  test.each([
    [100, 3, [34, 33, 33]],
    [100, 4, [25, 25, 25, 25]],
    [5, 2, [3, 2]],
    [0, 2, [0, 0]],
  ])("split %i into %i parts", (total, parts, expected) => {
    expect(splitCents(total, parts)).toEqual(expected);
  });

  test("rejects zero parts", () => {
    expect(() => splitCents(100, 0)).toThrow("parts must be positive");
  });
});
