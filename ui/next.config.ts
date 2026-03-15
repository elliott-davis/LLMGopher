import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://gateway:8080/:path*",
      },
    ];
  },
};

export default nextConfig;
