import { useTranslation } from "react-i18next";
import { ArrowDownLeft, ArrowUpRight, ArrowLeftRight } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import type { Transaction } from "@/types/finance";

const typeConfig = {
  income: { icon: ArrowDownLeft, color: "text-green-600", sign: "+" },
  expense: { icon: ArrowUpRight, color: "text-red-500", sign: "-" },
  transfer: { icon: ArrowLeftRight, color: "text-blue-500", sign: "" },
};

export function TransactionList({
  transactions,
  currencySymbol,
  onSelect,
}: {
  transactions: Transaction[];
  currencySymbol: string;
  onSelect?: (tx: Transaction) => void;
}) {
  const { t } = useTranslation();

  if (transactions.length === 0) {
    return (
      <p className="text-center py-8 text-muted-foreground">
        {t("no_transactions")}
      </p>
    );
  }

  function formatAmount(amount: string, type: string) {
    const cfg = typeConfig[type as keyof typeof typeConfig];
    const formatted = parseFloat(amount).toLocaleString(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
    return `${cfg.sign}${currencySymbol}${formatted}`;
  }

  return (
    <div className="divide-y">
      {transactions.map((tx) => {
        const cfg = typeConfig[tx.type];
        const Icon = cfg.icon;
        return (
          <button
            key={tx.id}
            className="flex items-center gap-3 py-3 px-2 w-full text-left hover:bg-muted/50 rounded transition-colors"
            onClick={() => onSelect?.(tx)}
            type="button"
          >
            <div className={`rounded-full p-2 bg-muted ${cfg.color}`}>
              <Icon className="h-4 w-4" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">
                {tx.notes || tx.type}
              </p>
              <p className="text-xs text-muted-foreground">
                {new Date(tx.date).toLocaleDateString()}
              </p>
            </div>
            <div className="flex items-center gap-2">
              {tx.tags?.map((tag) => (
                <Badge
                  key={tag.id}
                  variant="outline"
                  className="text-xs"
                  style={{ borderColor: tag.color, color: tag.color }}
                >
                  {tag.name}
                </Badge>
              ))}
              <span className={`text-sm font-semibold ${cfg.color}`}>
                {formatAmount(tx.amount, tx.type)}
              </span>
            </div>
          </button>
        );
      })}
    </div>
  );
}
