import { SidebarTrigger } from "@/components/ui/sidebar";
import { Separator } from "@/components/ui/separator";

export function AppHeader() {
  return (
    <header className="flex h-14 items-center gap-2 border-b px-4 md:hidden">
      <SidebarTrigger />
      <Separator orientation="vertical" className="h-6" />
    </header>
  );
}
