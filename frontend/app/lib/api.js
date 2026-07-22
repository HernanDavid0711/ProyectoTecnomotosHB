export const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8081";

export const storageKeys = {
  token: "tecnomotos_token",
  user: "tecnomotos_user",
  expiresAt: "tecnomotos_expires_at"
};

export function getSession() {
  if (typeof window === "undefined") return null;
  const token = localStorage.getItem(storageKeys.token);
  const user = localStorage.getItem(storageKeys.user);
  if (!token || !user) return null;

  try {
    return {
      token,
      user: JSON.parse(user),
      expiresAt: localStorage.getItem(storageKeys.expiresAt)
    };
  } catch {
    clearSession();
    return null;
  }
}

export function saveSession(payload) {
  localStorage.setItem(storageKeys.token, payload.token);
  localStorage.setItem(storageKeys.user, JSON.stringify(payload.data));
  localStorage.setItem(storageKeys.expiresAt, payload.expires_at);
}

export function clearSession() {
  localStorage.removeItem(storageKeys.token);
  localStorage.removeItem(storageKeys.user);
  localStorage.removeItem(storageKeys.expiresAt);
}

export function authHeaders() {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${localStorage.getItem(storageKeys.token)}`
  };
}

export async function apiGet(path) {
  const response = await fetch(`${API_BASE_URL}${path}`, { headers: authHeaders() });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.message || "No se pudo consultar la informacion");
  return payload;
}

export async function apiPost(path, body) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(body)
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.message || "No se pudo guardar la informacion");
  return payload;
}

export async function apiPut(path, body) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: "PUT",
    headers: authHeaders(),
    body: JSON.stringify(body)
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.message || "No se pudo actualizar la informacion");
  return payload;
}

export async function apiDelete(path) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: "DELETE",
    headers: authHeaders()
  });
  if (response.status === 204) return {};
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.message || "No se pudo eliminar la informacion");
  return payload;
}

export async function login(usuario, contrasena) {
  const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ usuario, contrasena })
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.message || "No se pudo iniciar sesion");
  return payload;
}
