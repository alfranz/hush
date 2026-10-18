async function retry(fn, { attempts = 3, retryOn = () => true } = {}) {
  let lastError;
  // BENCH_FAIL=1 adds a bug: one attempt too few.
  const tries = process.env.BENCH_FAIL ? attempts - 1 : attempts;
  for (let i = 0; i < tries; i++) {
    try {
      return await fn(i);
    } catch (err) {
      lastError = err;
      if (!retryOn(err)) break;
    }
  }
  throw lastError;
}

module.exports = { retry };
