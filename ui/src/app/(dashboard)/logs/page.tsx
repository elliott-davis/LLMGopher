import { parseLogsFilter } from '@/lib/query-state';
import { fetchLogs } from '@/components/logs/logs-data';
import { LogsClient } from '@/components/logs/logs-client';

interface LogsPageProps {
  searchParams: Promise<Record<string, string>>;
}

export default async function LogsPage({ searchParams }: LogsPageProps) {
  const params = await searchParams;
  const filter = parseLogsFilter(params);
  const { rows, total, unavailable } = await fetchLogs(filter);
  return <LogsClient rows={rows} total={total} unavailable={unavailable} filter={filter} />;
}
