import pytest

from shop.pricing import TAX_RATES, tax_cents


@pytest.mark.parametrize("country", sorted(TAX_RATES))
def test_tax_zero_amount(country):
    assert tax_cents(0, country) == 0


@pytest.mark.parametrize(
    "amount,country,expected",
    [
        (10000, "DE", 1900),
        (10000, "FR", 2000),
        (10000, "US", 0),
        (10000, "GB", 2000),
        (10000, "IT", 2200),
        (999, "DE", 190),
        (1, "IT", 0),
    ],
)
def test_tax(amount, country, expected):
    assert tax_cents(amount, country) == expected


def test_unknown_country():
    with pytest.raises(KeyError):
        tax_cents(100, "XX")
