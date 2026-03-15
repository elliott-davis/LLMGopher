import CreateProviderModal from "@/components/CreateProviderModal";
import ProviderRowActions from "@/components/ProviderRowActions";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Provider } from "@/lib/types";

const PROVIDERS_ENDPOINT = "http://gateway:8080/v1/admin/providers";

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

export default async function ProvidersPage() {
  const { providers, unavailable } = await fetchProviders();

  return (
    <div className="mx-auto w-full max-w-7xl space-y-6">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Providers</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Create and manage provider instances used by model aliases.
          </p>
        </div>
        <CreateProviderModal />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Configured Providers</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Provider ID</TableHead>
                <TableHead>Base URL</TableHead>
                <TableHead>Auth Type</TableHead>
                <TableHead>Credentials</TableHead>
                <TableHead>Updated</TableHead>
                <TableHead className="w-[1%]">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {unavailable ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-muted-foreground">
                    Backend unavailable. Try refreshing in a moment.
                  </TableCell>
                </TableRow>
              ) : providers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-muted-foreground">
                    No providers are currently loaded.
                  </TableCell>
                </TableRow>
              ) : (
                providers.map((provider) => (
                  <TableRow key={provider.id}>
                    <TableCell className="font-medium">{provider.name}</TableCell>
                    <TableCell className="font-mono text-xs">
                      {provider.id}
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      {provider.base_url}
                    </TableCell>
                    <TableCell>{provider.auth_type}</TableCell>
                    <TableCell>
                      {provider.has_credentials ? "Uploaded" : "None"}
                    </TableCell>
                    <TableCell>
                      {provider.updated_at
                        ? new Date(provider.updated_at).toLocaleString()
                        : "-"}
                    </TableCell>
                    <TableCell>
                      <ProviderRowActions provider={provider} />
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
