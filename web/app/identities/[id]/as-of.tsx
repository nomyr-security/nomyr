"use client";

import {useSearchParams} from "next/navigation";
import {formatAsOf} from "../../../lib/i18n/date-time";

export function AsOf() {
  const searchParams = useSearchParams();
  const display = formatAsOf(searchParams.get("as_of"), {locale: "en", timeZone: "UTC"});
  return <p aria-live="polite">As of: {display} · timezone UTC</p>;
}
