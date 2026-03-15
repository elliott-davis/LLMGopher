export type SupportedCredentialProvider = "openai" | "anthropic" | "google";

type ValidationAPIResponse = {
  valid?: boolean;
  error?: string;
  message?: string;
  code?: string;
};

export class CredentialValidationError extends Error {
  readonly code: string;

  constructor(message: string, code: string) {
    super(message);
    this.code = code;
  }
}

function mapValidationFailure(
  code: string | undefined,
  fallbackMessage: string
): CredentialValidationError {
  switch (code) {
    case "invalid_api_key":
      return new CredentialValidationError("Invalid API Key", code);
    case "quota_exceeded":
      return new CredentialValidationError("Quota Exceeded", code);
    case "network_error":
      return new CredentialValidationError("Network Error", code);
    default:
      return new CredentialValidationError(fallbackMessage, code ?? "provider_error");
  }
}

export async function validateCredential(
  provider: SupportedCredentialProvider,
  apiKey: string
): Promise<void> {
  const response = await fetch("/api/v1/credentials/validate", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      provider,
      apiKey,
    }),
  });

  if (response.ok) {
    return;
  }

  let payload: ValidationAPIResponse | null = null;
  try {
    payload = (await response.json()) as ValidationAPIResponse;
  } catch {
    payload = null;
  }

  const message = payload?.error ?? payload?.message ?? "Validation failed";
  throw mapValidationFailure(payload?.code, message);
}
