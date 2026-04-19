import { createFileRoute } from "@tanstack/react-router";
import { AccountDetailPage } from "@/components/accounts/account-detail-page";

export const Route = createFileRoute("/_authenticated/accounts/$accountId")({
  component: AccountDetailPage,
});
