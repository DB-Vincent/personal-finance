import { useState } from "react";
import { Link } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { Plus, Archive, Trash2, Wallet } from "lucide-react";
import { useAccounts, useNetWorth, useArchiveAccount, useDeleteAccount } from "@/services/finance";
import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { AddAccountDialog } from "./add-account-dialog";
import type { Account } from "@/types/finance";
import { toast } from "sonner";

const typeLabels: Record<string, string> = {
  checking: "Checking",
  savings: "Savings",
  credit_card: "Credit Card",
  cash: "Cash",
  investment: "Investment",
  loan: "Loan",
  other: "Other",
};

export function AccountsPage() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const { data: accounts, isLoading } = useAccounts();
  const { data: netWorth } = useNetWorth();
  const archiveMutation = useArchiveAccount();
  const deleteMutation = useDeleteAccount();
  const [addOpen, setAddOpen] = useState(false);

  const currencySymbol = user?.currency_symbol || "€";

  function formatAmount(amount: string) {
    return `${currencySymbol}${parseFloat(amount).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  }

  function handleArchive(id: string) {
    archiveMutation.mutate(id, {
      onSuccess: () => toast.success(t("account_archived")),
    });
  }

  function handleDelete(account: Account) {
    if (!confirm(t("confirm_delete_account", { name: account.name }))) return;
    deleteMutation.mutate(account.id, {
      onSuccess: () => toast.success(t("account_deleted")),
      onError: () => toast.error(t("account_has_transactions")),
    });
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t("accounts")}</h1>
        <Button onClick={() => setAddOpen(true)} size="sm">
          <Plus className="h-4 w-4 mr-1" />
          {t("add_account")}
        </Button>
      </div>

      {netWorth && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              {t("net_worth")}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{formatAmount(netWorth.total)}</p>
          </CardContent>
        </Card>
      )}

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-32" />
          ))}
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {accounts?.map((account) => (
            <Card key={account.id} className="group relative">
              <Link
                to="/accounts/$accountId"
                params={{ accountId: account.id }}
                className="block"
              >
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Wallet className="h-4 w-4 text-muted-foreground" />
                      <CardTitle className="text-base">{account.name}</CardTitle>
                    </div>
                    <Badge variant="secondary" className="group-hover:opacity-0 transition-opacity">{typeLabels[account.type]}</Badge>
                  </div>
                </CardHeader>
                <CardContent>
                  <p
                    className={`text-2xl font-semibold ${parseFloat(account.balance) < 0 ? "text-red-500" : ""}`}
                  >
                    {formatAmount(account.balance)}
                  </p>
                </CardContent>
              </Link>
              <div className="absolute top-3 right-3 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7"
                  onClick={(e) => { e.preventDefault(); handleArchive(account.id); }}
                >
                  <Archive className="h-3.5 w-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 text-destructive"
                  onClick={(e) => { e.preventDefault(); handleDelete(account); }}
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              </div>
            </Card>
          ))}

          {accounts?.length === 0 && (
            <div className="col-span-full text-center py-12 text-muted-foreground">
              <Wallet className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>{t("no_accounts")}</p>
            </div>
          )}
        </div>
      )}

      <AddAccountDialog open={addOpen} onOpenChange={setAddOpen} />
    </div>
  );
}
