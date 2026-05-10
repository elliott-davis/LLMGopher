'use client';

import { useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import type { LogRow, LogDetail, LogsFilter } from '@/lib/admin-surface-contracts';
import { serializeLogsFilter, filterToSearchString } from '@/lib/query-state';
import { PageHeader } from '@/components/admin-surface-layout';
import { EmptyState, UnavailableState } from '@/components/admin-surface-state';
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { LogTableRow } from './logs-table';
import { RequestInspector } from './request-inspector';

interface LogsClientProps {
  rows: LogRow[];
  total: number;
  unavailable: boolean;
  filter: LogsFilter;
}

const STATUS_FILTERS = ['all', '2xx', '4xx', '5xx', 'fallback'] as const;

export function LogsClient({ rows, total: _total, unavailable, filter }: LogsClientProps) {
  const router = useRouter();
  const [selectedDetail, setSelectedDetail] = useState<LogDetail | null>(null);
  const activeStatus = filter.status ?? 'all';

  const setFilter = useCallback((status: string) => {
    const next: LogsFilter = { ...filter, status: status === 'all' ? undefined : status as LogsFilter['status'] };
    const qs = filterToSearchString(serializeLogsFilter(next));
    router.push(`/logs${qs}`);
  }, [filter, router]);

  const openInspector = useCallback(async (id: string) => {
    try {
      const res = await fetch(`/api/logs/${id}`);
      if (res.ok) setSelectedDetail(await res.json());
    } catch {
      // ignore
    }
  }, []);

  return (
    <>
      <PageHeader title="Logs" />
      <div className="content-area">
        <div className="flex gap-2 mb-4" role="toolbar" aria-label="Filter by status">
          {STATUS_FILTERS.map((s) => (
            <button
              key={s}
              data-testid={`filter-status-${s}`}
              aria-pressed={activeStatus === s}
              onClick={() => setFilter(s)}
              className={`px-3 py-1 rounded text-sm border ${activeStatus === s ? 'bg-primary text-primary-foreground border-primary' : 'border-border bg-background'}`}
            >
              {s === 'all' ? 'All' : s}
            </button>
          ))}
        </div>

        {unavailable && <UnavailableState message="Logs are not yet available in this environment." />}
        {!unavailable && rows.length === 0 && (
          <EmptyState message="No log rows match the current filter." onClear={() => setFilter('all')} />
        )}
        {!unavailable && rows.length > 0 && (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Status</TableHead>
                <TableHead>Request ID</TableHead>
                <TableHead>Model</TableHead>
                <TableHead>API Key</TableHead>
                <TableHead>Latency</TableHead>
                <TableHead>Time</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((row) => (
                <LogTableRow key={row.id} row={row} onClick={openInspector} />
              ))}
            </TableBody>
          </Table>
        )}
      </div>

      {selectedDetail && (
        <RequestInspector detail={selectedDetail} onClose={() => setSelectedDetail(null)} />
      )}
    </>
  );
}
