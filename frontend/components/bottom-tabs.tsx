"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const tabs = [
  { href: "/overview", label: "总览" },
  { href: "/trade", label: "交易" },
  { href: "/strategy", label: "策略" },
  { href: "/review", label: "复盘" },
  { href: "/me", label: "我的" }
];

export default function BottomTabs() {
  const pathname = usePathname();

  return (
    <nav className="bottom-tabs" aria-label="Bottom Navigation">
      {tabs.map((tab) => {
        const active = pathname === tab.href;
        return (
          <Link key={tab.href} href={tab.href} className={active ? "tab active" : "tab"}>
            {tab.label}
          </Link>
        );
      })}
    </nav>
  );
}
