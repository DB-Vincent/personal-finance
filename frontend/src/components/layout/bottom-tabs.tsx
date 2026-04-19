import {
  LayoutDashboard,
  Wallet,
  ArrowLeftRight,
  PiggyBank,
  MoreHorizontal,
} from "lucide-react";
import { Link, useLocation } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { cn } from "@/lib/utils";

const tabs = [
  { key: "dashboard", icon: LayoutDashboard, href: "/" },
  { key: "accounts", icon: Wallet, href: "/accounts" },
  { key: "transactions", icon: ArrowLeftRight, href: "/transactions" },
  { key: "budgets", icon: PiggyBank, href: "/budgets" },
  { key: "settings", icon: MoreHorizontal, href: "/settings" },
] as const;

export function BottomTabs() {
  const { t } = useTranslation();
  const location = useLocation();

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 border-t bg-background md:hidden">
      <div className="flex items-center justify-around h-16">
        {tabs.map((tab) => {
          const isActive = location.pathname === tab.href;
          return (
            <Link
              key={tab.key}
              to={tab.href as "/"}
              className={cn(
                "flex flex-col items-center gap-1 px-3 py-2 text-xs",
                isActive
                  ? "text-primary"
                  : "text-muted-foreground"
              )}
            >
              <tab.icon className="h-5 w-5" />
              <span>{t(tab.key)}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
