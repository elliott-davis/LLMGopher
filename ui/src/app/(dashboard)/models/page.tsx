import Link from "next/link";
import CreateModelModal from "@/components/CreateModelModal";
import ModelRowActions from "@/components/ModelRowActions";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Model, Provider } from "@/lib/types";

const MODELS_ENDPOINT = "http://gateway:8080/v1/admin/models";
const PROVIDERS_ENDPOINT = "http://gateway:8080/v1/admin/providers";

async function fetchModels(): Promise<{
  models: Model[];
  unavailable: boolean;
}> {
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

async function fetchProviders(): Promise<{
  providers: Provider[];
  unavailable: boolean;
}> {
  try {
    const response = await fetch(PROVIDERS_ENDPOINT, { cache: "no-store" });
    if (!response.ok) {
      return { providers: [], unavailable: true };
    }

    const providers = (await response.json()) as Provider[];
    return { providers, unavailable: false };
  } catch {
    return { providers: [], unavailable: true };
  }
}

export default async function ModelsPage() {
  const [
    { models, unavailable },
    { providers, unavailable: providersUnavailable },
  ] = await Promise.all([fetchModels(), fetchProviders()]);

  return (
    <div className="mx-auto w-full max-w-7xl space-y-6">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Models</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Live model state from the gateway in-memory cache.
          </p>
          {providersUnavailable ? (
            <p className="mt-1 text-xs text-muted-foreground">
              Providers are currently unavailable.
            </p>
          ) : providers.length === 0 ? (
            <p className="mt-1 text-xs text-muted-foreground">
              No providers found. Create one first to add models.
            </p>
          ) : null}
        </div>
        <div className="flex items-center gap-2">
          <CreateModelModal providers={providers} />
          {providersUnavailable || providers.length === 0 ? (
            <Button variant="outline" render={<Link href="/providers" />}>
              Go to Providers
            </Button>
          ) : null}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Configured Models</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Alias</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Provider ID</TableHead>
                <TableHead>Context Window</TableHead>
                <TableHead>Updated</TableHead>
                <TableHead className="w-[1%]">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {unavailable ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-muted-foreground">
                    Backend unavailable. Try refreshing in a moment.
                  </TableCell>
                </TableRow>
              ) : models.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-muted-foreground">
                    No models are currently loaded.
                  </TableCell>
                </TableRow>
              ) : (
                models.map((model) => (
                  <TableRow key={model.id}>
                    <TableCell className="font-medium">{model.alias}</TableCell>
                    <TableCell>{model.name}</TableCell>
                    <TableCell className="font-mono text-xs">
                      {model.provider_id}
                    </TableCell>
                    <TableCell>{model.context_window}</TableCell>
                    <TableCell>
                      {new Date(model.updated_at).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <ModelRowActions model={model} providers={providers} />
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Routing Rules</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Future rules engine UI will define traffic splits, failover
          priorities, and policy constraints.
        </CardContent>
      </Card>
    </div>
  );
}
