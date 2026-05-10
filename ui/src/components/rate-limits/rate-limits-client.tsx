'use client';

import type { RateLimitRule } from '@/lib/admin-surface-contracts';
import { PageHeader } from '@/components/admin-surface-layout';
import { UnavailableState, EmptyState } from '@/components/admin-surface-state';
import { RateLimitRules } from './rate-limit-rules';

export function RateLimitsClient({ rules, unavailable }: { rules: RateLimitRule[]; unavailable: boolean }) {
  return (
    <>
      <PageHeader title="Rate Limits" />
      <div className="content-area">
        {unavailable && <UnavailableState message="Rate limits are not yet available in this environment." />}
        {!unavailable && rules.length === 0 && <EmptyState message="No rate limit rules configured." />}
        {!unavailable && rules.length > 0 && <RateLimitRules rules={rules} />}
      </div>
    </>
  );
}
