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
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarSeparator,
} from "@/components/ui/sidebar";
import { ModeToggle } from "@/components/theme/mode-toggle";

const mainNav = [
  { key: "dashboard", icon: LayoutDashboard, href: "/" },
  { key: "accounts", icon: Wallet, href: "/accounts" },
  { key: "transactions", icon: ArrowLeftRight, href: "/transactions" },
] as const;

const planNav = [
  { key: "budgets", icon: PiggyBank, href: "/budgets" },
  { key: "reports", icon: BarChart3, href: "/reports" },
] as const;

export function AppSidebar() {
  const { t } = useTranslation();
  const { logout, user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const initials = user?.display_name
    ? user.display_name
        .split(" ")
        .map((n) => n[0])
        .join("")
        .toUpperCase()
        .slice(0, 2)
    : user?.email?.charAt(0).toUpperCase() || "?";

  return (
    <Sidebar>
      <SidebarHeader className="p-4 pb-2">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground text-xs font-medium">
            {initials}
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-sm font-semibold truncate">
              {user?.display_name || t("app_name")}
            </p>
            {user?.email && (
              <p className="text-xs text-muted-foreground truncate">
                {user.email}
              </p>
            )}
          </div>
        </div>
      </SidebarHeader>
      <SidebarSeparator />
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {mainNav.map((item) => (
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
        <SidebarGroup>
          <SidebarGroupLabel>Planning</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {planNav.map((item) => (
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
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton
                  isActive={location.pathname === "/settings"}
                  render={
                    <Link to="/settings">
                      <Settings className="h-4 w-4" />
                      <span>{t("settings")}</span>
                    </Link>
                  }
                />
              </SidebarMenuItem>
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
