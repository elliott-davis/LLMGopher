"use client";

import { APIKey, Model } from "@/lib/types";
import { buildModelAllowlistOptions } from "@/lib/key-lifecycle";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

type APIKeyLifecycleFormProps = {
  models: Model[];
  apiKey?: APIKey;
  disabled?: boolean;
  idPrefix: string;
};

function formatDateTimeLocal(value: string | null | undefined): string {
  if (!value) {
    return "";
  }

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "";
  }

  return parsed.toISOString().slice(0, 16);
}

function formatMetadata(value: Record<string, string> | null | undefined): string {
  if (!value || Object.keys(value).length === 0) {
    return "";
  }

  return JSON.stringify(value, null, 2);
}

export default function APIKeyLifecycleForm({
  models,
  apiKey,
  disabled = false,
  idPrefix,
}: APIKeyLifecycleFormProps) {
  const selectedModels = apiKey?.allowed_models ?? [];
  const modelOptions = buildModelAllowlistOptions(models, selectedModels);

  return (
    <>
      <div className="space-y-2">
        <Label htmlFor={`${idPrefix}-name`}>Name</Label>
        <Input
          id={`${idPrefix}-name`}
          name="name"
          required
          disabled={disabled}
          placeholder="Production Service Key"
          defaultValue={apiKey?.name ?? ""}
        />
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor={`${idPrefix}-rate-limit`}>Rate Limit (RPS)</Label>
          <Input
            id={`${idPrefix}-rate-limit`}
            name="rate_limit_rps"
            type="number"
            min={0}
            step={1}
            defaultValue={apiKey?.rate_limit_rps ?? 100}
            disabled={disabled}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor={`${idPrefix}-expires-at`}>Expiration</Label>
          <Input
            id={`${idPrefix}-expires-at`}
            name="expires_at"
            type="datetime-local"
            defaultValue={formatDateTimeLocal(apiKey?.expires_at)}
            disabled={disabled}
          />
        </div>
      </div>

      {apiKey ? (
        <input type="hidden" name="is_active" value={String(apiKey.is_active)} />
      ) : null}

      <div className="space-y-2">
        <Label htmlFor={`${idPrefix}-metadata`}>Metadata JSON</Label>
        <Textarea
          id={`${idPrefix}-metadata`}
          name="metadata"
          placeholder='{"owner":"platform"}'
          defaultValue={formatMetadata(apiKey?.metadata)}
          disabled={disabled}
        />
        <p className="text-xs text-muted-foreground">
          Optional JSON object. Values must be strings.
        </p>
      </div>

      <fieldset className="space-y-2" disabled={disabled}>
        <legend className="text-sm font-medium">Allowed Models</legend>
        <p className="text-xs text-muted-foreground">
          Leave all unchecked for unrestricted model access.
        </p>
        <div className="max-h-40 space-y-2 overflow-y-auto rounded-md border p-3">
          {modelOptions.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No models are configured. This key will be unrestricted.
            </p>
          ) : (
            modelOptions.map((model) => (
              <label
                key={model.id}
                className="flex items-start gap-2 text-sm"
                htmlFor={`${idPrefix}-model-${model.id}`}
              >
                <Checkbox
                  id={`${idPrefix}-model-${model.id}`}
                  name="allowed_models"
                  value={model.id}
                  defaultChecked={selectedModels.includes(model.id)}
                />
                <span>
                  {model.label}
                  {!model.isAvailable ? (
                    <span className="ml-1 text-xs text-muted-foreground">
                      preserved
                    </span>
                  ) : null}
                </span>
              </label>
            ))
          )}
        </div>
      </fieldset>
    </>
  );
}
