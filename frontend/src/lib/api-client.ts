import type { ApiError } from "@/types/api";

const API_BASE = "/api/v1";

let accessToken: string | null = null;
let refreshPromise: Promise<string | null> | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken() {
  return accessToken;
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = localStorage.getItem("refresh_token");
  if (!refreshToken) return null;

  try {
    const res = await fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!res.ok) {
      localStorage.removeItem("refresh_token");
      return null;
    }

    const data = await res.json();
    accessToken = data.access_token;
    localStorage.setItem("refresh_token", data.refresh_token);
    return data.access_token;
  } catch {
    localStorage.removeItem("refresh_token");
    return null;
  }
}

export async function apiClient<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");

  if (accessToken) {
    headers.set("Authorization", `Bearer ${accessToken}`);
  }

  let res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (res.status === 401 && accessToken) {
    if (!refreshPromise) {
      refreshPromise = refreshAccessToken();
    }
    const newToken = await refreshPromise;
    refreshPromise = null;

    if (newToken) {
      headers.set("Authorization", `Bearer ${newToken}`);
      res = await fetch(`${API_BASE}${path}`, { ...options, headers });
    }
  }

  if (!res.ok) {
    const errorBody: ApiError = await res.json().catch(() => ({
      error: {
        code: res.status,
        status: "UNKNOWN",
        message: res.statusText,
      },
    }));
    throw errorBody;
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}
