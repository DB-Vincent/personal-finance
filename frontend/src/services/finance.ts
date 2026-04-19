import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";
import type { ListResponse } from "@/types/api";
import type {
  Account,
  CategoryGroup,
  Tag,
  Transaction,
  NetWorth,
  CreateAccountRequest,
  UpdateAccountRequest,
  CreateTransactionRequest,
  UpdateTransactionRequest,
  CreateCategoryRequest,
  CreateTagRequest,
  UpdateTagRequest,
  TransactionFilters,
} from "@/types/finance";

// --- Categories ---

export function useCategories(includeArchived = false) {
  return useQuery({
    queryKey: ["categories", { includeArchived }],
    queryFn: () => {
      const params = includeArchived ? "?include_archived=true" : "";
      return apiClient<CategoryGroup[]>(`/finance/categories${params}`);
    },
  });
}

export function useCreateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateCategoryRequest) =>
      apiClient("/finance/categories", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["categories"] }),
  });
}

export function useArchiveCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient(`/finance/categories/${id}/archive`, { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["categories"] }),
  });
}

// --- Accounts ---

export function useAccounts(includeArchived = false) {
  return useQuery({
    queryKey: ["accounts", { includeArchived }],
    queryFn: () => {
      const params = includeArchived ? "?include_archived=true" : "";
      return apiClient<Account[]>(`/finance/accounts${params}`);
    },
  });
}

export function useAccount(id: string) {
  return useQuery({
    queryKey: ["accounts", id],
    queryFn: () => apiClient<Account>(`/finance/accounts/${id}`),
    enabled: !!id,
  });
}

export function useNetWorth() {
  return useQuery({
    queryKey: ["net-worth"],
    queryFn: () => apiClient<NetWorth>("/finance/accounts/net-worth"),
  });
}

export function useCreateAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateAccountRequest) =>
      apiClient<Account>("/finance/accounts", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["accounts"] });
      qc.invalidateQueries({ queryKey: ["net-worth"] });
    },
  });
}

export function useUpdateAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateAccountRequest }) =>
      apiClient<Account>(`/finance/accounts/${id}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["accounts"] });
    },
  });
}

export function useArchiveAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient(`/finance/accounts/${id}/archive`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["accounts"] });
      qc.invalidateQueries({ queryKey: ["net-worth"] });
    },
  });
}

export function useDeleteAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient(`/finance/accounts/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["accounts"] });
      qc.invalidateQueries({ queryKey: ["net-worth"] });
    },
  });
}

// --- Tags ---

export function useTags() {
  return useQuery({
    queryKey: ["tags"],
    queryFn: () => apiClient<Tag[]>("/finance/tags"),
  });
}

export function useCreateTag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateTagRequest) =>
      apiClient<Tag>("/finance/tags", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tags"] }),
  });
}

export function useUpdateTag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateTagRequest }) =>
      apiClient<Tag>(`/finance/tags/${id}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tags"] }),
  });
}

export function useDeleteTag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient(`/finance/tags/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tags"] }),
  });
}

// --- Transactions ---

export function useTransactions(filters: TransactionFilters = {}) {
  return useQuery({
    queryKey: ["transactions", filters],
    queryFn: () => {
      const params = new URLSearchParams();
      for (const [k, v] of Object.entries(filters)) {
        if (v !== undefined && v !== "") params.set(k, String(v));
      }
      const qs = params.toString();
      return apiClient<ListResponse<Transaction>>(
        `/finance/transactions${qs ? `?${qs}` : ""}`
      );
    },
  });
}

export function useTransaction(id: string) {
  return useQuery({
    queryKey: ["transactions", id],
    queryFn: () => apiClient<Transaction>(`/finance/transactions/${id}`),
    enabled: !!id,
  });
}

export function useCreateTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateTransactionRequest) =>
      apiClient<Transaction>("/finance/transactions", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions"] });
      qc.invalidateQueries({ queryKey: ["accounts"] });
      qc.invalidateQueries({ queryKey: ["net-worth"] });
    },
  });
}

export function useUpdateTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: UpdateTransactionRequest;
    }) =>
      apiClient<Transaction>(`/finance/transactions/${id}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions"] });
      qc.invalidateQueries({ queryKey: ["accounts"] });
      qc.invalidateQueries({ queryKey: ["net-worth"] });
    },
  });
}

export function useDeleteTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient(`/finance/transactions/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions"] });
      qc.invalidateQueries({ queryKey: ["accounts"] });
      qc.invalidateQueries({ queryKey: ["net-worth"] });
    },
  });
}
