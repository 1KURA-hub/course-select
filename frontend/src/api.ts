import type { ApiResponse, AuthUser, BenchmarkStatus, Course, Selection } from "./types";

const apiBase = import.meta.env.DEV ? "http://127.0.0.1:8080" : "";
export const tokenKey = "course_select_token";
export const userKey = "course_select_user";

export function loadAuth(): AuthUser | null {
  const token = localStorage.getItem(tokenKey);
  const raw = localStorage.getItem(userKey);
  if (!token || !raw) return null;
  try {
    const user = JSON.parse(raw) as Omit<AuthUser, "token">;
    return { ...user, token };
  } catch {
    return null;
  }
}

export function saveAuth(auth: AuthUser) {
  localStorage.setItem(tokenKey, auth.token);
  localStorage.setItem(userKey, JSON.stringify({ name: auth.name, id: auth.id }));
}

export function clearAuth() {
  localStorage.removeItem(tokenKey);
  localStorage.removeItem(userKey);
}

export async function apiRequest<T>(path: string, token?: string, options: RequestInit = {}) {
  const headers = new Headers(options.headers);
  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (options.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const resp = await fetch(`${apiBase}${path}`, { ...options, headers });
  const text = await resp.text();
  const payload = text ? (JSON.parse(text) as ApiResponse<T>) : {};
  if (!resp.ok) {
    throw new Error(payload.msg || `请求失败: ${resp.status}`);
  }
  return payload;
}

export const api = {
  login: (sid: string, password: string) =>
    apiRequest("/login", undefined, { method: "POST", body: JSON.stringify({ sid, password }) }),
  demoLogin: () => apiRequest("/demo-login", undefined, { method: "POST" }),
  getCourses: () => apiRequest<Course[]>("/courses"),
  getCourse: (id: number) => apiRequest<Course>(`/courses/${id}`),
  getSelections: (token: string) => apiRequest<Selection[]>("/auth/selections", token),
  selectCourse: (id: number, token: string) => apiRequest(`/auth/select/${id}`, token, { method: "POST" }),
  dropCourse: (id: number, token: string) => apiRequest(`/auth/select/${id}`, token, { method: "DELETE" }),
  getResult: (id: number, token: string) => apiRequest(`/auth/result/${id}`, token),
  startBenchmark: (payload: { stock: number; users: number; duration: string; course_id?: number }) =>
    apiRequest<BenchmarkStatus>("/benchmark/start", undefined, { method: "POST", body: JSON.stringify(payload) }),
  getBenchmarkStatus: () => apiRequest<BenchmarkStatus>("/benchmark/status")
};
