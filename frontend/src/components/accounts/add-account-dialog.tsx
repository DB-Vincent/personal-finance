import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useTranslation } from "react-i18next";
import { useCreateAccount } from "@/services/finance";
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

const schema = z.object({
  name: z.string().min(1).max(255),
  type: z.enum(["checking", "savings", "credit_card", "cash", "investment", "loan", "other"]),
  starting_balance: z.number(),
});

type FormData = z.infer<typeof schema>;

const accountTypes = [
  { value: "checking", label: "Checking" },
  { value: "savings", label: "Savings" },
  { value: "credit_card", label: "Credit Card" },
  { value: "cash", label: "Cash" },
  { value: "investment", label: "Investment" },
  { value: "loan", label: "Loan" },
  { value: "other", label: "Other" },
];

export function AddAccountDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { t } = useTranslation();
  const createAccount = useCreateAccount();

  const form = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", type: "checking", starting_balance: 0 },
  });

  function onSubmit(data: FormData) {
    createAccount.mutate(data, {
      onSuccess: () => {
        toast.success(t("account_created"));
        form.reset();
        onOpenChange(false);
      },
      onError: () => toast.error(t("error")),
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("add_account")}</DialogTitle>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">{t("account_name")}</Label>
            <Input id="name" {...form.register("name")} />
            {form.formState.errors.name && (
              <p className="text-sm text-destructive">{form.formState.errors.name.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label>{t("account_type")}</Label>
            <Select
              value={form.watch("type")}
              onValueChange={(v) => { if (v) form.setValue("type", v as FormData["type"]); }}
            >
              <SelectTrigger>
                <span className="flex flex-1 text-left truncate">
                  {accountTypes.find((at) => at.value === form.watch("type"))?.label}
                </span>
              </SelectTrigger>
              <SelectContent>
                {accountTypes.map((t) => (
                  <SelectItem key={t.value} value={t.value}>
                    {t.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="balance">{t("starting_balance")}</Label>
            <Input
              id="balance"
              type="number"
              step="0.01"
              {...form.register("starting_balance")}
            />
          </div>

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t("cancel")}
            </Button>
            <Button type="submit" disabled={createAccount.isPending}>
              {t("save")}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
