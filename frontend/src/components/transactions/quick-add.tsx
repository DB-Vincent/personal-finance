import { useState, useRef, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Plus } from "lucide-react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useCreateTransaction, useAccounts, useCategories } from "@/services/finance";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectGroup,
  SelectLabel,
} from "@/components/ui/select";
import { toast } from "sonner";

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

export function QuickAdd() {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const amountRef = useRef<HTMLInputElement | null>(null);
  const { data: accounts } = useAccounts();
  const { data: categoryGroups } = useCategories();
  const createTx = useCreateTransaction();

  const today = new Date().toISOString().split("T")[0];
  const defaultAccount = localStorage.getItem("last_used_account") || "";

  const form = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      amount: undefined,
      type: "expense",
      account_id: defaultAccount,
      category_id: "",
      date: today,
      notes: "",
    },
  });

  const txType = form.watch("type");
  const watchedCategoryId = form.watch("category_id");
  const watchedAccountId = form.watch("account_id");
  const watchedTransferAccountId = form.watch("transfer_account_id");

  const allCategories = categoryGroups?.flatMap((g) => g.categories);
  const selectedCategoryName = allCategories?.find((c) => c.id === watchedCategoryId)?.name;
  const selectedAccountName = accounts?.find((a) => a.id === watchedAccountId)?.name;
  const selectedTransferAccountName = accounts?.find((a) => a.id === watchedTransferAccountId)?.name;

  useEffect(() => {
    if (open) {
      setTimeout(() => amountRef.current?.focus(), 100);
    }
  }, [open]);

  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if (e.key === "n" && !e.metaKey && !e.ctrlKey && !e.altKey) {
        const tag = (e.target as HTMLElement).tagName;
        if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
        e.preventDefault();
        setOpen(true);
      }
    }
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, []);

  function onSubmit(data: FormData) {
    const payload = {
      account_id: data.account_id,
      type: data.type,
      amount: data.amount,
      category_id: data.category_id || undefined,
      transfer_account_id: data.transfer_account_id || undefined,
      date: data.date,
      notes: data.notes || undefined,
    };

    createTx.mutate(payload, {
      onSuccess: () => {
        toast.success(t("transaction_added"));
        localStorage.setItem("last_used_account", data.account_id);
        form.reset({
          amount: undefined,
          type: "expense",
          account_id: data.account_id,
          category_id: "",
          date: today,
          notes: "",
        });
        setOpen(false);
      },
      onError: () => toast.error(t("error")),
    });
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger
        render={
          <Button
            size="icon"
            className="fixed bottom-20 right-4 z-50 h-14 w-14 rounded-full shadow-lg md:bottom-6 md:h-12 md:w-12"
          >
            <Plus className="h-6 w-6" />
          </Button>
        }
      />
      <SheetContent side="bottom" className="max-h-[85vh] overflow-y-auto rounded-t-xl sm:max-w-lg sm:mx-auto">
        <SheetHeader>
          <SheetTitle>{t("add_transaction")}</SheetTitle>
        </SheetHeader>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 px-4 pb-4">
          <div className="space-y-2">
            <Label>{t("amount")}</Label>
            <Input
              type="number"
              step="0.01"
              min="0"
              inputMode="decimal"
              className="text-2xl h-14 font-semibold"
              {...form.register("amount", { valueAsNumber: true })}
              ref={(e) => {
                form.register("amount").ref(e);
                amountRef.current = e;
              }}
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
                  {categoryGroups
                    ?.filter((g) =>
                      txType === "income"
                        ? g.categories.some((c) => c.is_income)
                        : g.categories.some((c) => !c.is_income)
                    )
                    .map((group) => (
                      <SelectGroup key={group.group_name}>
                        <SelectLabel>{group.group_name}</SelectLabel>
                        {group.categories
                          .filter((c) =>
                            txType === "income" ? c.is_income : !c.is_income
                          )
                          .map((cat) => (
                            <SelectItem key={cat.id} value={cat.id}>
                              {cat.name}
                            </SelectItem>
                          ))}
                      </SelectGroup>
                    ))}
                </SelectContent>
              </Select>
            </div>
          )}

          <div className="space-y-2">
            <Label>{txType === "transfer" ? t("from_account") : t("account")}</Label>
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
                {accounts?.map((a) => (
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
                    ?.filter((a) => a.id !== watchedAccountId)
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

          <details>
            <summary className="text-sm text-muted-foreground cursor-pointer">
              {t("more_details")}
            </summary>
            <div className="mt-3 space-y-2">
              <Label>{t("notes")}</Label>
              <Input {...form.register("notes")} placeholder={t("notes_placeholder")} />
            </div>
          </details>

          <Button type="submit" className="w-full" disabled={createTx.isPending}>
            {t("add_transaction")}
          </Button>
        </form>
      </SheetContent>
    </Sheet>
  );
}
