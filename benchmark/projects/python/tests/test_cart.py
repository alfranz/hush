import pytest

from shop.cart import Cart


def test_empty_cart():
    cart = Cart()
    assert cart.count() == 0
    assert cart.subtotal_cents() == 0


@pytest.mark.parametrize("quantity", [1, 2, 3, 5, 8, 13])
def test_add_quantity(quantity):
    cart = Cart()
    cart.add("book", 1200, quantity)
    assert cart.count() == quantity
    assert cart.subtotal_cents() == 1200 * quantity


def test_add_same_sku_merges():
    cart = Cart()
    cart.add("pen", 150)
    cart.add("pen", 150, 2)
    assert len(cart.items) == 1
    assert cart.count() == 3


@pytest.mark.parametrize("quantity", [0, -1, -10])
def test_add_invalid_quantity(quantity):
    with pytest.raises(ValueError):
        Cart().add("pen", 150, quantity)


def test_remove():
    cart = Cart()
    cart.add("pen", 150)
    cart.add("book", 1200)
    cart.remove("pen")
    assert [i.sku for i in cart.items] == ["book"]


def test_remove_missing_is_noop():
    cart = Cart()
    cart.add("pen", 150)
    cart.remove("lamp")
    assert cart.count() == 1
