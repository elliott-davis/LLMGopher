import { parseAuditFilter } from '@/lib/query-state';
import { fetchAuditEntries } from '@/components/audit/audit-data';
import { AuditClient } from '@/components/audit/audit-client';

interface AuditPageProps {
  searchParams: Promise<Record<string, string>>;
}

export default async function AuditPage({ searchParams }: AuditPageProps) {
  const params = await searchParams;
  const filter = parseAuditFilter(params);
  const { entries, total, unavailable } = await fetchAuditEntries(filter);
  return <AuditClient entries={entries} total={total} unavailable={unavailable} filter={filter} />;
}
