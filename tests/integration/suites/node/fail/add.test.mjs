import { test } from 'node:test';
import assert from 'node:assert/strict';

test('addition', () => {
  assert.equal(1 + 1, 2);
});

test('bad math', () => {
  assert.equal(1 + 1, 3);
});
