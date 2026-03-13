import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "webchat-git",
  description: "GitHub Issue-based AI chat harness",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ko" className="dark">
      <body className="antialiased bg-zinc-950 text-zinc-100">
        {children}
      </body>
    </html>
  );
}
