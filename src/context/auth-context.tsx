"use client";

import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";
import { apiFetch, saveTokens, clearTokens, getAccessToken } from "@/lib/api";

export type UserRole = "ORGANIZER" | "PARTICIPANT";

export interface AuthUser {
  id: string;
  name: string;
  email: string;
  role: UserRole;
}

interface AuthContextValue {
  user: AuthUser | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string, role: UserRole) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const stored = localStorage.getItem("pulse_user");
    const token = getAccessToken();
    if (stored && token) {
      try { setUser(JSON.parse(stored)); } catch { clearTokens(); }
      // Sync role from server in background (catches stale localStorage)
      fetch(
        `${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/v1/auth/me`,
        { headers: { Authorization: `Bearer ${token}` } }
      ).then((r) => {
        if (r.ok) r.json().then((me) => {
          localStorage.setItem("pulse_user", JSON.stringify(me));
          setUser(me);
        });
      }).catch(() => {});
    }
    setLoading(false);
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/v1/auth/login`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      }
    );
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error ?? "Неверный email или пароль");
    }
    const { accessToken, refreshToken } = await res.json();
    saveTokens(accessToken, refreshToken);
    // Получаем профиль пользователя
    const meRes = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/v1/auth/me`,
      { headers: { Authorization: `Bearer ${accessToken}` } }
    );
    if (meRes.ok) {
      const me: AuthUser = await meRes.json();
      localStorage.setItem("pulse_user", JSON.stringify(me));
      setUser(me);
    }
  }, []);

  const register = useCallback(async (name: string, email: string, password: string, role: UserRole) => {
    const res = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/v1/auth/register`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, email, password, role }),
      }
    );
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error ?? "Ошибка регистрации");
    }
    const { accessToken, refreshToken } = await res.json();
    saveTokens(accessToken, refreshToken);
    // Сохраняем базовый профиль сразу, не ждём /me
    const u: AuthUser = { id: "", name, email, role };
    localStorage.setItem("pulse_user", JSON.stringify(u));
    setUser(u);
  }, []);

  const logout = useCallback(() => {
    clearTokens();
    setUser(null);
    window.location.href = "/login";
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
  return ctx;
}
