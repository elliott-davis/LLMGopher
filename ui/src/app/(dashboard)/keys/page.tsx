import CreateAPIKeyModal from "@/components/CreateAPIKeyModal";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { APIKey } from "@/lib/types";

const KEYS_ENDPOINT = "http://gateway:8080/v1/admin/keys";

async function fetchKeys(): Promise<{ keys: APIKey[]; unavailable: boolean }> {
  try {
    const response = await fetch(KEYS_ENDPOINT, { cache: "no-store" });
    if (!response.ok) {
      return { keys: [], unavailable: true };
    }

    const keys = (await response.json()) as APIKey[];
    return { keys, unavailable: false };
  } catch {
    return { keys: [], unavailable: true };
  }
}

export default async function KeysPage() {
  const { keys, unavailable } = await fetchKeys();

  return (
    <div className="mx-auto w-full max-w-7xl">
      <div className="mb-6 flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">API Keys</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Live key state from the gateway in-memory cache.
          </p>
        </div>
        <CreateAPIKeyModal />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Key Inventory</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Key Hash</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Rate Limit (RPS)</TableHead>
                <TableHead>Updated</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {unavailable ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-muted-foreground">
                    Backend unavailable. Try refreshing in a moment.
                  </TableCell>
                </TableRow>
              ) : keys.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-muted-foreground">
                    No API keys are currently loaded.
                  </TableCell>
                </TableRow>
              ) : (
                keys.map((key) => (
                  <TableRow key={key.id}>
                    <TableCell className="font-medium">{key.name}</TableCell>
                    <TableCell className="font-mono text-xs">
                      {key.key_hash}
                    </TableCell>
                    <TableCell>
                      {key.is_active ? "Active" : "Inactive"}
                    </TableCell>
                    <TableCell>{key.rate_limit_rps}</TableCell>
                    <TableCell>
                      {new Date(key.updated_at).toLocaleString()}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
