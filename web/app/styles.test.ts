import {readFileSync} from "node:fs";
import {resolve} from "node:path";
import {describe, expect, it} from "vitest";

const stylesheet = readFileSync(resolve(process.cwd(), "app/styles.css"), "utf8");

function token(name: string) {
  const value = stylesheet.match(new RegExp(`--${name}:\\s*(#[0-9a-fA-F]{6})`))?.[1];
  if (!value) throw new Error(`Missing color token ${name}`);
  return value;
}

function luminance(hex: string) {
  const channels = hex.slice(1).match(/.{2}/g)!.map((channel) => parseInt(channel, 16) / 255)
    .map((channel) => channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4);
  return channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
}

function contrast(foreground: string, background: string) {
  const values = [luminance(foreground), luminance(background)].sort((left, right) => right - left);
  return (values[0] + 0.05) / (values[1] + 0.05);
}

describe("interface color contrast tokens", () => {
  it("meets WCAG AA for body, muted, and accent text on their intended surfaces", () => {
    expect(contrast(token("text"), token("canvas"))).toBeGreaterThanOrEqual(4.5);
    expect(contrast(token("muted"), token("canvas"))).toBeGreaterThanOrEqual(4.5);
    expect(contrast(token("accent"), token("canvas"))).toBeGreaterThanOrEqual(4.5);
    expect(contrast(token("text"), token("surface"))).toBeGreaterThanOrEqual(4.5);
    expect(contrast(token("danger"), token("surface"))).toBeGreaterThanOrEqual(4.5);
    expect(contrast(token("warning"), token("surface"))).toBeGreaterThanOrEqual(4.5);
  });
});
