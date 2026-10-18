from dataclasses import dataclass, field


@dataclass
class Item:
    sku: str
    price_cents: int
    quantity: int = 1


@dataclass
class Cart:
    items: list[Item] = field(default_factory=list)

    def add(self, sku: str, price_cents: int, quantity: int = 1) -> None:
        if quantity <= 0:
            raise ValueError("quantity must be positive")
        for item in self.items:
            if item.sku == sku:
                item.quantity += quantity
                return
        self.items.append(Item(sku, price_cents, quantity))

    def remove(self, sku: str) -> None:
        self.items = [i for i in self.items if i.sku != sku]

    def count(self) -> int:
        return sum(i.quantity for i in self.items)

    def subtotal_cents(self) -> int:
        return sum(i.price_cents * i.quantity for i in self.items)
