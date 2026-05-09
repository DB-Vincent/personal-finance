import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useTransactions, useAccounts, useCategories, useDeleteTransaction } from "@/services/finance";
import { useAuth } from "@/hooks/use-auth";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Search } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { TransactionList } from "./transaction-list";
import { EditTransactionDialog } from "./edit-transaction-dialog";
import type { Transaction, TransactionFilters } from "@/types/finance";
import { toast } from "sonner";

export function TransactionsPage() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const currencySymbol = user?.currency_symbol || "€";

  const [filters, setFilters] = useState<TransactionFilters>({ page_size: 50 });
  const [selected, setSelected] = useState<Transaction | null>(null);

  const { data, isLoading } = useTransactions(filters);
  const { data: accounts } = useAccounts();
  const { data: categoryGroups } = useCategories();
  const deleteTx = useDeleteTransaction();

  const allCategories = categoryGroups?.flatMap((g) => g.categories) || [];

  function handleDelete(id: string) {
    if (!confirm(t("confirm_delete_transaction"))) return;
    deleteTx.mutate(id, {
      onSuccess: () => toast.success(t("transaction_deleted")),
    });
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">{t("transactions")}</h1>

      <div className="flex flex-wrap gap-2">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={t("search")}
            className="w-48 pl-9"
          value={filters.search || ""}
            onChange={(e) => setFilters((f) => ({ ...f, search: e.target.value, page_token: undefined }))}
          />
        </div>
        <Select
          value={filters.account_id || "all"}
          onValueChange={(v) => {
            const accountId = v === "all" || v === null ? undefined : v;
            setFilters((f) => ({ ...f, account_id: accountId, page_token: undefined }));
          }}
        >
          <SelectTrigger className="w-40">
            <SelectValue placeholder={t("all_accounts")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t("all_accounts")}</SelectItem>
            {accounts?.map((a) => (
              <SelectItem key={a.id} value={a.id}>
                {a.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select
          value={filters.type || "all"}
          onValueChange={(v) => {
            const txType = v === "all" || v === null ? undefined : v;
            setFilters((f) => ({ ...f, type: txType, page_token: undefined }));
          }}
        >
          <SelectTrigger className="w-32">
            <SelectValue placeholder={t("all_types")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t("all_types")}</SelectItem>
            <SelectItem value="income">{t("income")}</SelectItem>
            <SelectItem value="expense">{t("expense")}</SelectItem>
            <SelectItem value="transfer">{t("transfer")}</SelectItem>
          </SelectContent>
        </Select>
        <Input
          type="date"
          className="w-40"
          value={filters.date_from || ""}
          onChange={(e) => setFilters((f) => ({ ...f, date_from: e.target.value || undefined, page_token: undefined }))}
        />
        <Input
          type="date"
          className="w-40"
          value={filters.date_to || ""}
          onChange={(e) => setFilters((f) => ({ ...f, date_to: e.target.value || undefined, page_token: undefined }))}
        />
      </div>

      {isLoading ? (
        <Skeleton className="h-48" />
      ) : (
        <>
          <Card className="py-0">
            <CardContent className="p-0">
              <TransactionList
                transactions={data?.items || []}
                currencySymbol={currencySymbol}
                onSelect={(tx) => setSelected(tx)}
              />
            </CardContent>
          </Card>
          {data?.total_size !== undefined && (
            <p className="text-sm text-muted-foreground mt-3">
              {(data.items || []).length} of {data.total_size} {t("transactions").toLowerCase()}
            </p>
          )}
          {data?.next_page_token && (
            <Button
              variant="outline"
              className="mt-3"
              onClick={() => setFilters((f) => ({ ...f, page_token: data.next_page_token }))}
            >
              {t("load_more")}
            </Button>
          )}
        </>
      )}

      {selected && (
        <EditTransactionDialog
          transaction={selected}
          accounts={accounts || []}
          categories={allCategories}
          onClose={() => setSelected(null)}
          onDelete={() => {
            handleDelete(selected.id);
            setSelected(null);
          }}
        />
      )}
    </div>
  );
}
