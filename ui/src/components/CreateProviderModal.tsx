"use client";

import { useRouter } from "next/navigation";
import { useRef, useState, useTransition } from "react";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";

import { createProvider, waitForProviderCreationSync } from "@/lib/actions";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  CredentialValidationError,
  validateCredential,
  type SupportedCredentialProvider,
} from "@/services/validation";

const BASE_URL_PRESETS: Record<string, string> = {
  openai: "https://api.openai.com/v1",
  anthropic: "https://api.anthropic.com",
  vertex: "https://us-central1-aiplatform.googleapis.com/v1beta1",
};

export default function CreateProviderModal() {
  const router = useRouter();
  const formRef = useRef<HTMLFormElement>(null);
  const [open, setOpen] = useState(false);
  const [isPending, startTransition] = useTransition();
  const [isWaitingForSync, setIsWaitingForSync] = useState(false);
  const [isValidatingCredential, setIsValidatingCredential] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [authType, setAuthType] = useState("bearer");
  const [urlPreset, setURLPreset] = useState<"openai" | "anthropic" | "vertex" | "custom">("openai");
  const [baseURL, setBaseURL] = useState(BASE_URL_PRESETS.openai);
  const requiresServiceAccount = authType === "vertex_service_account";
  const requiresBearerToken = authType === "bearer";

  const resolveProviderForValidation = (
    presetValue: string,
    baseURLValue: string
  ): SupportedCredentialProvider | null => {
    if (presetValue === "openai") return "openai";
    if (presetValue === "anthropic") return "anthropic";
    if (presetValue === "vertex") return "google";

    const normalizedBaseURL = baseURLValue.toLowerCase();
    if (normalizedBaseURL.includes("api.openai.com")) return "openai";
    if (normalizedBaseURL.includes("api.anthropic.com")) return "anthropic";
    if (
      normalizedBaseURL.includes("generativelanguage.googleapis.com") ||
      normalizedBaseURL.includes("googleapis.com")
    ) {
      return "google";
    }

    return null;
  };

  const runCredentialValidation = async (formData: FormData): Promise<boolean> => {
    const selectedAuthType = String(formData.get("auth_type") ?? "").trim();
    if (selectedAuthType !== "bearer") {
      return true;
    }

    const apiKey = String(formData.get("credential_token") ?? "").trim();
    if (!apiKey) {
      const message = "Invalid API Key";
      setErrorMessage(message);
      toast.error(message);
      return false;
    }

    const provider = resolveProviderForValidation(
      String(formData.get("url_preset") ?? "").trim(),
      String(formData.get("base_url") ?? "").trim()
    );
    if (!provider) {
      return true;
    }

    setIsValidatingCredential(true);
    try {
      await validateCredential(provider, apiKey);
      toast.success("Connection successful");
      return true;
    } catch (error) {
      const message =
        error instanceof CredentialValidationError
          ? error.message
          : "Validation failed";
      setErrorMessage(message);
      toast.error(message);
      return false;
    } finally {
      setIsValidatingCredential(false);
    }
  };

  const handleTestConnection = () => {
    if (!formRef.current) {
      return;
    }
    setErrorMessage(null);
    const formData = new FormData(formRef.current);
    void runCredentialValidation(formData);
  };

  const handleAuthTypeChange = (value: string | null) => {
    if (value == null) return;
    setAuthType(value);
    if (value === "vertex_service_account" && urlPreset !== "custom") {
      setURLPreset("vertex");
      setBaseURL(BASE_URL_PRESETS.vertex);
    }
  };

  const handlePresetChange = (value: "openai" | "anthropic" | "vertex" | "custom" | null) => {
    if (value == null) return;
    setURLPreset(value);
    if (value !== "custom") {
      setBaseURL(BASE_URL_PRESETS[value]);
    }
  };

  const handleSubmit = (formData: FormData) => {
    setErrorMessage(null);
    const providerName = String(formData.get("name") ?? "").trim();

    startTransition(() => {
      void (async () => {
        try {
          const isValid = await runCredentialValidation(formData);
          if (!isValid) {
            return;
          }

          await createProvider(formData);
          setIsWaitingForSync(true);
          toast.message("Provider created. Waiting for cache sync...");

          const synced = await waitForProviderCreationSync(providerName, 30000, 1000);
          router.refresh();

          if (!synced) {
            toast.warning("Provider created, but cache sync is taking longer than expected.");
            return;
          }

          toast.success("Provider created");
          setOpen(false);
        } catch (error) {
          const message = error instanceof Error ? error.message : "Failed to create provider";
          setErrorMessage(message);
          toast.error(message);
        } finally {
          setIsWaitingForSync(false);
        }
      })();
    });
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (isPending || isWaitingForSync) {
          return;
        }
        setOpen(nextOpen);
        if (!nextOpen) {
          setErrorMessage(null);
        }
      }}
    >
      <DialogTrigger render={<Button />}>Add Provider</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Provider</DialogTitle>
          <DialogDescription>
            Create a reusable provider instance that models can reference.
          </DialogDescription>
        </DialogHeader>

        <form ref={formRef} action={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="provider-name">Name</Label>
            <Input id="provider-name" name="name" required placeholder="OpenAI Production" />
          </div>

          <div className="space-y-2">
            <Label htmlFor="provider-url-preset">Provider URL Preset</Label>
            <Select name="url_preset" value={urlPreset} onValueChange={handlePresetChange}>
              <SelectTrigger id="provider-url-preset" className="w-full">
                <SelectValue placeholder="Select provider URL preset" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="openai">OpenAI Cloud</SelectItem>
                <SelectItem value="anthropic">Anthropic Cloud</SelectItem>
                <SelectItem value="vertex">Google Vertex AI</SelectItem>
                <SelectItem value="custom">Custom URL</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="provider-base-url">Base URL</Label>
            <Input
              id="provider-base-url"
              name="base_url"
              type="url"
              required
              value={baseURL}
              onChange={(event) => {
                const next = event.target.value;
                setBaseURL(next);
                if (next === BASE_URL_PRESETS.openai) {
                  setURLPreset("openai");
                } else if (next === BASE_URL_PRESETS.anthropic) {
                  setURLPreset("anthropic");
                } else if (next === BASE_URL_PRESETS.vertex) {
                  setURLPreset("vertex");
                } else {
                  setURLPreset("custom");
                }
              }}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="provider-auth-type">Auth Type</Label>
            <Select name="auth_type" value={authType} onValueChange={handleAuthTypeChange} required>
              <SelectTrigger id="provider-auth-type" className="w-full">
                <SelectValue placeholder="Select auth type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="bearer">Bearer token</SelectItem>
                <SelectItem value="vertex_service_account">Vertex service account JSON</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {requiresBearerToken ? (
            <div className="space-y-2">
              <Label htmlFor="provider-token">Bearer Token</Label>
              <Input
                id="provider-token"
                name="credential_token"
                type="password"
                required
                placeholder="sk-live-..."
              />
            </div>
          ) : null}

          {requiresServiceAccount ? (
            <div className="space-y-2">
              <Label htmlFor="provider-credential-file">Vertex Service Account JSON</Label>
              <Input
                id="provider-credential-file"
                name="credential_file"
                type="file"
                accept=".json,application/json"
                required
              />
            </div>
          ) : null}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={handleTestConnection}
              disabled={isPending || isWaitingForSync || isValidatingCredential}
            >
              {isValidatingCredential ? "Testing..." : "Test Connection"}
            </Button>
            <Button
              type="submit"
              disabled={isPending || isWaitingForSync || isValidatingCredential}
            >
              {isPending || isWaitingForSync || isValidatingCredential
                ? "Saving..."
                : "Create Provider"}
            </Button>
          </DialogFooter>
          {isWaitingForSync ? (
            <p className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="size-4 animate-spin" aria-hidden />
              Saving provider and waiting for cache sync...
            </p>
          ) : null}
          {isValidatingCredential ? (
            <p className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="size-4 animate-spin" aria-hidden />
              Validating credentials...
            </p>
          ) : null}
          {errorMessage ? <p className="text-sm text-destructive">{errorMessage}</p> : null}
        </form>
      </DialogContent>
    </Dialog>
  );
}
