// Date formatting helpers. All inputs are RFC3339 strings that may be empty.

function parse(s: string | undefined): Date | null {
  if (!s) return null;
  const d = new Date(s);
  // Go's zero time marshals as 0001-01-01; treat it as "no date".
  return isNaN(d.getTime()) || d.getFullYear() < 1970 ? null : d;
}

const rtf = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });

const UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
  ["year", 365 * 24 * 3600],
  ["month", 30 * 24 * 3600],
  ["week", 7 * 24 * 3600],
  ["day", 24 * 3600],
  ["hour", 3600],
  ["minute", 60],
];

// "3 days ago", "yesterday", "just now".
export function relativeTime(s: string | undefined): string {
  const d = parse(s);
  if (!d) return "";
  const secs = (d.getTime() - Date.now()) / 1000;
  for (const [unit, size] of UNITS) {
    if (Math.abs(secs) >= size) return rtf.format(Math.round(secs / size), unit);
  }
  return "just now";
}

// "Nov 4, 2026".
export function fmtDate(s: string | undefined): string {
  const d = parse(s);
  return d ? d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" }) : "";
}

// Day bucket for grouping the feed: "Today", "Yesterday", "Monday, Nov 2", or
// "Undated" for releases without a timestamp (e.g. plain git tags).
export function dayLabel(s: string | undefined): string {
  const d = parse(s);
  if (!d) return "Undated";
  const startOf = (x: Date) => new Date(x.getFullYear(), x.getMonth(), x.getDate()).getTime();
  const days = Math.round((startOf(new Date()) - startOf(d)) / 86_400_000);
  if (days === 0) return "Today";
  if (days === 1) return "Yesterday";
  const sameYear = d.getFullYear() === new Date().getFullYear();
  return d.toLocaleDateString(undefined, {
    weekday: days < 7 ? "long" : undefined,
    month: "short",
    day: "numeric",
    year: sameYear ? undefined : "numeric",
  });
}

// Release notes are markdown; strip the noisiest syntax for a plain excerpt.
export function plainText(md: string | undefined): string {
  return (md ?? "")
    .replace(/<!--[\s\S]*?-->/g, "")
    .replace(/!\[[^\]]*]\([^)]*\)/g, "")
    .replace(/\[([^\]]+)]\([^)]*\)/g, "$1")
    .replace(/^#{1,6}\s+/gm, "")
    .replace(/[*_`]{1,3}/g, "")
    .replace(/\r/g, "")
    .replace(/\n{3,}/g, "\n\n")
    .trim();
}
