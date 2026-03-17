import type { NextConfig } from "next";
const nextConfig: NextConfig = {
  serverExternalPackages: ['node-pty', '@anthropic-ai/claude-agent-sdk'],
};
export default nextConfig;
