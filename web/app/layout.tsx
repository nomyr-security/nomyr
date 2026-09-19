import type { Metadata, Viewport } from "next";
import "./styles.css";

const assetPath = "";
export const viewport: Viewport = { themeColor: "#0C3B43" };

export const metadata: Metadata = {
  title: "Nomyr",
  description: "Non-human identity security platform",
  manifest: `${assetPath}/site.webmanifest`,
  icons: {
    icon: [
      { url: `${assetPath}/favicon.ico`, sizes: "16x16 32x32 48x48" },
      { url: `${assetPath}/favicon.svg`, type: "image/svg+xml" },
    ],
    apple: [{ url: `${assetPath}/apple-touch-icon.png`, sizes: "180x180" }],
    other: [{ rel: "mask-icon", url: `${assetPath}/safari-pinned-tab.svg`, color: "#0C3B43" }],
  },
};

export default function RootLayout({children}: Readonly<{children: React.ReactNode}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
