import os

TAX_RATES = {"DE": 19, "FR": 20, "US": 0, "GB": 20, "IT": 22}


def tax_cents(amount_cents: int, country: str) -> int:
    rate = TAX_RATES[country]
    return round(amount_cents * rate / 100)


def discount_cents(amount_cents: int, code: str | None) -> int:
    if code is None:
        return 0
    if code == "SAVE10":
        return amount_cents // 10
    if code == "FLAT500":
        # BENCH_FAIL=1 adds a bug: the flat discount is applied twice.
        flat = 1000 if os.environ.get("BENCH_FAIL") else 500
        return min(flat, amount_cents)
    raise KeyError(f"unknown discount code: {code}")


def total_cents(amount_cents: int, country: str, code: str | None = None) -> int:
    discounted = amount_cents - discount_cents(amount_cents, code)
    return discounted + tax_cents(discounted, country)
