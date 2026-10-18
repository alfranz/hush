/// Formats a byte count with binary units.
pub fn bytes(n: u64) -> String {
    const UNITS: [&str; 5] = ["B", "KiB", "MiB", "GiB", "TiB"];
    let mut value = n as f64;
    let mut unit = 0;
    while value >= 1024.0 && unit < UNITS.len() - 1 {
        value /= 1024.0;
        unit += 1;
    }
    if unit == 0 {
        format!("{n} B")
    } else {
        format!("{value:.1} {}", UNITS[unit])
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    macro_rules! bytes_tests {
        ($($name:ident: $input:expr => $want:expr,)*) => {
            $(#[test] fn $name() { assert_eq!(bytes($input), $want); })*
        };
    }

    bytes_tests! {
        zero_bytes: 0 => "0 B",
        small: 512 => "512 B",
        one_kib: 1024 => "1.0 KiB",
        kib: 1536 => "1.5 KiB",
        one_mib: 1024 * 1024 => "1.0 MiB",
        gib: 3 * 1024 * 1024 * 1024 => "3.0 GiB",
        tib: 2 * 1024u64.pow(4) => "2.0 TiB",
        huge: 5000 * 1024u64.pow(4) => "5000.0 TiB",
    }
}
