const REDACTED = '[REDACTED]';
const DEFAULT_PREVIEW_LENGTH = 120;

const SENSITIVE_HEADER_PATTERNS = [
  'authorization',
  'cookie',
  'x-api-key',
  'x-auth-token',
  'x-secret',
  'x-credential',
  'api-key',
  'token',
  'secret',
  'credential',
  'key',
  'password',
  'bearer',
];

const HEX_BASE64_RE = /^[a-zA-Z0-9+/=_-]{20,}$/;
const BEARER_RE = /^Bearer\s/i;
const SK_RE = /^sk-/;
const SENSITIVE_WORD_RE = /\b(key|secret|token|password|credential)\b/i;

export function redactHeaders(headers: Record<string, string>): Record<string, string> {
  const result: Record<string, string> = {};
  for (const [name, value] of Object.entries(headers)) {
    const lower = name.toLowerCase();
    const isSensitive = SENSITIVE_HEADER_PATTERNS.some((pattern) => lower === pattern || lower.includes(pattern));
    result[name] = isSensitive ? REDACTED : value;
  }
  return result;
}

export function redactValue(value: string): string {
  if (BEARER_RE.test(value)) return REDACTED;
  if (SK_RE.test(value)) return REDACTED;
  if (HEX_BASE64_RE.test(value)) return REDACTED;
  if (SENSITIVE_WORD_RE.test(value)) return REDACTED;
  return value;
}

export function truncatePreview(text: string, maxLength: number = DEFAULT_PREVIEW_LENGTH): string {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength) + '…';
}

export function redactPromptPreview(prompt: string): string {
  return truncatePreview(prompt);
}

export function redactResponsePreview(response: string): string {
  return truncatePreview(response);
}

export function redactErrorSummary(error: string): string {
  if (BEARER_RE.test(error) || SK_RE.test(error) || HEX_BASE64_RE.test(error) || SENSITIVE_WORD_RE.test(error)) {
    return REDACTED;
  }
  return truncatePreview(error);
}
