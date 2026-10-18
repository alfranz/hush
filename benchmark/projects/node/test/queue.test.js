const { Queue } = require("../src/queue");

describe("Queue", () => {
  test("starts empty", () => {
    expect(new Queue().size).toBe(0);
  });

  test("push returns the new size", () => {
    const q = new Queue();
    expect(q.push("a")).toBe(1);
    expect(q.push("b")).toBe(2);
  });

  test("shift is first in, first out", () => {
    const q = new Queue();
    q.push("a");
    q.push("b");
    expect(q.shift()).toBe("a");
    expect(q.shift()).toBe("b");
  });

  test("shift on empty returns undefined", () => {
    expect(new Queue().shift()).toBeUndefined();
  });
});
