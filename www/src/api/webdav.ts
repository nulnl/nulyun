import { fetchURL, fetchJSON } from "./utils";

export interface WebDAVToken {
  id: number;
  userId: number;
  name: string;
  token: string;
  path: string;
  canRead: boolean;
  canWrite: boolean;
  canDelete: boolean;
  status: "active" | "suspended";
  createdAt: string;
  updatedAt: string;
}

export interface CreateTokenRequest {
  name: string;
  path: string;
  canRead: boolean;
  canWrite: boolean;
  canDelete: boolean;
}

export interface UpdateTokenRequest {
  name: string;
  path: string;
  canRead: boolean;
  canWrite: boolean;
  canDelete: boolean;
}

// Get all WebDAV tokens for the current user
export async function listTokens(): Promise<WebDAVToken[]> {
  return fetchJSON<WebDAVToken[]>("/api/webdav/tokens", {});
}

// Get details for a single token
export async function getToken(id: number): Promise<WebDAVToken> {
  return fetchJSON<WebDAVToken>(`/api/webdav/tokens/${id}`, {});
}

// Create a new WebDAV token
export async function createToken(
  data: CreateTokenRequest
): Promise<WebDAVToken> {
  return fetchJSON<WebDAVToken>("/api/webdav/tokens", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

// Update a WebDAV token
export async function updateToken(
  id: number,
  data: UpdateTokenRequest
): Promise<WebDAVToken> {
  return fetchJSON<WebDAVToken>(`/api/webdav/tokens/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

// Delete a WebDAV token
export async function deleteToken(id: number): Promise<void> {
  await fetchURL(`/api/webdav/tokens/${id}`, {
    method: "DELETE",
  });
}

// Suspend a WebDAV token
export async function suspendToken(id: number): Promise<void> {
  await fetchURL(`/api/webdav/tokens/${id}/suspend`, {
    method: "POST",
  });
}

// Activate a WebDAV token
export async function activateToken(id: number): Promise<void> {
  await fetchURL(`/api/webdav/tokens/${id}/activate`, {
    method: "POST",
  });
}
