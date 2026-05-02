import CreateAPIKeyModal from "@/components/CreateAPIKeyModal";
import APIKeyRowActions from "@/components/APIKeyRowActions";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { describeModelAllowlist } from "@/lib/key-lifecycle";
import { APIKey, Model } from "@/lib/types";

const KEYS_ENDPOINT = "http://gateway:8080/v1/admin/keys";
const MODELS_ENDPOINT = "http://gateway:8080/v1/admin/models";

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

async function fetchModels(): Promise<{ models: Model[]; unavailable: boolean }> {
  try {
    const response = await fetch(MODELS_ENDPOINT, { cache: "no-store" });
    if (!response.ok) {
      return { models: [], unavailable: true };
    }

    const models = (await response.json()) as Model[];
    return { models, unavailable: false };
  } catch {
    return { models: [], unavailable: true };
  }
}

function formatDate(value: string | null | undefined): string {
  if (!value) {
    return "Never";
  }

  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? "Unknown" : parsed.toLocaleString();
}

function isExpired(value: string | null | undefined): boolean {
  if (!value) {
    return false;
  }

  const parsed = new Date(value);
  return !Number.isNaN(parsed.getTime()) && parsed.getTime() <= Date.now();
}

function summarizeMetadata(metadata: Record<string, string> | null | undefined): string {
  if (!metadata || Object.keys(metadata).length === 0) {
    return "None";
  }

  return Object.entries(metadata)
    .map(([key, value]) => `${key}: ${value}`)
    .join(", ");
}

export default async function KeysPage() {
  const [
    { keys, unavailable },
    { models, unavailable: modelsUnavailable },
  ] = await Promise.all([fetchKeys(), fetchModels()]);

  return (
    <div className="mx-auto w-full max-w-7xl">
      <div className="mb-6 flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">API Keys</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Live key state from the gateway in-memory cache.
          </p>
          {modelsUnavailable ? (
            <p className="mt-1 text-xs text-muted-foreground">
              Models are currently unavailable; allowlist labels may be limited.
            </p>
          ) : null}
        </div>
        <CreateAPIKeyModal models={models} />
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
                <TableHead>Expiration</TableHead>
                <TableHead>Models</TableHead>
                <TableHead>Metadata</TableHead>
                <TableHead>Updated</TableHead>
                <TableHead className="w-[1%]">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {unavailable ? (
                <TableRow>
                  <TableCell colSpan={9} className="text-muted-foreground">
                    Backend unavailable. Try refreshing in a moment.
                  </TableCell>
                </TableRow>
              ) : keys.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={9} className="text-muted-foreground">
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
                      <span
                        className={
                          key.is_active && !isExpired(key.expires_at)
                            ? "text-emerald-600"
                            : "text-muted-foreground"
                        }
                      >
                        {isExpired(key.expires_at)
                          ? "Expired"
                          : key.is_active
                            ? "Active"
                            : "Inactive"}
                      </span>
                    </TableCell>
                    <TableCell>{key.rate_limit_rps}</TableCell>
                    <TableCell>{formatDate(key.expires_at)}</TableCell>
                    <TableCell className="max-w-64 truncate">
                      {describeModelAllowlist(key.allowed_models, models)}
                    </TableCell>
                    <TableCell className="max-w-48 truncate">
                      {summarizeMetadata(key.metadata)}
                    </TableCell>
                    <TableCell>
                      {new Date(key.updated_at).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <APIKeyRowActions apiKey={key} models={models} />
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
