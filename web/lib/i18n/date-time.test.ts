import {describe, expect, it} from "vitest";

import {formatAsOf, formatDateTime} from "./date-time";
import {getMessages} from "./messages";

describe("locale and timezone helpers", () => {
  it("formats an instant in the explicitly supplied timezone", () => {
    const utc = formatDateTime("2025-01-15T12:00:00Z", {locale: "en-GB", timeZone: "UTC"});
    const india = formatDateTime("2025-01-15T12:00:00Z", {locale: "en-GB", timeZone: "Asia/Kolkata"});
    expect(utc).not.toEqual(india);
    expect(india).toContain("17:30");
  });

  it("falls back to the English message catalog", () => {
    expect(getMessages("untranslated").appName).toBe("Nomyr");
  });

  it("requires an explicit offset for the static route as_of value", () => {
    expect(formatAsOf("2025-01-15T12:00:00", {locale: "en-GB", timeZone: "UTC"})).toContain("include timezone offset");
    expect(formatAsOf("2025-01-15T12:00:00Z", {locale: "en-GB", timeZone: "UTC"})).toContain("12:00");
    expect(formatAsOf(null, {locale: "en-GB", timeZone: "UTC"})).toBe("latest synthetic snapshot");
  });
});
