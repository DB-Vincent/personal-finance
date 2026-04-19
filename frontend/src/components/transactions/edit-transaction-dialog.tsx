import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useTranslation } from "react-i18next";
import { Trash2 } from "lucide-react";
import { useUpdateTransaction } from "@/services/finance";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { toast } from "sonner";
import type { Transaction, Account, Category } from "@/types/finance";

const schema = z.object({
  amount: z.number().positive(),
  type: z.enum(["income", "expense", "transfer"]),
  account_id: z.string().min(1),
  category_id: z.string().optional(),
  transfer_account_id: z.string().optional(),
  date: z.string().min(1),
  notes: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

export function EditTransactionDialog({
  transaction,
  accounts,
  categories,
  onClose,
  onDelete,
}: {
  transaction: Transaction;
  accounts: Account[];
  categories: Category[];
  onClose: () => void;
  onDelete: () => void;
}) {
  const { t } = useTranslation();
  const updateTx = useUpdateTransaction();

  const form = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      amount: parseFloat(transaction.amount),
      type: transaction.type,
      account_id: transaction.account_id,
      category_id: transaction.category_id || "",
      transfer_account_id: transaction.transfer_account_id || "",
      date: transaction.date.split("T")[0],
      notes: transaction.notes || "",
    },
  });

  const txType = form.watch("type");
  const watchedCategoryId = form.watch("category_id");
  const watchedAccountId = form.watch("account_id");
  const watchedTransferAccountId = form.watch("transfer_account_id");

  const selectedCategoryName = categories.find((c) => c.id === watchedCategoryId)?.name;
  const selectedAccountName = accounts.find((a) => a.id === watchedAccountId)?.name;
  const selectedTransferAccountName = accounts.find((a) => a.id === watchedTransferAccountId)?.name;

  function onSubmit(data: FormData) {
    updateTx.mutate(
      {
        id: transaction.id,
        data: {
          amount: data.amount,
          type: data.type,
          account_id: data.account_id,
          category_id: data.category_id || undefined,
          transfer_account_id: data.transfer_account_id || undefined,
          date: data.date,
          notes: data.notes || undefined,
        },
      },
      {
        onSuccess: () => {
          toast.success(t("transaction_updated"));
          onClose();
        },
        onError: () => toast.error(t("error")),
      }
    );
  }

  return (
    <Dialog open onOpenChange={() => onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("edit_transaction")}</DialogTitle>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label>{t("amount")}</Label>
            <Input
              type="number"
              step="0.01"
              min="0"
              {...form.register("amount", { valueAsNumber: true })}
            />
          </div>

          <div className="flex gap-1">
            {(["expense", "income", "transfer"] as const).map((type) => (
              <Button
                key={type}
                type="button"
                variant={txType === type ? "default" : "outline"}
                size="sm"
                className="flex-1"
                onClick={() => form.setValue("type", type)}
              >
                {t(type)}
              </Button>
            ))}
          </div>

          {txType !== "transfer" && (
            <div className="space-y-2">
              <Label>{t("category")}</Label>
              <Select
                value={watchedCategoryId ?? ""}
                onValueChange={(v) => form.setValue("category_id", v ?? "")}
              >
                <SelectTrigger>
                  <span className="flex flex-1 text-left truncate">
                    {selectedCategoryName || <span className="text-muted-foreground">{t("select_category")}</span>}
                  </span>
                </SelectTrigger>
                <SelectContent>
                  {categories
                    .filter((c) => (txType === "income" ? c.is_income : !c.is_income))
                    .map((cat) => (
                      <SelectItem key={cat.id} value={cat.id}>
                        {cat.name}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>
          )}

          <div className="space-y-2">
            <Label>{t("account")}</Label>
            <Select
              value={watchedAccountId}
              onValueChange={(v) => form.setValue("account_id", v ?? "")}
            >
              <SelectTrigger>
                <span className="flex flex-1 text-left truncate">
                  {selectedAccountName || <span className="text-muted-foreground">{t("select_account")}</span>}
                </span>
              </SelectTrigger>
              <SelectContent>
                {accounts.map((a) => (
                  <SelectItem key={a.id} value={a.id}>
                    {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {txType === "transfer" && (
            <div className="space-y-2">
              <Label>{t("to_account")}</Label>
              <Select
                value={watchedTransferAccountId ?? ""}
                onValueChange={(v) => form.setValue("transfer_account_id", v ?? "")}
              >
                <SelectTrigger>
                  <span className="flex flex-1 text-left truncate">
                    {selectedTransferAccountName || <span className="text-muted-foreground">{t("select_account")}</span>}
                  </span>
                </SelectTrigger>
                <SelectContent>
                  {accounts
                    .filter((a) => a.id !== watchedAccountId)
                    .map((a) => (
                      <SelectItem key={a.id} value={a.id}>
                        {a.name}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>
          )}

          <div className="space-y-2">
            <Label>{t("date")}</Label>
            <Input type="date" {...form.register("date")} />
          </div>

          <div className="space-y-2">
            <Label>{t("notes")}</Label>
            <Input {...form.register("notes")} />
          </div>

          <div className="flex justify-between">
            <Button type="button" variant="destructive" size="sm" onClick={onDelete}>
              <Trash2 className="h-4 w-4 mr-1" />
              {t("delete")}
            </Button>
            <div className="flex gap-2">
              <Button type="button" variant="outline" onClick={onClose}>
                {t("cancel")}
              </Button>
              <Button type="submit" disabled={updateTx.isPending}>
                {t("save")}
              </Button>
            </div>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
