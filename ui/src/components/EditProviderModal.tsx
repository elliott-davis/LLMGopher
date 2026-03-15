"use client";

import { useRouter } from "next/navigation";
import { useEffect, useRef, useState, useTransition } from "react";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";

import { updateProvider } from "@/lib/actions";
import { Provider } from "@/lib/types";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
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

type ProviderURLPreset = "openai" | "anthropic" | "vertex" | "custom";

type EditProviderModalProps = {
  provider: Provider;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

function inferPreset(baseURL: string): ProviderURLPreset {
  if (baseURL === BASE_URL_PRESETS.openai) {
    return "openai";
  }
  if (baseURL === BASE_URL_PRESETS.anthropic) {
    return "anthropic";
  }
  if (baseURL === BASE_URL_PRESETS.vertex) {
    return "vertex";
  }
  return "custom";
}

export default function EditProviderModal({
  provider,
  open,
  onOpenChange,
}: EditProviderModalProps) {
  const router = useRouter();
  const formRef = useRef<HTMLFormElement>(null);
  const [isPending, startTransition] = useTransition();
  const [isValidatingCredential, setIsValidatingCredential] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [authType, setAuthType] = useState(provider.auth_type);
  const [baseURL, setBaseURL] = useState(provider.base_url);
  const [urlPreset, setURLPreset] = useState<ProviderURLPreset>(inferPreset(provider.base_url));
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
      // Credential is optional during edits; blank means keep existing value.
      return true;
    }

    const provider = resolveProviderForValidation(
      urlPreset,
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

  useEffect(() => {
    if (open) {
      setAuthType(provider.auth_type);
      setBaseURL(provider.base_url);
      setURLPreset(inferPreset(provider.base_url));
      setErrorMessage(null);
    }
  }, [open, provider.auth_type, provider.base_url]);

  const handleAuthTypeChange = (value: string | null) => {
    if (value == null) return;
    setAuthType(value);
    if (value === "vertex_service_account" && urlPreset !== "custom") {
      setURLPreset("vertex");
      setBaseURL(BASE_URL_PRESETS.vertex);
    }
  };

  const handlePresetChange = (value: ProviderURLPreset | null) => {
    if (value == null) return;
    setURLPreset(value);
    if (value !== "custom") {
      setBaseURL(BASE_URL_PRESETS[value]);
    }
  };

  const handleSubmit = (formData: FormData) => {
    setErrorMessage(null);

    startTransition(() => {
      void (async () => {
        try {
          const isValid = await runCredentialValidation(formData);
          if (!isValid) {
            return;
          }
          await updateProvider(provider.id, formData);
          toast.success("Provider updated");
          onOpenChange(false);
          router.refresh();
        } catch (error) {
          const message =
            error instanceof Error ? error.message : "Failed to update provider";
          setErrorMessage(message);
          toast.error(message);
        }
      })();
    });
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        onOpenChange(nextOpen);
        if (!nextOpen) {
          setErrorMessage(null);
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Provider</DialogTitle>
          <DialogDescription>
            Update provider settings and rotate credentials when needed.
          </DialogDescription>
        </DialogHeader>

        <form ref={formRef} action={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-provider-name">Name</Label>
            <Input
              id="edit-provider-name"
              name="name"
              required
              defaultValue={provider.name}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-provider-url-preset">Provider URL Preset</Label>
            <Select value={urlPreset} onValueChange={handlePresetChange}>
              <SelectTrigger id="edit-provider-url-preset" className="w-full">
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
            <Label htmlFor="edit-provider-base-url">Base URL</Label>
            <Input
              id="edit-provider-base-url"
              name="base_url"
              type="url"
              required
              value={baseURL}
              onChange={(event) => {
                const next = event.target.value;
                setBaseURL(next);
                setURLPreset(inferPreset(next));
              }}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-provider-auth-type">Auth Type</Label>
            <Select
              name="auth_type"
              value={authType}
              onValueChange={handleAuthTypeChange}
              required
            >
              <SelectTrigger id="edit-provider-auth-type" className="w-full">
                <SelectValue placeholder="Select auth type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="bearer">Bearer token</SelectItem>
                <SelectItem value="vertex_service_account">
                  Vertex service account JSON
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {requiresBearerToken ? (
            <div className="space-y-2">
              <Label htmlFor="edit-provider-token">Bearer Token (optional)</Label>
              <Input
                id="edit-provider-token"
                name="credential_token"
                type="password"
                placeholder="Leave blank to keep existing credentials"
              />
            </div>
          ) : null}

          {requiresServiceAccount ? (
            <div className="space-y-2">
              <Label htmlFor="edit-provider-credential-file">
                Vertex Service Account JSON (optional)
              </Label>
              <Input
                id="edit-provider-credential-file"
                name="credential_file"
                type="file"
                accept=".json,application/json"
              />
            </div>
          ) : null}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isPending || isValidatingCredential}
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={handleTestConnection}
              disabled={isPending || isValidatingCredential}
            >
              {isValidatingCredential ? "Testing..." : "Test Connection"}
            </Button>
            <Button type="submit" disabled={isPending || isValidatingCredential}>
              {isPending || isValidatingCredential ? "Saving..." : "Save changes"}
            </Button>
          </DialogFooter>
          {isValidatingCredential ? (
            <p className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="size-4 animate-spin" aria-hidden />
              Validating credentials...
            </p>
          ) : null}
          {errorMessage ? (
            <p className="text-sm text-destructive">{errorMessage}</p>
          ) : null}
        </form>
      </DialogContent>
    </Dialog>
  );
}
