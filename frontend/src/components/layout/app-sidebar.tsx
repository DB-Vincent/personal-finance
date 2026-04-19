import {
  LayoutDashboard,
  Wallet,
  ArrowLeftRight,
  PiggyBank,
  BarChart3,
  Settings,
  LogOut,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link, useLocation, useNavigate } from "@tanstack/react-router";
import { useAuth } from "@/hooks/use-auth";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { ModeToggle } from "@/components/theme/mode-toggle";

const navItems = [
  { key: "dashboard", icon: LayoutDashboard, href: "/" },
  { key: "accounts", icon: Wallet, href: "/accounts" },
  { key: "transactions", icon: ArrowLeftRight, href: "/transactions" },
  { key: "budgets", icon: PiggyBank, href: "/budgets" },
  { key: "reports", icon: BarChart3, href: "/reports" },
  { key: "settings", icon: Settings, href: "/settings" },
] as const;

export function AppSidebar() {
  const { t } = useTranslation();
  const { logout, user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <Sidebar>
      <SidebarHeader className="p-4">
        <h1 className="text-lg font-semibold">{t("app_name")}</h1>
        {user && (
          <p className="text-sm text-muted-foreground truncate">
            {user.email}
          </p>
        )}
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => (
                <SidebarMenuItem key={item.key}>
                  <SidebarMenuButton
                    isActive={location.pathname === item.href}
                    render={
                      <Link to={item.href as "/"}>
                        <item.icon className="h-4 w-4" />
                        <span>{t(item.key)}</span>
                      </Link>
                    }
                  />
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter className="p-2 flex flex-row items-center justify-between">
        <ModeToggle />
        <SidebarMenuButton onClick={() => { logout(); navigate({ to: "/login" }); }}>
          <LogOut className="h-4 w-4" />
          <span>{t("logout")}</span>
        </SidebarMenuButton>
      </SidebarFooter>
    </Sidebar>
  );
}
