import { APIKeyFormValues } from "@/lib/types";

export function extractGatewayErrorMessage(payload: unknown, fallback: string): string {
  if (typeof payload !== "object" || payload === null) {
    return fallback;
  }

  const maybeError = (payload as { error?: unknown }).error;
  if (typeof maybeError === "object" && maybeError !== null) {
    const maybeNestedMessage = (maybeError as { message?: unknown }).message;
    if (typeof maybeNestedMessage === "string" && maybeNestedMessage.trim()) {
      return maybeNestedMessage;
    }
  }

  if (typeof maybeError === "string" && maybeError.trim()) {
    return maybeError;
  }

  const maybeMessage = (payload as { message?: unknown }).message;
  if (typeof maybeMessage === "string" && maybeMessage.trim()) {
    return maybeMessage;
  }

  return fallback;
}

function parseOptionalTimestamp(value: FormDataEntryValue | null): string | null {
  const timestamp = String(value ?? "").trim();
  if (!timestamp) {
    return null;
  }

  const parsed = new Date(timestamp);
  if (Number.isNaN(parsed.getTime())) {
    throw new Error("Expiration must be a valid date and time");
  }

  return parsed.toISOString();
}

export function parseAPIKeyMetadata(
  value: FormDataEntryValue | null
): Record<string, string> {
  const raw = String(value ?? "").trim();
  if (!raw) {
    return {};
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    throw new Error("Metadata must be valid JSON");
  }

  if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
    throw new Error("Metadata must be a JSON object");
  }

  const metadata: Record<string, string> = {};
  for (const [key, value] of Object.entries(parsed)) {
    if (typeof value !== "string") {
      throw new Error("Metadata values must be strings");
    }
    metadata[key] = value;
  }

  return metadata;
}

export function parseAPIKeyFormValues(formData: FormData): APIKeyFormValues {
  const name = String(formData.get("name") ?? "").trim();
  const rateLimitText = String(formData.get("rate_limit_rps") ?? "").trim();
  const rateLimitRPS = Number(rateLimitText);

  if (
    !name ||
    !Number.isFinite(rateLimitRPS) ||
    rateLimitRPS < 0 ||
    !Number.isInteger(rateLimitRPS)
  ) {
    throw new Error("Invalid API key form data");
  }

  const isActiveValue = formData.get("is_active");

  return {
    name,
    rate_limit_rps: rateLimitRPS,
    expires_at: parseOptionalTimestamp(formData.get("expires_at")),
    metadata: parseAPIKeyMetadata(formData.get("metadata")),
    allowed_models: formData
      .getAll("allowed_models")
      .map((value) => String(value).trim())
      .filter(Boolean),
    ...(isActiveValue === null
      ? {}
      : {
          is_active:
            String(isActiveValue) === "true" || String(isActiveValue) === "on",
        }),
  };
}
