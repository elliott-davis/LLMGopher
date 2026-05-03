import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function UsageLoading() {
  return (
    <div className="mx-auto w-full max-w-7xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Usage & Audit</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Loading analytics panels...
        </p>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Preparing grouped usage summary</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Fetching analytics data from secured admin endpoints.
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Preparing daily usage trend</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Loading daily request, token, and spend data.
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Preparing audit log search</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Loading paginated request records with redacted fields.
        </CardContent>
      </Card>
    </div>
  );
}
