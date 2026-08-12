import { API_BASE_URL } from '../config/api';
import { SessionHistoryResponse, SessionResponse } from '../types/Training';

export class ApiError extends Error {
  public readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: { Accept: 'application/json', ...init?.headers },
    ...init,
  });
  if (!response.ok) {
    let message = `Request failed (${response.status})`;
    try {
      const body = await response.json() as { message?: string; error?: string };
      message = body.message || body.error || message;
    } catch {
      // An empty error body is valid for this API.
    }
    throw new ApiError(response.status, message);
  }
  return response.json() as Promise<T>;
}

export const trainingApi = {
  start: () => request<SessionResponse>('/training/sessions', { method: 'POST' }),
  active: () => request<SessionResponse>('/training/sessions/active'),
  get: (id: string) => request<SessionResponse>(`/training/sessions/${encodeURIComponent(id)}`),
  sync: (id: string) => request<SessionResponse>(`/training/sessions/${encodeURIComponent(id)}/sync`, { method: 'POST' }),
  abort: (id: string) => request<SessionResponse>(`/training/sessions/${encodeURIComponent(id)}/abort`, { method: 'POST' }),
  history: (page: number) => request<SessionHistoryResponse>(`/training/sessions?page=${page}`),
};
