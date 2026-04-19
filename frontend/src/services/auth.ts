import { apiClient, setAccessToken } from "@/lib/api-client";
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  TokenResponse,
  User,
  UpdateProfileRequest,
} from "@/types/auth";

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const res = await apiClient<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(data),
  });
  setAccessToken(res.access_token);
  localStorage.setItem("refresh_token", res.refresh_token);
  return res;
}

export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const res = await apiClient<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify(data),
  });
  setAccessToken(res.access_token);
  localStorage.setItem("refresh_token", res.refresh_token);
  return res;
}

export async function refreshTokens(): Promise<TokenResponse> {
  const refreshToken = localStorage.getItem("refresh_token");
  if (!refreshToken) throw new Error("No refresh token");

  const res = await apiClient<TokenResponse>("/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
  setAccessToken(res.access_token);
  localStorage.setItem("refresh_token", res.refresh_token);
  return res;
}

export async function getProfile(): Promise<User> {
  return apiClient<User>("/users/me");
}

export async function updateProfile(
  data: UpdateProfileRequest
): Promise<User> {
  return apiClient<User>("/users/me", {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export function logout() {
  setAccessToken(null);
  localStorage.removeItem("refresh_token");
}
