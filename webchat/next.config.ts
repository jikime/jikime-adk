import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  serverExternalPackages: ['node-pty', '@anthropic-ai/claude-agent-sdk'],
  // Gzip/Brotli compression for static assets and API responses
  compress: true,
  // Remove X-Powered-By header to avoid leaking server info
  poweredByHeader: false,
  experimental: {
    // Tree-shake lucide-react — only bundle icons actually used
    optimizePackageImports: ['lucide-react'],
  },
};

export default nextConfig;
