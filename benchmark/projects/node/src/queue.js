class Queue {
  constructor() {
    this.items = [];
  }
  push(item) {
    this.items.push(item);
    return this.items.length;
  }
  shift() {
    return this.items.shift();
  }
  get size() {
    return this.items.length;
  }
}

module.exports = { Queue };
