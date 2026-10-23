import type { NextConfig } from "next";

const config: NextConfig = {
  reactStrictMode: true,
  trailingSlash: true,
  turbopack: {
    root: process.cwd(),
  },
};

if (process.env.NODE_ENV === "production") {
  config.output = "export";
} else {
  config.skipTrailingSlashRedirect = true;
  config.rewrites = async () => [
    { source: "/api/v1/:path*", destination: "http://127.0.0.1:8080/api/v1/:path*" },
  ];
}

export default config;
