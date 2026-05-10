import { fetchRateLimits } from '@/components/rate-limits/rate-limits-data';
import { RateLimitsClient } from '@/components/rate-limits/rate-limits-client';

export default async function RateLimitsPage() {
  const { rules, unavailable } = await fetchRateLimits();
  return <RateLimitsClient rules={rules} unavailable={unavailable} />;
}
