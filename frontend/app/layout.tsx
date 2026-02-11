import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "HyperClaw Review",
  description: "Mobile-first trade review"
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
