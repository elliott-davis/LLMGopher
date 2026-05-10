'use client';

import type { LogRow } from '@/lib/admin-surface-contracts';
import { formatLatency } from '@/lib/admin-surface-format';
import { TableRow, TableCell } from '@/components/ui/table';

interface LogRowProps {
  row: LogRow;
  onClick: (id: string) => void;
}

export function LogTableRow({ row, onClick }: LogRowProps) {
  const isFallback = row.provider_chain.length > 1 && row.provider_chain[0]?.status === 'failed';
  return (
    <TableRow
      data-testid={`log-row-${row.id}`}
      onClick={() => onClick(row.id)}
      className="cursor-pointer hover:bg-muted/50"
    >
      <TableCell>
        <StatusBadge statusCode={row.status_code} />
      </TableCell>
      <TableCell className="font-mono text-xs">{row.request_id}</TableCell>
      <TableCell>{row.model}</TableCell>
      <TableCell>{row.api_key_id}</TableCell>
      <TableCell>{formatLatency(row.latency_ms)}</TableCell>
      <TableCell className="text-xs text-muted-foreground">
        {new Date(row.timestamp).toLocaleTimeString()}
      </TableCell>
      {isFallback && (
        <TableCell>
          <span className="text-xs text-amber-600" aria-label="Fallback route used">Fallback</span>
        </TableCell>
      )}
    </TableRow>
  );
}

interface StatusBadgeProps {
  statusCode: number;
}

export function StatusBadge({ statusCode }: StatusBadgeProps) {
  const isError = statusCode >= 500;
  const isClientError = statusCode >= 400 && statusCode < 500;
  return (
    <span
      data-testid="status-badge"
      data-status={statusCode}
      className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium ${
        isError ? 'bg-red-100 text-red-700' :
        isClientError ? 'bg-amber-100 text-amber-700' :
        'bg-green-100 text-green-700'
      }`}
    >
      {statusCode}
    </span>
  );
}

interface ProviderChainSummaryProps {
  chain: LogRow['provider_chain'];
}

export function ProviderChainSummary({ chain }: ProviderChainSummaryProps) {
  if (chain.length === 0) return <span className="text-muted-foreground text-xs">—</span>;
  return (
    <span className="text-xs">
      {chain.map((stage, i) => (
        <span key={stage.provider_id} className={stage.status === 'failed' ? 'text-red-600' : 'text-green-600'}>
          {i > 0 && ' → '}{stage.provider_id}
        </span>
      ))}
    </span>
  );
}
