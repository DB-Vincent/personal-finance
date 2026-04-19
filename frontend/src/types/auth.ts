export interface User {
  id: string;
  email: string;
  display_name: string;
  currency_symbol: string;
  role: string;
  is_disabled: boolean;
  create_time: string;
  update_time: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
}

export interface UpdateProfileRequest {
  display_name?: string;
  currency_symbol?: string;
}
