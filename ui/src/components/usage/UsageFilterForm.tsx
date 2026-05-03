import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AnalyticsFilters } from "@/lib/types";

interface UsageFilterFormProps {
  filters: AnalyticsFilters;
}

export default function UsageFilterForm({ filters }: UsageFilterFormProps) {
  return (
    <form action="/usage" method="get" className="rounded-xl border p-4">
      <h2 className="text-base font-semibold tracking-tight">Filters</h2>
      <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-3">
        <label className="text-sm">
          <span className="mb-1 block text-muted-foreground">Group by</span>
          <select
            name="group_by"
            defaultValue={filters.group_by}
            className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm"
          >
            <option value="model">model</option>
            <option value="provider">provider</option>
            <option value="api_key">api_key</option>
          </select>
        </label>
        <label className="text-sm">
          <span className="mb-1 block text-muted-foreground">From</span>
          <Input name="from" placeholder="2026-04-01T00:00:00Z" defaultValue={filters.from ?? ""} />
        </label>
        <label className="text-sm">
          <span className="mb-1 block text-muted-foreground">To</span>
          <Input name="to" placeholder="2026-04-30T00:00:00Z" defaultValue={filters.to ?? ""} />
        </label>
        <label className="text-sm">
          <span className="mb-1 block text-muted-foreground">API Key ID</span>
          <Input name="api_key_id" defaultValue={filters.api_key_id ?? ""} />
        </label>
        <label className="text-sm">
          <span className="mb-1 block text-muted-foreground">Model</span>
          <Input name="model" defaultValue={filters.model ?? ""} />
        </label>
        <label className="text-sm">
          <span className="mb-1 block text-muted-foreground">Provider</span>
          <Input name="provider" defaultValue={filters.provider ?? ""} />
        </label>
        <label className="text-sm">
          <span className="mb-1 block text-muted-foreground">Status</span>
          <select
            name="status"
            defaultValue={filters.status ?? ""}
            className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm"
          >
            <option value="">all</option>
            <option value="success">success</option>
            <option value="error">error</option>
          </select>
        </label>
        <label className="text-sm">
          <span className="mb-1 block text-muted-foreground">Limit</span>
          <Input name="limit" type="number" min={1} max={1000} defaultValue={filters.limit} />
        </label>
        <label className="text-sm">
          <span className="mb-1 block text-muted-foreground">Offset</span>
          <Input name="offset" type="number" min={0} defaultValue={filters.offset} />
        </label>
      </div>

      <input type="hidden" name="prev_from" value={filters.from ?? ""} />
      <input type="hidden" name="prev_to" value={filters.to ?? ""} />
      <input type="hidden" name="prev_group_by" value={filters.group_by} />
      <input type="hidden" name="prev_api_key_id" value={filters.api_key_id ?? ""} />
      <input type="hidden" name="prev_model" value={filters.model ?? ""} />
      <input type="hidden" name="prev_provider" value={filters.provider ?? ""} />
      <input type="hidden" name="prev_status" value={filters.status ?? ""} />

      <div className="mt-4 flex gap-2">
        <Button type="submit">Apply filters</Button>
        <Button variant="outline" render={<a href="/usage" />}>
          Reset
        </Button>
      </div>
    </form>
  );
}
