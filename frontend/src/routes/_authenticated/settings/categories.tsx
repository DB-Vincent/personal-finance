import { createFileRoute } from "@tanstack/react-router";
import { CategoriesPage } from "@/components/categories/categories-page";

export const Route = createFileRoute("/_authenticated/settings/categories")({
  component: CategoriesPage,
});
