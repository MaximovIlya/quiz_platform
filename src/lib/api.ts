const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export function getAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("pulse_access_token");
}

export function getRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("pulse_refresh_token");
}

export function saveTokens(accessToken: string, refreshToken: string) {
  localStorage.setItem("pulse_access_token", accessToken);
  localStorage.setItem("pulse_refresh_token", refreshToken);
  // Дублируем access token в cookie для middleware
  document.cookie = `pulse_token=${accessToken}; path=/; max-age=900; SameSite=Strict`;
}

export function clearTokens() {
  localStorage.removeItem("pulse_access_token");
  localStorage.removeItem("pulse_refresh_token");
  localStorage.removeItem("pulse_user");
  document.cookie = "pulse_token=; path=/; max-age=0";
}

async function tryRefresh(): Promise<string | null> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return null;

  const res = await fetch(`${API_BASE}/api/v1/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refreshToken }),
  });

  if (!res.ok) {
    clearTokens();
    return null;
  }

  const { accessToken, refreshToken: newRefresh } = await res.json();
  saveTokens(accessToken, newRefresh);
  return accessToken;
}

export async function apiFetch(path: string, init?: RequestInit): Promise<Response> {
  const token = getAccessToken();

  const isFormData = init?.body instanceof FormData;
  const headers: Record<string, string> = {
    ...(!isFormData ? { "Content-Type": "application/json" } : {}),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(init?.headers as Record<string, string> | undefined),
  };

  const res = await fetch(`${API_BASE}/api/v1${path}`, { ...init, headers });

  if (res.status === 401) {
    const newToken = await tryRefresh();
    if (!newToken) {
      window.location.href = "/login";
      return res;
    }
    // Повтор с новым токеном
    return apiFetch(path, init);
  }

  return res;
}
