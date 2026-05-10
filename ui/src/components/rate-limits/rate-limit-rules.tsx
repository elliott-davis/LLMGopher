'use client';

import type { RateLimitRule } from '@/lib/admin-surface-contracts';

export function RateLimitRules({ rules }: { rules: RateLimitRule[] }) {
  return (
    <div className="space-y-2">
      {rules.map((rule) => (
        <div
          key={rule.id}
          data-testid={`rate-limit-row-${rule.id}`}
          className="border rounded p-3 flex items-center justify-between"
        >
          <div>
            <p className="text-sm font-medium capitalize">{rule.scope}: {rule.scope_id}</p>
            <p className="text-xs text-muted-foreground">
              {rule.rps} RPS{rule.tpm != null ? ` · ${rule.tpm.toLocaleString()} TPM` : ''}
            </p>
          </div>
          <div className="flex items-center gap-2">
            {rule.tripped && (
              <span
                data-testid="rate-limit-tripped-pill"
                className="text-xs text-red-700 bg-red-100 px-2 py-0.5 rounded"
                aria-label="Rate limit currently tripped"
              >
                Tripped
              </span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
