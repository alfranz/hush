const { retry } = require("../src/retry");

describe("retry", () => {
  test("returns the first success", async () => {
    await expect(retry(async () => "ok")).resolves.toBe("ok");
  });

  test("retries until success", async () => {
    let calls = 0;
    const result = await retry(async () => {
      calls++;
      if (calls < 3) throw new Error("server error 500");
      return "ok";
    });
    expect(result).toBe("ok");
    expect(calls).toBe(3);
  });

  test("throws the last error", async () => {
    await expect(
      retry(async (i) => {
        throw new Error(`fail ${i}`);
      })
    ).rejects.toThrow(/^fail \d$/);
  });

  test("stops when retryOn returns false", async () => {
    let calls = 0;
    await expect(
      retry(
        async () => {
          calls++;
          throw new Error("bad request 400");
        },
        { retryOn: () => false }
      )
    ).rejects.toThrow("bad request 400");
    expect(calls).toBe(1);
  });
});
