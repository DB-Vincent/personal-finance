import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { Link } from "@tanstack/react-router";
import { useMemo } from "react";
import { useAuth } from "@/hooks/use-auth";
import { useNetWorth, useAccounts, useTransactions } from "@/services/finance";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { TransactionList } from "@/components/transactions/transaction-list";
import {
  TrendingUp,
  TrendingDown,
  Wallet,
  ArrowRight,
  Landmark,
} from "lucide-react";

export const Route = createFileRoute("/_authenticated/")({
  component: DashboardPage,
});

function DashboardPage() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const currencySymbol = user?.currency_symbol || "€";

  const { data: netWorth, isLoading: nwLoading } = useNetWorth();
  const { data: accounts, isLoading: accLoading } = useAccounts();

  const now = new Date();
  const monthStart = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-01`;
  const monthEnd = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(new Date(now.getFullYear(), now.getMonth() + 1, 0).getDate()).padStart(2, "0")}`;

  const { data: monthTx } = useTransactions({
    date_from: monthStart,
    date_to: monthEnd,
    page_size: 200,
  });

  const { data: recentTx, isLoading: txLoading } = useTransactions({
    page_size: 5,
  });

  const { monthlyIncome, monthlyExpenses } = useMemo(() => {
    if (!monthTx?.items) return { monthlyIncome: 0, monthlyExpenses: 0 };
    let income = 0;
    let expenses = 0;
    for (const tx of monthTx.items) {
      const amount = parseFloat(tx.amount);
      if (tx.type === "income") income += amount;
      else if (tx.type === "expense") expenses += amount;
    }
    return { monthlyIncome: income, monthlyExpenses: expenses };
  }, [monthTx]);

  function formatAmount(amount: number | string) {
    const num = typeof amount === "string" ? parseFloat(amount) : amount;
    return `${currencySymbol}${num.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  }

  const greeting = user?.display_name
    ? `${t("welcome_back")}, ${user.display_name}`
    : t("welcome_back");

  const hasData = accounts && accounts.length > 0;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">{greeting}</h1>
        <p className="text-sm text-muted-foreground">{t("this_month")}</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              {t("net_worth")}
            </CardTitle>
            <Landmark className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {nwLoading ? (
              <Skeleton className="h-7 w-32" />
            ) : (
              <p className="text-2xl font-bold">
                {formatAmount(netWorth?.total || "0")}
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              {t("monthly_income")}
            </CardTitle>
            <TrendingUp className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            {!monthTx ? (
              <Skeleton className="h-7 w-28" />
            ) : (
              <p className="text-2xl font-bold text-green-600">
                {formatAmount(monthlyIncome)}
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              {t("monthly_expenses")}
            </CardTitle>
            <TrendingDown className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            {!monthTx ? (
              <Skeleton className="h-7 w-28" />
            ) : (
              <p className="text-2xl font-bold text-red-500">
                {formatAmount(monthlyExpenses)}
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              {t("total_accounts")}
            </CardTitle>
            <Wallet className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {accLoading ? (
              <Skeleton className="h-7 w-12" />
            ) : (
              <p className="text-2xl font-bold">{accounts?.length || 0}</p>
            )}
          </CardContent>
        </Card>
      </div>

      {hasData ? (
        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold">{t("recent_transactions")}</h2>
            <Link
              to="/transactions"
              className="text-sm text-muted-foreground hover:text-foreground transition-colors flex items-center gap-1"
            >
              {t("view_all")}
              <ArrowRight className="h-3.5 w-3.5" />
            </Link>
          </div>
          {txLoading ? (
            <Skeleton className="h-48" />
          ) : recentTx?.items && recentTx.items.length > 0 ? (
            <Card>
              <CardContent className="p-0 py-1">
                <TransactionList
                  transactions={recentTx.items}
                  currencySymbol={currencySymbol}
                />
              </CardContent>
            </Card>
          ) : (
            <p className="text-muted-foreground text-sm py-8 text-center">
              {t("no_transactions")}
            </p>
          )}
        </div>
      ) : !accLoading ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <Wallet className="h-10 w-10 text-muted-foreground mb-3" />
            <h3 className="font-semibold mb-1">{t("no_data_yet")}</h3>
            <p className="text-sm text-muted-foreground max-w-sm">
              {t("get_started")}
            </p>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}
