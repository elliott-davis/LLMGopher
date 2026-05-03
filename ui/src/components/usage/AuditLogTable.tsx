import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatSmallUSD, toFilterQueryParams } from "@/lib/analytics";
import { AnalyticsFilters, AuditRecord } from "@/lib/types";

interface AuditLogTableProps {
  rows: AuditRecord[];
  total: number;
  limit: number;
  offset: number;
  filters: AnalyticsFilters;
}

function toPageHref(filters: AnalyticsFilters, offset: number): string {
  const next = toFilterQueryParams({ ...filters, offset });
  return `/usage?${next.toString()}`;
}

export default function AuditLogTable({
  rows,
  total,
  limit,
  offset,
  filters,
}: AuditLogTableProps) {
  const previousOffset = Math.max(0, offset - limit);
  const nextOffset = offset + limit;
  const hasPrevious = offset > 0;
  const hasNext = nextOffset < total;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Audit Log Search</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Timestamp</TableHead>
              <TableHead>Request ID</TableHead>
              <TableHead>API Key ID</TableHead>
              <TableHead>Model</TableHead>
              <TableHead>Provider</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Latency (ms)</TableHead>
              <TableHead>Total Tokens</TableHead>
              <TableHead>Cost</TableHead>
              <TableHead>Streaming</TableHead>
              <TableHead>Error</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => (
              <TableRow
                key={row.id}
                className={row.status_code >= 400 ? "bg-red-50/40" : undefined}
              >
                <TableCell>{new Date(row.created_at).toLocaleString()}</TableCell>
                <TableCell className="font-mono text-xs">{row.request_id}</TableCell>
                <TableCell>{row.api_key_id}</TableCell>
                <TableCell>{row.model}</TableCell>
                <TableCell>{row.provider}</TableCell>
                <TableCell>{row.status_code}</TableCell>
                <TableCell>{row.latency_ms}</TableCell>
                <TableCell>{row.total_tokens.toLocaleString()}</TableCell>
                <TableCell>{formatSmallUSD(row.cost_usd)}</TableCell>
                <TableCell>{row.streaming ? "yes" : "no"}</TableCell>
                <TableCell>{row.error_message || "-"}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>

        <div className="flex items-center justify-between text-sm">
          <p className="text-muted-foreground">
            Showing {rows.length} of {total} records
          </p>
          <div className="flex items-center gap-3">
            {hasPrevious ? (
              <Link className="text-primary hover:underline" href={toPageHref(filters, previousOffset)}>
                Previous
              </Link>
            ) : (
              <span className="text-muted-foreground">Previous</span>
            )}
            {hasNext ? (
              <Link className="text-primary hover:underline" href={toPageHref(filters, nextOffset)}>
                Next
              </Link>
            ) : (
              <span className="text-muted-foreground">Next</span>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
