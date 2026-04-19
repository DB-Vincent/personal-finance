import { useParams, Link } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { ArrowLeft } from "lucide-react";
import { useAccount, useTransactions } from "@/services/finance";
import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { TransactionList } from "@/components/transactions/transaction-list";

export function AccountDetailPage() {
  const { accountId } = useParams({ from: "/_authenticated/accounts/$accountId" });
  const { t } = useTranslation();
  const { user } = useAuth();
  const { data: account, isLoading: accountLoading } = useAccount(accountId);
  const { data: transactions, isLoading: txLoading } = useTransactions({
    account_id: accountId,
    page_size: 50,
  });

  const currencySymbol = user?.currency_symbol || "€";

  function formatAmount(amount: string) {
    return `${currencySymbol}${parseFloat(amount).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  }

  if (accountLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-32" />
      </div>
    );
  }

  if (!account) return <p>{t("account_not_found")}</p>;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link to="/accounts">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <div>
          <h1 className="text-2xl font-bold">{account.name}</h1>
          <Badge variant="secondary" className="mt-1">{account.type}</Badge>
        </div>
      </div>

      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">
            {t("balance")}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p
            className={`text-3xl font-bold ${parseFloat(account.balance) < 0 ? "text-red-500" : ""}`}
          >
            {formatAmount(account.balance)}
          </p>
        </CardContent>
      </Card>

      <div>
        <h2 className="text-lg font-semibold mb-4">{t("transactions")}</h2>
        {txLoading ? (
          <Skeleton className="h-48" />
        ) : (
          <TransactionList
            transactions={transactions?.items || []}
            currencySymbol={currencySymbol}
          />
        )}
      </div>
    </div>
  );
}
