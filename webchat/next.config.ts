import type { NextConfig } from "next";

const securityHeaders = [
  // Prevent browsers from MIME-sniffing the content type
  { key: 'X-Content-Type-Options', value: 'nosniff' },
  // Block iframe embedding (clickjacking protection)
  { key: 'X-Frame-Options', value: 'DENY' },
  // Strict referrer to avoid leaking URL info to third parties
  { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
  // Prevent XSS via basic CSP (allows self + inline styles for Tailwind + WebSocket)
  {
    key: 'Content-Security-Policy',
    value: [
      "default-src 'self'",
      "script-src 'self' 'unsafe-inline' 'unsafe-eval'",   // Next.js hydration requires unsafe-inline/eval
      "style-src 'self' 'unsafe-inline'",                   // Tailwind uses inline styles
      "img-src 'self' data: blob:",
      "font-src 'self' data:",
      "connect-src 'self' ws: wss: http://localhost:* http://127.0.0.1:*",
      "worker-src 'self' blob:",
      "frame-ancestors 'none'",
    ].join('; '),
  },
  // Disable browser features not needed by the app
  { key: 'Permissions-Policy', value: 'camera=(), microphone=(self), geolocation=()' },
]

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
  async headers() {
    return [
      {
        source: '/:path*',
        headers: securityHeaders,
      },
    ]
  },
};

export default nextConfig;
