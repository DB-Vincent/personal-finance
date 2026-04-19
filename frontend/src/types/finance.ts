export interface Category {
  id: string;
  user_id: string;
  group_name: string;
  name: string;
  is_income: boolean;
  is_archived: boolean;
  create_time: string;
}

export interface CategoryGroup {
  group_name: string;
  categories: Category[];
}

export interface Account {
  id: string;
  user_id: string;
  name: string;
  type: AccountType;
  starting_balance: string;
  balance: string;
  is_archived: boolean;
  created_by: string;
  create_time: string;
  updated_by: string;
  update_time: string;
}

export type AccountType =
  | "checking"
  | "savings"
  | "credit_card"
  | "cash"
  | "investment"
  | "loan"
  | "other";

export interface Tag {
  id: string;
  user_id: string;
  name: string;
  color: string;
  create_time: string;
}

export interface Transaction {
  id: string;
  user_id: string;
  account_id: string;
  type: "income" | "expense" | "transfer";
  amount: string;
  category_id?: string;
  transfer_account_id?: string;
  date: string;
  notes?: string;
  tags?: Tag[];
  created_by: string;
  create_time: string;
  updated_by: string;
  update_time: string;
}

export interface NetWorth {
  total: string;
}

export interface CreateAccountRequest {
  name: string;
  type: AccountType;
  starting_balance: number;
}

export interface UpdateAccountRequest {
  name?: string;
  type?: AccountType;
}

export interface CreateTransactionRequest {
  account_id: string;
  type: "income" | "expense" | "transfer";
  amount: number;
  category_id?: string;
  transfer_account_id?: string;
  date: string;
  notes?: string;
  tag_ids?: string[];
}

export interface UpdateTransactionRequest {
  account_id?: string;
  type?: "income" | "expense" | "transfer";
  amount?: number;
  category_id?: string;
  transfer_account_id?: string;
  date?: string;
  notes?: string;
  tag_ids?: string[];
}

export interface CreateCategoryRequest {
  group_name: string;
  name: string;
  is_income: boolean;
}

export interface CreateTagRequest {
  name: string;
  color?: string;
}

export interface UpdateTagRequest {
  name?: string;
  color?: string;
}

export interface TransactionFilters {
  page_size?: number;
  page_token?: string;
  account_id?: string;
  category_id?: string;
  tag_id?: string;
  type?: string;
  date_from?: string;
  date_to?: string;
  amount_min?: string;
  amount_max?: string;
  search?: string;
}
