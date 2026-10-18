/// Parses a duration like "90s", "5m" or "2h" into seconds.
pub fn duration_secs(s: &str) -> Option<u64> {
    let (num, unit) = s.split_at(s.len().checked_sub(1)?);
    let n: u64 = num.parse().ok()?;
    match unit {
        "s" => Some(n),
        "m" => Some(n * 60),
        "h" => Some(n * 3600),
        _ => None,
    }
}

/// Parses "major.minor.patch".
pub fn version(s: &str) -> Option<(u32, u32, u32)> {
    let mut parts = s.split('.').map(|p| p.parse::<u32>().ok());
    let v = (parts.next()??, parts.next()??, parts.next()??);
    if parts.next().is_some() {
        return None;
    }
    Some(v)
}

#[cfg(test)]
mod tests {
    use super::*;

    macro_rules! duration_tests {
        ($($name:ident: $input:expr => $want:expr,)*) => {
            $(#[test] fn $name() { assert_eq!(duration_secs($input), $want); })*
        };
    }

    duration_tests! {
        seconds: "90s" => Some(90),
        minutes: "5m" => Some(300),
        hours: "2h" => Some(7200),
        zero: "0s" => Some(0),
        bad_unit: "5d" => None,
        no_number: "m" => None,
        empty: "" => None,
        negative: "-5s" => None,
    }

    macro_rules! version_tests {
        ($($name:ident: $input:expr => $want:expr,)*) => {
            $(#[test] fn $name() { assert_eq!(version($input), $want); })*
        };
    }

    version_tests! {
        version_simple: "1.2.3" => Some((1, 2, 3)),
        version_zero: "0.0.0" => Some((0, 0, 0)),
        version_large: "10.20.300" => Some((10, 20, 300)),
        version_short: "1.2" => None,
        version_long: "1.2.3.4" => None,
        version_text: "a.b.c" => None,
    }
}
