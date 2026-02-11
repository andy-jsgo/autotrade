import "./globals.css";
import type { Metadata } from "next";
import Providers from "../components/providers";
import BottomTabs from "../components/bottom-tabs";

export const metadata: Metadata = {
  title: "HyperClaw Trade Console",
  description: "Mobile-first trading, strategy status, and review console"
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Providers>
          {children}
          <BottomTabs />
        </Providers>
      </body>
    </html>
  );
}
