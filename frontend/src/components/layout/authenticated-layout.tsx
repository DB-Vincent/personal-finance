import type { ReactNode } from "react";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { AppSidebar } from "./app-sidebar";
import { AppHeader } from "./app-header";
import { BottomTabs } from "./bottom-tabs";
import { QuickAdd } from "@/components/transactions/quick-add";

export function AuthenticatedLayout({ children }: { children: ReactNode }) {
  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <AppHeader />
        <main className="flex-1 p-4 pb-20 md:p-6 md:pb-6">{children}</main>
        <BottomTabs />
        <QuickAdd />
      </SidebarInset>
    </SidebarProvider>
  );
}
