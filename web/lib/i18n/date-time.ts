export type DateTimeOptions = {locale: string; timeZone: string};

export function formatAsOf(value: string | null, options: DateTimeOptions): string {
  if (value === null) return "latest synthetic snapshot";
  if (!/(?:[zZ]|[+-]\d{2}:\d{2})$/.test(value)) return `${value} (unparsed; include timezone offset)`;
  const timestamp = Date.parse(value);
  return Number.isNaN(timestamp) ? `${value} (unparsed)` : formatDateTime(timestamp, options);
}

export function formatDateTime(value: string | number | Date, {locale, timeZone}: DateTimeOptions): string {
  return new Intl.DateTimeFormat(locale, {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone,
  }).format(new Date(value));
}
