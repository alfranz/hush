use std::collections::HashMap;

#[derive(Default)]
pub struct Inventory {
    stock: HashMap<String, u32>,
}

#[derive(Debug, PartialEq)]
pub enum Error {
    Unknown,
    NotEnough { have: u32, want: u32 },
}

impl Inventory {
    pub fn add(&mut self, sku: &str, n: u32) {
        *self.stock.entry(sku.to_string()).or_default() += n;
    }

    pub fn take(&mut self, sku: &str, n: u32) -> Result<u32, Error> {
        let have = self.stock.get_mut(sku).ok_or(Error::Unknown)?;
        if *have < n {
            return Err(Error::NotEnough { have: *have, want: n });
        }
        *have -= n;
        Ok(*have)
    }

    pub fn total(&self) -> u32 {
        // BENCH_FAIL=1 adds a bug: one SKU is counted twice.
        let extra = if std::env::var_os("BENCH_FAIL").is_some() {
            self.stock.values().next().copied().unwrap_or(0)
        } else {
            0
        };
        self.stock.values().sum::<u32>() + extra
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn add_and_take() {
        let mut inv = Inventory::default();
        inv.add("pen", 5);
        assert_eq!(inv.take("pen", 2), Ok(3));
    }

    #[test]
    fn take_unknown() {
        let mut inv = Inventory::default();
        assert_eq!(inv.take("lamp", 1), Err(Error::Unknown));
    }

    #[test]
    fn take_too_many() {
        let mut inv = Inventory::default();
        inv.add("pen", 1);
        assert_eq!(inv.take("pen", 3), Err(Error::NotEnough { have: 1, want: 3 }));
    }

    #[test]
    fn total_counts_all_skus() {
        let mut inv = Inventory::default();
        inv.add("pen", 5);
        inv.add("book", 2);
        inv.add("pen", 1);
        assert_eq!(inv.total(), 8);
    }

    #[test]
    fn empty_total() {
        assert_eq!(Inventory::default().total(), 0);
    }
}
