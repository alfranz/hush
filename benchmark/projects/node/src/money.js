function formatCents(cents, currency = "EUR") {
  const sign = cents < 0 ? "-" : "";
  const abs = Math.abs(cents);
  const units = Math.floor(abs / 100);
  const rest = String(abs % 100).padStart(2, "0");
  return `${sign}${units}.${rest} ${currency}`;
}

function splitCents(total, parts) {
  if (parts <= 0) throw new Error("parts must be positive");
  const base = Math.floor(total / parts);
  const extra = total % parts;
  return Array.from({ length: parts }, (_, i) => base + (i < extra ? 1 : 0));
}

module.exports = { formatCents, splitCents };
