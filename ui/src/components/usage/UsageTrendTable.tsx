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
import { DailyUsagePoint } from "@/lib/types";

interface UsageTrendTableProps {
  rows: DailyUsagePoint[];
}

export default function UsageTrendTable({ rows }: UsageTrendTableProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Daily Usage Trend</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Date</TableHead>
              <TableHead>Requests</TableHead>
              <TableHead>Total Tokens</TableHead>
              <TableHead>Cost</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => (
              <TableRow key={row.date}>
                <TableCell>{row.date}</TableCell>
                <TableCell>{row.requests.toLocaleString()}</TableCell>
                <TableCell>{row.total_tokens.toLocaleString()}</TableCell>
                <TableCell>{formatSmallUSD(row.cost_usd)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
