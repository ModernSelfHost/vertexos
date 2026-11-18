const BASE = "";

export async function api<T = any>(path: string, opts: RequestInit = {}): Promise<T> {
  const csrf = sessionStorage.getItem("vertex_csrf");
  const headers: Record<string, string> = { "Content-Type": "application/json", ...(opts.headers as any || {}) };
  if (csrf && !(opts.method === "GET" || opts.method === "HEAD")) headers["X-CSRF-Token"] = csrf;
  const res = await fetch(BASE + path, {
    credentials: "include",
    headers,
    ...opts,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}
