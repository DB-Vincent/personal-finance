import { createFileRoute, Link } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { FolderTree } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

export const Route = createFileRoute("/_authenticated/settings")({
  component: SettingsPage,
});

function SettingsPage() {
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">{t("settings")}</h1>
      <div className="grid gap-4 sm:grid-cols-2">
        <Link to="/settings/categories">
          <Card className="hover:bg-muted/50 transition-colors cursor-pointer">
            <CardContent className="flex items-center gap-3 pt-6">
              <FolderTree className="h-5 w-5 text-muted-foreground" />
              <span className="font-medium">{t("categories")}</span>
            </CardContent>
          </Card>
        </Link>
      </div>
    </div>
  );
}
