const VALID_SCOPES = ['model', 'key', 'team'] as const;

export function validateRateLimitRule(rule: {
  scope: string;
  scope_id: string;
  rps: number;
  tpm?: number;
}): string[] {
  const errors: string[] = [];
  if (!VALID_SCOPES.includes(rule.scope as typeof VALID_SCOPES[number])) {
    errors.push(`Scope must be one of: ${VALID_SCOPES.join(', ')}.`);
  }
  if (rule.rps < 0) errors.push('RPS must be non-negative.');
  if (rule.tpm != null && rule.tpm < 0) errors.push('TPM must be non-negative.');
  return errors;
}
