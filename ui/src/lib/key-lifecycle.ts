import { Model } from "@/lib/types";

export type ModelAllowlistOption = {
  id: string;
  label: string;
  isAvailable: boolean;
};

export function getModelIdentifier(model: Model): string {
  return model.alias || model.id;
}

export function getModelLabel(model: Model): string {
  const identifier = getModelIdentifier(model);
  const provider = model.provider_id ? ` (${model.provider_id})` : "";
  return model.name && model.name !== identifier
    ? `${identifier} - ${model.name}${provider}`
    : `${identifier}${provider}`;
}

export function buildModelAllowlistOptions(
  models: Model[],
  selectedIds: string[] = []
): ModelAllowlistOption[] {
  const seen = new Set<string>();
  const availableOptions = models.map((model) => {
    const id = getModelIdentifier(model);
    seen.add(id);
    return {
      id,
      label: getModelLabel(model),
      isAvailable: true,
    };
  });

  const staleOptions = selectedIds
    .filter((id) => id && !seen.has(id))
    .map((id) => ({
      id,
      label: `${id} (stale)`,
      isAvailable: false,
    }));

  return [...availableOptions, ...staleOptions].sort((a, b) =>
    a.label.localeCompare(b.label)
  );
}

export function describeModelAllowlist(
  allowedModels: string[] | null | undefined,
  models: Model[]
): string {
  if (!allowedModels || allowedModels.length === 0) {
    return "Unrestricted";
  }

  const labels = buildModelAllowlistOptions(models, allowedModels)
    .filter((option) => allowedModels.includes(option.id))
    .map((option) => option.label);

  return labels.join(", ");
}
