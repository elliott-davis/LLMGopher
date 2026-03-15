"use server";

import { revalidatePath } from "next/cache";

const GATEWAY_BASE = "http://gateway:8080";
const CREATE_MODEL_ENDPOINT = `${GATEWAY_BASE}/v1/admin/models`;
const GET_MODELS_ENDPOINT = `${GATEWAY_BASE}/v1/admin/models`;
const CREATE_PROVIDER_ENDPOINT = `${GATEWAY_BASE}/v1/admin/providers`;
const GET_PROVIDERS_ENDPOINT = `${GATEWAY_BASE}/v1/admin/providers`;
const CREATE_API_KEY_ENDPOINT = "http://gateway:8080/v1/admin/keys";

export async function createModel(formData: FormData) {
  const alias = String(formData.get("alias") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const providerID = String(formData.get("provider_id") ?? "").trim();
  const contextWindow = Number(String(formData.get("context_window") ?? "").trim());

  if (!alias || !name || !providerID || !Number.isFinite(contextWindow) || contextWindow <= 0) {
    throw new Error("Invalid model form data");
  }

  const response = await fetch(CREATE_MODEL_ENDPOINT, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      alias,
      name,
      provider_id: providerID,
      context_window: contextWindow,
    }),
  });

  if (!response.ok) {
    throw new Error("Failed to create model");
  }

  revalidatePath("/models");
}

export async function updateModel(id: string, formData: FormData) {
  const alias = String(formData.get("alias") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const providerID = String(formData.get("provider_id") ?? "").trim();
  const contextWindow = Number(String(formData.get("context_window") ?? "").trim());

  if (!id || !alias || !name || !providerID || !Number.isFinite(contextWindow) || contextWindow <= 0) {
    throw new Error("Invalid model form data");
  }

  const response = await fetch(`${GATEWAY_BASE}/v1/admin/models/${encodeURIComponent(id)}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      alias,
      name,
      provider_id: providerID,
      context_window: contextWindow,
    }),
  });

  if (!response.ok) {
    throw new Error("Failed to update model");
  }

  revalidatePath("/models");
}

export async function deleteModel(id: string) {
  if (!id) {
    throw new Error("Model id is required");
  }

  const response = await fetch(`${GATEWAY_BASE}/v1/admin/models/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });

  if (!response.ok) {
    throw new Error("Failed to delete model");
  }

  revalidatePath("/models");
}

type GatewayModel = {
  id: string;
};

type GatewayProvider = {
  id: string;
  name: string;
};

function sleep(ms: number) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

export async function waitForModelDeletionSync(
  id: string,
  timeoutMs = 30000,
  intervalMs = 1000
): Promise<boolean> {
  if (!id) {
    throw new Error("Model id is required");
  }
  if (timeoutMs <= 0 || intervalMs <= 0) {
    throw new Error("Invalid polling configuration");
  }

  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const response = await fetch(GET_MODELS_ENDPOINT, {
      cache: "no-store",
    });
    if (!response.ok) {
      throw new Error("Failed to fetch models while waiting for deletion sync");
    }

    const models = (await response.json()) as GatewayModel[];
    const stillExists = models.some((model) => model.id === id);
    if (!stillExists) {
      revalidatePath("/models");
      return true;
    }

    await sleep(intervalMs);
  }

  return false;
}

export async function createProvider(formData: FormData) {
  const name = String(formData.get("name") ?? "").trim();
  const baseURL = String(formData.get("base_url") ?? "").trim();
  const authType = String(formData.get("auth_type") ?? "").trim();
  const credentialToken = String(formData.get("credential_token") ?? "").trim();
  const credentialFile = formData.get("credential_file");

  if (!name || !baseURL || !authType) {
    throw new Error("Invalid provider form data");
  }
  if (authType === "vertex_service_account" && !(credentialFile instanceof File && credentialFile.size > 0)) {
    throw new Error("Vertex providers require a service account JSON file");
  }
  if (authType === "bearer" && !credentialToken) {
    throw new Error("Bearer providers require a token");
  }

  const payload = new FormData();
  payload.append("name", name);
  payload.append("base_url", baseURL);
  payload.append("auth_type", authType);
  if (credentialToken) {
    payload.append("credential_token", credentialToken);
  }
  if (credentialFile instanceof File && credentialFile.size > 0) {
    payload.append("credential_file", credentialFile);
  }

  const response = await fetch(CREATE_PROVIDER_ENDPOINT, {
    method: "POST",
    body: payload,
  });

  if (!response.ok) {
    throw new Error("Failed to create provider");
  }

  revalidatePath("/providers");
  revalidatePath("/models");
}

export async function waitForProviderCreationSync(
  name: string,
  timeoutMs = 30000,
  intervalMs = 1000
): Promise<boolean> {
  const providerName = name.trim();
  if (!providerName) {
    throw new Error("Provider name is required");
  }
  if (timeoutMs <= 0 || intervalMs <= 0) {
    throw new Error("Invalid polling configuration");
  }

  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const response = await fetch(GET_PROVIDERS_ENDPOINT, {
      cache: "no-store",
    });
    if (!response.ok) {
      throw new Error("Failed to fetch providers while waiting for creation sync");
    }

    const providers = (await response.json()) as GatewayProvider[];
    const found = providers.some((provider) => provider.name === providerName);
    if (found) {
      revalidatePath("/providers");
      return true;
    }

    await sleep(intervalMs);
  }

  return false;
}

function extractGatewayErrorMessage(payload: unknown, fallback: string): string {
  if (typeof payload !== "object" || payload === null) {
    return fallback;
  }

  const maybeError = (payload as { error?: unknown }).error;
  if (typeof maybeError === "string" && maybeError.trim()) {
    return maybeError;
  }

  const maybeMessage = (payload as { message?: unknown }).message;
  if (typeof maybeMessage === "string" && maybeMessage.trim()) {
    return maybeMessage;
  }

  return fallback;
}

export async function updateProvider(id: string, formData: FormData) {
  const name = String(formData.get("name") ?? "").trim();
  const baseURL = String(formData.get("base_url") ?? "").trim();
  const authType = String(formData.get("auth_type") ?? "").trim();
  const credentialToken = String(formData.get("credential_token") ?? "").trim();
  const credentialFile = formData.get("credential_file");

  if (!id || !name || !baseURL || !authType) {
    throw new Error("Invalid provider form data");
  }

  const payload = new FormData();
  payload.append("name", name);
  payload.append("base_url", baseURL);
  payload.append("auth_type", authType);
  if (credentialToken) {
    payload.append("credential_token", credentialToken);
  }
  if (credentialFile instanceof File && credentialFile.size > 0) {
    payload.append("credential_file", credentialFile);
  }

  const response = await fetch(
    `${GET_PROVIDERS_ENDPOINT}/${encodeURIComponent(id)}`,
    {
      method: "PUT",
      body: payload,
    }
  );

  if (!response.ok) {
    let message = "Failed to update provider";
    try {
      const errorPayload = (await response.json()) as unknown;
      message = extractGatewayErrorMessage(errorPayload, message);
    } catch {
      // ignore parsing errors and use fallback message
    }
    throw new Error(message);
  }

  revalidatePath("/providers");
  revalidatePath("/models");
}

type DeleteProviderResult =
  | { success: true }
  | { success: false; error: string };

export async function deleteProvider(id: string): Promise<DeleteProviderResult> {
  if (!id) {
    throw new Error("Provider id is required");
  }

  const response = await fetch(
    `${GET_PROVIDERS_ENDPOINT}/${encodeURIComponent(id)}`,
    {
      method: "DELETE",
    }
  );

  if (response.status === 409) {
    let message = "Cannot delete provider because it is currently in use.";
    try {
      const errorPayload = (await response.json()) as unknown;
      message = extractGatewayErrorMessage(errorPayload, message);
    } catch {
      // ignore parsing errors and use fallback message
    }
    return { success: false, error: message };
  }

  if (!response.ok) {
    throw new Error("Failed to delete provider");
  }

  revalidatePath("/providers");
  return { success: true };
}

export async function waitForProviderDeletionSync(
  id: string,
  timeoutMs = 30000,
  intervalMs = 1000
): Promise<boolean> {
  if (!id) {
    throw new Error("Provider id is required");
  }
  if (timeoutMs <= 0 || intervalMs <= 0) {
    throw new Error("Invalid polling configuration");
  }

  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const response = await fetch(GET_PROVIDERS_ENDPOINT, {
      cache: "no-store",
    });
    if (!response.ok) {
      throw new Error("Failed to fetch providers while waiting for deletion sync");
    }

    const providers = (await response.json()) as GatewayProvider[];
    const stillExists = providers.some((provider) => provider.id === id);
    if (!stillExists) {
      revalidatePath("/providers");
      return true;
    }

    await sleep(intervalMs);
  }

  return false;
}

type CreateAPIKeyResponse = {
  api_key: string;
};

export async function createAPIKey(formData: FormData): Promise<string> {
  const name = String(formData.get("name") ?? "").trim();
  const rateLimitRPS = Number(String(formData.get("rate_limit_rps") ?? "").trim());

  if (!name || !Number.isFinite(rateLimitRPS) || rateLimitRPS < 0) {
    throw new Error("Invalid API key form data");
  }

  const response = await fetch(CREATE_API_KEY_ENDPOINT, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      name,
      rate_limit_rps: rateLimitRPS,
    }),
  });

  if (!response.ok) {
    throw new Error("Failed to create API key");
  }

  const payload = (await response.json()) as CreateAPIKeyResponse;
  if (!payload.api_key) {
    throw new Error("API key was not returned by the gateway");
  }

  revalidatePath("/keys");
  return payload.api_key;
}
