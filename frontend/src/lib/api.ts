export const API_BASE = (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1").replace(/\/$/, "");

export async function apiFetch(path: string, init: RequestInit = {}) {
  return fetch(`${API_BASE}${path}`, {
    ...init,
    credentials: "include",
    headers: { "Content-Type": "application/json", ...init.headers },
  });
}

export function websocketURL(conversationID: string) {
  const url = new URL(API_BASE);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = `${url.pathname}/ws`.replace(/\/+/g, "/");
  url.searchParams.set("conversation_id", conversationID);
  return url.toString();
}

export type Message = {
  id: string;
  conversation_id: string;
  sender_type: "anonymous" | "admin";
  content: string;
  created_at: string;
  read_at: string | null;
};
