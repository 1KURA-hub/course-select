const BASE = "";

async function request(path, options = {}) {
  const headers = { "Content-Type": "application/json" };
  const token = localStorage.getItem("token");
  if (token) headers["Authorization"] = `Bearer ${token}`;
  const res = await fetch(BASE + path, { ...options, headers: { ...headers, ...options.headers } });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.msg || data.error || `HTTP ${res.status}`);
  return data;
}

export const api = {
  register: (sid, password, name) =>
    request("/register", { method: "POST", body: JSON.stringify({ sid, password, name }) }),

  login: (sid, password) =>
    request("/login", { method: "POST", body: JSON.stringify({ sid, password }) }),

  getCourses: () => request("/courses"),

  getCourse: (id) => request(`/courses/${id}`),

  selectCourse: (id) =>
    request(`/auth/select/${id}`, { method: "POST" }),

  dropCourse: (id) =>
    request(`/auth/select/${id}`, { method: "DELETE" }),

  getResult: (id) => request(`/auth/result/${id}`),

  getSelections: () => request("/auth/selections"),
};
