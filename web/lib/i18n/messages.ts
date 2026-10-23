export const messages = {
  en: {
    appName: "Nomyr",
    demoLabel: "Local development demo",
    syntheticOnly: "Synthetic data only. Live provider connections and security actions are not implemented.",
    asOf: "As of",
  },
} as const;

export type Locale = keyof typeof messages;
export function getMessages(locale: string = "en") {
  return messages[locale as Locale] ?? messages.en;
}
