import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { api } from "./api";

interface User { id: number; username: string; role: string; }
interface AuthCtx { user: User | null; loading: boolean; login: (u: string, p: string) => Promise<void>; logout: () => Promise<void>; }
const Ctx = createContext<AuthCtx>(null as any);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api("/api/v1/auth/me").then((u: any) => setUser(u)).catch(() => { sessionStorage.removeItem("vertex_csrf"); }).finally(() => setLoading(false));
  }, []);

  async function login(username: string, password: string) {
    const res: any = await api("/api/v1/auth/login", { method: "POST", body: JSON.stringify({ username, password }) });
    setUser(res.user);
    if (res.csrf) sessionStorage.setItem("vertex_csrf", res.csrf);
  }
  async function logout() {
    await api("/api/v1/auth/logout", { method: "POST" });
    setUser(null);
    sessionStorage.removeItem("vertex_csrf");
  }
  return <Ctx.Provider value={{ user, loading, login, logout }}>{children}</Ctx.Provider>;
}

export function useAuth() { return useContext(Ctx); }
