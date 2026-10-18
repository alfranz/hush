const { slugify } = require("../src/slug");

describe("slugify", () => {
  test.each([
    ["Hello World", "hello-world"],
    ["  spaces  around ", "spaces-around"],
    ["Crème brûlée", "creme-brulee"],
    ["a--b__c", "a-b-c"],
    ["123 Go!", "123-go"],
    ["", ""],
    ["UPPER lower", "upper-lower"],
    ["dots.and.more", "dots-and-more"],
  ])("%j -> %j", (input, expected) => {
    expect(slugify(input)).toBe(expected);
  });
});
