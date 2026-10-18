import pytest

from shop.pricing import discount_cents, total_cents


@pytest.mark.parametrize("amount", [0, 100, 999, 10000])
def test_no_code(amount):
    assert discount_cents(amount, None) == 0


@pytest.mark.parametrize("amount,expected", [(1000, 100), (999, 99), (5, 0)])
def test_save10(amount, expected):
    assert discount_cents(amount, "SAVE10") == expected


@pytest.mark.parametrize("amount,expected", [(10000, 500), (300, 300)])
def test_flat500(amount, expected):
    assert discount_cents(amount, "FLAT500") == expected


def test_unknown_code():
    with pytest.raises(KeyError):
        discount_cents(100, "NOPE")


@pytest.mark.parametrize(
    "amount,country,code,expected",
    [
        (10000, "DE", None, 11900),
        (10000, "DE", "SAVE10", 10710),
        (10000, "US", None, 10000),
    ],
)
def test_total(amount, country, code, expected):
    assert total_cents(amount, country, code) == expected
