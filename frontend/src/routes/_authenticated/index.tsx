import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { useAuth } from "@/hooks/use-auth";

export const Route = createFileRoute("/_authenticated/")({
  component: DashboardPage,
});

function DashboardPage() {
  const { t } = useTranslation();
  const { user } = useAuth();

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">{t("dashboard")}</h1>
      <p className="text-muted-foreground">
        Welcome back{user?.display_name ? `, ${user.display_name}` : ""}!
      </p>
    </div>
  );
}
