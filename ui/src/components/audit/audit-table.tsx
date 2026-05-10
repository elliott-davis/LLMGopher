'use client';

import type { AuditRecord } from '@/lib/types';
import { formatLatency, formatCurrency, formatTokenCount } from '@/lib/admin-surface-format';
import { redactErrorSummary } from '@/lib/redaction';
import { TableRow, TableCell } from '@/components/ui/table';

interface AuditRowProps {
  entry: AuditRecord;
}

export function AuditTableRow({ entry }: AuditRowProps) {
  return (
    <TableRow data-testid={`audit-row-${entry.id}`}>
      <TableCell className="font-mono text-xs">{entry.request_id}</TableCell>
      <TableCell>{entry.api_key_id}</TableCell>
      <TableCell>{entry.model}</TableCell>
      <TableCell>{entry.provider}</TableCell>
      <TableCell>
        <span
          data-testid="audit-status"
          className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium ${
            entry.status_code >= 500 ? 'bg-red-100 text-red-700' :
            entry.status_code >= 400 ? 'bg-amber-100 text-amber-700' :
            'bg-green-100 text-green-700'
          }`}
        >
          {entry.status_code}
        </span>
      </TableCell>
      <TableCell>{formatTokenCount(entry.total_tokens)}</TableCell>
      <TableCell>{formatCurrency(entry.cost_usd)}</TableCell>
      <TableCell>{formatLatency(entry.latency_ms)}</TableCell>
      {entry.error_message && (
        <TableCell className="text-xs text-muted-foreground">
          {redactErrorSummary(entry.error_message)}
        </TableCell>
      )}
    </TableRow>
  );
}
