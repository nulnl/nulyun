import { fetchURL, fetchJSON } from "./utils";

export interface PasskeyCredential {
  id: number;
  name: string;
  createdAt: string;
  lastUsedAt: string;
}

export interface RegisterPasskeyRequest {
  name: string;
}

// Get all passkeys for the current user
export async function listPasskeys(): Promise<PasskeyCredential[]> {
  return fetchJSON<PasskeyCredential[]>("/api/passkeys", {});
}

// Begin passkey registration
export async function beginRegistration(): Promise<any> {
  const res = await fetchURL("/api/passkeys/register/begin", {
    method: "POST",
  });
  const options = await res.json();

  // Normalize PascalCase -> camelCase for @simplewebauthn/browser
  // Older Go struct serializes as { "Response": { ... Challenge, RelyingParty, User } }
  // Map either `PublicKey` or `Response` (PascalCase) to `publicKey` (camelCase) expected by the browser lib.
  const src =
    (options as any).PublicKey ||
    (options as any).Response ||
    (options as any).response;
  if (src) {
    (options as any).publicKey = src;
    delete (options as any).PublicKey;
    delete (options as any).Response;
    delete (options as any).response;

    const pk = (options as any).publicKey;
    if (!pk) return options;

    if (pk.Challenge && !pk.challenge) pk.challenge = pk.Challenge;
    if (pk.challenge && typeof pk.challenge === "string") {
      // keep as-is; @simplewebauthn will handle base64url strings
    }

    if (pk.RelyingParty && !pk.rp) pk.rp = pk.RelyingParty;
    if (pk.RP && !pk.rp) pk.rp = pk.RP;

    if (pk.User && !pk.user) pk.user = pk.User;
    if (pk.User && pk.User.ID && !pk.user.id) pk.user.id = pk.User.ID;

    if (pk.Parameters && !pk.pubKeyCredParams)
      pk.pubKeyCredParams = pk.Parameters;
    if (pk.PubKeyCredParams && !pk.pubKeyCredParams)
      pk.pubKeyCredParams = pk.PubKeyCredParams;

    if (pk.CredentialExcludeList && !pk.excludeCredentials)
      pk.excludeCredentials = pk.CredentialExcludeList;
    if (pk.ExcludeCredentials && !pk.excludeCredentials)
      pk.excludeCredentials = pk.ExcludeCredentials;

    if (pk.AuthenticatorSelection && !pk.authenticatorSelection)
      pk.authenticatorSelection = pk.AuthenticatorSelection;
    if (pk.Attestation && !pk.attestation) pk.attestation = pk.Attestation;
    if (pk.Timeout && !pk.timeout) pk.timeout = pk.Timeout;
  }

  return options;
}

// Finish passkey registration
export async function finishRegistration(
  credential: any,
  name: string
): Promise<PasskeyCredential> {
  return fetchJSON<PasskeyCredential>("/api/passkeys/register/finish", {
    method: "POST",
    body: JSON.stringify({ name, ...credential }),
  });
}

// Delete a passkey
export async function deletePasskey(id: number): Promise<void> {
  await fetchURL(`/api/passkeys/${id}`, {
    method: "DELETE",
  });
}

// Begin passkey login
export async function beginLogin(): Promise<{
  options: any;
  sessionId: string;
}> {
  const res = await fetch("/api/passkey/login/begin", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!res.ok) {
    throw new Error(`HTTP error! status: ${res.status}`);
  }

  const sessionId = res.headers.get("X-Passkey-Session-ID");
  if (!sessionId) {
    throw new Error("No session ID in response");
  }

  const options = await res.json();
  return { options, sessionId };
}

// Finish passkey login
export async function finishLogin(
  credential: any,
  sessionId: string
): Promise<any> {
  const res = await fetch("/api/passkey/login/finish", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Passkey-Session-ID": sessionId,
    },
    body: JSON.stringify(credential),
  });

  if (!res.ok) {
    throw new Error(`HTTP error! status: ${res.status}`);
  }

  return res.json();
}
