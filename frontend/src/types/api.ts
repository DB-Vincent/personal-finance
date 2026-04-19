export interface ApiError {
  error: {
    code: number;
    status: string;
    message: string;
    details?: { field?: string; reason: string }[];
  };
}

export interface ListResponse<T> {
  items: T[];
  next_page_token?: string;
  total_size: number;
}
