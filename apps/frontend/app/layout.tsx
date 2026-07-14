import type { Metadata } from "next";
import { display, body, mono } from "./fonts";
import { AmbientBackground } from "../components/ambient-bg";
import "../styles/globals.css";

export const metadata: Metadata = {
  title: "EcoMatrix — God's Eye",
  description: "Real-time observability for the EcoMatrix multi-agent economy.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html
      lang="zh-CN"
      className={`${display.variable} ${body.variable} ${mono.variable}`}
      suppressHydrationWarning
    >
      <body className="relative min-h-screen bg-canvas text-ink antialiased">
        <AmbientBackground />
        {children}
      </body>
    </html>
  );
}
