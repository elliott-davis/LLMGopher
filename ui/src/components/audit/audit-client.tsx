'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import type { AuditRecord } from '@/lib/types';
import type { AuditFilter } from '@/lib/admin-surface-contracts';
import { serializeAuditFilter, filterToSearchString } from '@/lib/query-state';
import { PageHeader } from '@/components/admin-surface-layout';
import { UnavailableState, EmptyState } from '@/components/admin-surface-state';
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { AuditTableRow } from './audit-table';

interface AuditClientProps {
  entries: AuditRecord[];
  total: number;
  unavailable: boolean;
  filter: AuditFilter;
}

export function AuditClient({ entries, total: _total, unavailable, filter }: AuditClientProps) {
  const router = useRouter();
  const [actor, setActor] = useState(filter.actor ?? '');
  const [action, setAction] = useState(filter.action ?? '');
  const [from, setFrom] = useState(filter.from ?? '');
  const [to, setTo] = useState(filter.to ?? '');

  const applyFilter = () => {
    const next: AuditFilter = {
      actor: actor || undefined,
      action: action || undefined,
      from: from || undefined,
      to: to || undefined,
    };
    const qs = filterToSearchString(serializeAuditFilter(next));
    router.push(`/audit${qs}`);
  };

  return (
    <>
      <PageHeader title="Audit log" />
      <div className="content-area">
        <div className="flex flex-wrap gap-2 mb-4">
          <input
            data-testid="audit-filter-actor"
            placeholder="Actor (API key ID)"
            value={actor}
            onChange={(e) => setActor(e.target.value)}
            className="border rounded px-2 py-1 text-sm"
          />
          <input
            data-testid="audit-filter-action"
            placeholder="Action / Model"
            value={action}
            onChange={(e) => setAction(e.target.value)}
            className="border rounded px-2 py-1 text-sm"
          />
          <input
            data-testid="audit-filter-from"
            type="date"
            value={from}
            onChange={(e) => setFrom(e.target.value)}
            className="border rounded px-2 py-1 text-sm"
          />
          <input
            data-testid="audit-filter-to"
            type="date"
            value={to}
            onChange={(e) => setTo(e.target.value)}
            className="border rounded px-2 py-1 text-sm"
          />
          <button
            data-testid="audit-filter-apply"
            onClick={applyFilter}
            className="px-3 py-1 rounded text-sm bg-primary text-primary-foreground"
          >
            Apply
          </button>
        </div>

        {unavailable && <UnavailableState message="Audit log is not yet available in this environment." />}
        {!unavailable && entries.length === 0 && <EmptyState message="No audit entries match the current filter." />}
        {!unavailable && entries.length > 0 && (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Request ID</TableHead>
                <TableHead>Actor</TableHead>
                <TableHead>Model</TableHead>
                <TableHead>Provider</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Tokens</TableHead>
                <TableHead>Cost</TableHead>
                <TableHead>Latency</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {entries.map((entry) => (
                <AuditTableRow key={entry.id} entry={entry} />
              ))}
            </TableBody>
          </Table>
        )}
      </div>
    </>
  );
}
