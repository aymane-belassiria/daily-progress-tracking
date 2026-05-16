const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:4000";
const TOKEN_KEY = "progress-tracker-token";

export function getStoredToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function setStoredToken(token) {
  if (!token) {
    localStorage.removeItem(TOKEN_KEY);
    return;
  }

  localStorage.setItem(TOKEN_KEY, token);
}

async function request(path, options = {}) {
  const token = getStoredToken();
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers || {})
    }
  });

  if (response.status === 204) {
    return null;
  }

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || "Request failed.");
  }

  return data;
}

export const api = {
  login(payload) {
    return request("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  },
  me() {
    return request("/api/auth/me");
  },
  dashboard() {
    return request("/api/dashboard");
  },
  saveGoal(payload, id) {
    return request(id ? `/api/goals/${id}` : "/api/goals", {
      method: id ? "PUT" : "POST",
      body: JSON.stringify(payload)
    });
  },
  deleteGoal(id) {
    return request(`/api/goals/${id}`, { method: "DELETE" });
  },
  saveEntry(payload) {
    return request("/api/entries", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  },
  generateRoadmap(prompt) {
    return request("/api/ai/roadmap", {
      method: "POST",
      body: JSON.stringify({ prompt })
    });
  },
  generateHints(prompt) {
    return request("/api/ai/hints", {
      method: "POST",
      body: JSON.stringify({ prompt })
    });
  }
};
