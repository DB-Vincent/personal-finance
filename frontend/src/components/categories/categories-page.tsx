import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Plus, Archive } from "lucide-react";
import { useCategories, useCreateCategory, useArchiveCategory } from "@/services/finance";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";

export function CategoriesPage() {
  const { t } = useTranslation();
  const [showArchived, setShowArchived] = useState(false);
  const [addOpen, setAddOpen] = useState(false);
  const { data: groups, isLoading } = useCategories(showArchived);
  const archiveMutation = useArchiveCategory();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t("categories")}</h1>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowArchived(!showArchived)}
          >
            {showArchived ? t("hide_archived") : t("show_archived")}
          </Button>
          <Button size="sm" onClick={() => setAddOpen(true)}>
            <Plus className="h-4 w-4 mr-1" />
            {t("add_category")}
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-24" />
          ))}
        </div>
      ) : (
        <div className="space-y-4">
          {groups?.map((group) => (
            <Card key={group.group_name}>
              <CardHeader className="pb-2">
                <CardTitle className="text-base">{group.group_name}</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {group.categories.map((cat) => (
                    <div key={cat.id} className="flex items-center gap-1">
                      <Badge
                        variant={cat.is_archived ? "outline" : cat.is_income ? "default" : "secondary"}
                        className={cat.is_archived ? "opacity-50" : ""}
                      >
                        {cat.name}
                      </Badge>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-5 w-5"
                        onClick={() =>
                          archiveMutation.mutate(cat.id, {
                            onSuccess: () =>
                              toast.success(
                                cat.is_archived ? t("category_unarchived") : t("category_archived")
                              ),
                          })
                        }
                      >
                        <Archive className="h-3 w-3" />
                      </Button>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <AddCategoryDialog open={addOpen} onOpenChange={setAddOpen} />
    </div>
  );
}

function AddCategoryDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { t } = useTranslation();
  const createCategory = useCreateCategory();
  const [name, setName] = useState("");
  const [groupName, setGroupName] = useState("");
  const [isIncome, setIsIncome] = useState(false);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    createCategory.mutate(
      { name, group_name: groupName, is_income: isIncome },
      {
        onSuccess: () => {
          toast.success(t("category_created"));
          setName("");
          setGroupName("");
          onOpenChange(false);
        },
        onError: () => toast.error(t("error")),
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("add_category")}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>{t("category_name")}</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} required />
          </div>
          <div className="space-y-2">
            <Label>{t("group")}</Label>
            <Input value={groupName} onChange={(e) => setGroupName(e.target.value)} required />
          </div>
          <div className="space-y-2">
            <Label>{t("type")}</Label>
            <Select value={isIncome ? "income" : "expense"} onValueChange={(v) => setIsIncome(v === "income")}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="expense">{t("expense")}</SelectItem>
                <SelectItem value="income">{t("income")}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t("cancel")}
            </Button>
            <Button type="submit" disabled={createCategory.isPending}>
              {t("save")}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
