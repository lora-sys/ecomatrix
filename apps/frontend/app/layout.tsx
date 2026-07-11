import type { Metadata } from "next";
import "../styles/globals.css";

export const metadata: Metadata = {
  title: "EcoMatrix — God's Eye",
  description: "Real-time observability for the EcoMatrix multi-agent economy.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <body className="min-h-screen bg-canvas text-ink antialiased">
        {children}
      </body>
    </html>
  );
}
