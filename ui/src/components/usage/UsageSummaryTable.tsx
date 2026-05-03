import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatSmallUSD } from "@/lib/analytics";
import { UsageSummaryRow } from "@/lib/types";

interface UsageSummaryTableProps {
  rows: UsageSummaryRow[];
}

export default function UsageSummaryTable({ rows }: UsageSummaryTableProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Grouped Usage Summary</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Group</TableHead>
              <TableHead>Requests</TableHead>
              <TableHead>Prompt Tokens</TableHead>
              <TableHead>Completion Tokens</TableHead>
              <TableHead>Total Tokens</TableHead>
              <TableHead>Errors</TableHead>
              <TableHead>Avg Latency (ms)</TableHead>
              <TableHead>Cost</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => (
              <TableRow key={row.group}>
                <TableCell className="font-medium">{row.group}</TableCell>
                <TableCell>{row.requests.toLocaleString()}</TableCell>
                <TableCell>{row.prompt_tokens.toLocaleString()}</TableCell>
                <TableCell>{row.completion_tokens.toLocaleString()}</TableCell>
                <TableCell>{row.total_tokens.toLocaleString()}</TableCell>
                <TableCell>{row.errors.toLocaleString()}</TableCell>
                <TableCell>{row.avg_latency_ms.toFixed(1)}</TableCell>
                <TableCell>{formatSmallUSD(row.cost_usd)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
