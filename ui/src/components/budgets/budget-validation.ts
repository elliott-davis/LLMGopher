export function validateBudgetPolicy(policy: {
  limit_usd: number;
  usage_usd: number;
  alert_threshold: number;
}): string[] {
  const errors: string[] = [];
  if (policy.limit_usd < 0) errors.push('Limit must be non-negative.');
  if (policy.usage_usd < 0) errors.push('Usage must be non-negative.');
  if (policy.alert_threshold <= 0 || policy.alert_threshold > 1) {
    errors.push('Alert threshold must be between 0 (exclusive) and 1 (inclusive).');
  }
  return errors;
}
