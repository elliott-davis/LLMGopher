"use client";

import { useState, useTransition, useRef, type RefObject } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { createProvider, waitForProviderCreationSync } from "@/lib/actions";
import {
  validateCredential,
  CredentialValidationError,
  type SupportedCredentialProvider,
} from "@/services/validation";
import { ProviderLogo } from "./ProviderLogo";

const KINDS = [
  {
    id: "openai",
    label: "OpenAI",
    desc: "OpenAI API · gpt-4o family",
    baseUrl: "https://api.openai.com/v1",
    authType: "bearer" as const,
    validationKey: "openai" as SupportedCredentialProvider,
  },
  {
    id: "anthropic",
    label: "Anthropic",
    desc: "Anthropic API · claude family",
    baseUrl: "https://api.anthropic.com",
    authType: "bearer" as const,
    validationKey: "anthropic" as SupportedCredentialProvider,
  },
  {
    id: "vertex",
    label: "Vertex AI",
    desc: "Google Vertex · gemini family",
    baseUrl: "https://us-central1-aiplatform.googleapis.com/v1beta1",
    authType: "vertex_service_account" as const,
    validationKey: null,
  },
  {
    id: "bedrock",
    label: "AWS Bedrock",
    desc: "Bedrock · multi-tenant",
    baseUrl: "https://bedrock-runtime.us-east-1.amazonaws.com",
    authType: "bearer" as const,
    validationKey: null,
  },
  {
    id: "mistral",
    label: "Mistral",
    desc: "Mistral API · EU residency",
    baseUrl: "https://api.mistral.ai/v1",
    authType: "bearer" as const,
    validationKey: null,
  },
  {
    id: "vllm",
    label: "vLLM (self-hosted)",
    desc: "OpenAI-compatible endpoint",
    baseUrl: "http://localhost:8000/v1",
    authType: "bearer" as const,
    validationKey: null,
  },
] as const;

type KindId = (typeof KINDS)[number]["id"];

function CloseIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <path d="M18 6 6 18M6 6l12 12" />
    </svg>
  );
}

function ZapIcon() {
  return (
    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <path d="M13 2 3 14h9l-1 8 10-12h-9z" />
    </svg>
  );
}

// ── Step 1 ──────────────────────────────────────────────────────────────────
function TypeStep({
  kind,
  onSelect,
}: {
  kind: KindId;
  onSelect: (id: KindId) => void;
}) {
  return (
    <>
      <h4 style={{ margin: "0 0 14px", fontSize: 13, fontWeight: 600 }}>
        Choose a provider type
      </h4>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10 }}>
        {KINDS.map((k) => (
          <div
            key={k.id}
            role="button"
            tabIndex={0}
            data-testid={`provider-kind-${k.id}`}
            onClick={() => onSelect(k.id)}
            onKeyDown={(e) => e.key === "Enter" && onSelect(k.id)}
            className="card"
            style={{
              padding: 14,
              cursor: "pointer",
              borderColor: kind === k.id ? "var(--brand)" : "var(--border)",
              boxShadow:
                kind === k.id
                  ? "0 0 0 3px color-mix(in oklch, var(--brand) 20%, transparent), var(--shadow-1)"
                  : "var(--shadow-1)",
            }}
          >
            <div className="row" style={{ gap: 10 }}>
              <ProviderLogo name={k.id} size="lg" />
              <div>
                <div style={{ fontWeight: 600, fontSize: 13 }}>{k.label}</div>
                <div className="subtle" style={{ fontSize: 12, marginTop: 2 }}>
                  {k.desc}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}

// ── Step 2 ──────────────────────────────────────────────────────────────────
type ConnStatus = "idle" | "testing" | "ok" | "error";

function CredentialsStep({
  authType,
  name,
  setName,
  baseUrl,
  setBaseUrl,
  credentialToken,
  setCredentialToken,
  fileInputRef,
  connStatus,
  connMsg,
  onTest,
}: {
  authType: string;
  name: string;
  setName: (v: string) => void;
  baseUrl: string;
  setBaseUrl: (v: string) => void;
  credentialToken: string;
  setCredentialToken: (v: string) => void;
  fileInputRef: RefObject<HTMLInputElement | null>;
  connStatus: ConnStatus;
  connMsg: string;
  onTest: () => void;
}) {
  const isBearer = authType === "bearer";

  return (
    <>
      <h4 style={{ margin: "0 0 14px", fontSize: 13, fontWeight: 600 }}>
        Credentials &amp; endpoint
      </h4>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
        <div className="field" style={{ gridColumn: "1 / -1" }}>
          <label htmlFor="prov-name">Display name</label>
          <input
            id="prov-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="OpenAI · production"
          />
        </div>
        <div className="field" style={{ gridColumn: "1 / -1" }}>
          <label htmlFor="prov-url">Base URL</label>
          <input
            id="prov-url"
            className="mono"
            value={baseUrl}
            onChange={(e) => setBaseUrl(e.target.value)}
          />
        </div>
        {isBearer ? (
          <div className="field" style={{ gridColumn: "1 / -1" }}>
            <label htmlFor="prov-token">API key (encrypted at rest)</label>
            <input
              id="prov-token"
              type="password"
              value={credentialToken}
              onChange={(e) => setCredentialToken(e.target.value)}
              placeholder="sk-…"
            />
          </div>
        ) : (
          <div className="field" style={{ gridColumn: "1 / -1" }}>
            <label htmlFor="prov-file">Service account JSON</label>
            <input
              id="prov-file"
              type="file"
              accept=".json,application/json"
              ref={fileInputRef as RefObject<HTMLInputElement>}
            />
          </div>
        )}
      </div>

      {isBearer && (
        <>
          <div style={{ height: 14 }} />
          <button
            type="button"
            className="btn sm"
            onClick={onTest}
            disabled={connStatus === "testing" || !credentialToken}
          >
            {connStatus === "testing" ? "Testing…" : "Test connection"}
          </button>
        </>
      )}

      {connStatus !== "idle" && (
        <div
          className="card"
          style={{
            marginTop: 14,
            background: "var(--bg-sunken)",
            borderStyle: connStatus === "ok" ? "dashed" : "solid",
            borderColor:
              connStatus === "ok" ? "var(--border)" : "var(--danger)",
          }}
        >
          <div className="card-body" style={{ padding: "10px 14px" }}>
            <div className="row" style={{ gap: 10 }}>
              <span
                className={`pill ${connStatus === "ok" ? "success" : "danger"}`}
              >
                <span className="dot" />
                {connStatus === "ok" ? "connection ok" : "connection failed"}
              </span>
              <span className="subtle" style={{ fontSize: 12 }}>{connMsg}</span>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

// ── Step 3 ──────────────────────────────────────────────────────────────────
function ConfirmStep({
  name,
  baseUrl,
  kindId,
}: {
  name: string;
  baseUrl: string;
  kindId: KindId;
}) {
  const k = KINDS.find((k) => k.id === kindId)!;
  return (
    <>
      <h4 style={{ margin: "0 0 14px", fontSize: 13, fontWeight: 600 }}>
        Review &amp; create
      </h4>
      <div className="card" style={{ background: "var(--bg-sunken)" }}>
        <div
          className="card-body"
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 12,
            padding: "14px 18px",
          }}
        >
          {[
            { label: "Display name", value: name },
            { label: "Provider type", value: k.label },
            { label: "Base URL", value: baseUrl, mono: true },
            {
              label: "Auth",
              value:
                k.authType === "bearer" ? "Bearer token" : "Service account JSON",
            },
          ].map(({ label, value, mono }) => (
            <div
              key={label}
              className="row between"
              style={{ fontSize: 13, paddingBottom: 12, borderBottom: "1px solid var(--border)" }}
            >
              <span style={{ color: "var(--fg-muted)", minWidth: 130 }}>
                {label}
              </span>
              <span
                className={mono ? "mono" : ""}
                style={{ fontWeight: 500, fontSize: mono ? 12 : 13, textAlign: "right" }}
              >
                {value || <span className="subtle">—</span>}
              </span>
            </div>
          ))}
        </div>
      </div>
      <p className="subtle" style={{ marginTop: 14, fontSize: 12.5, lineHeight: 1.5 }}>
        After creation, add models from the <strong>Models</strong> page by selecting this provider. The provider will be available immediately after cache sync (~5 s).
      </p>
    </>
  );
}

// ── Main drawer ──────────────────────────────────────────────────────────────
interface AddProviderDrawerProps {
  open: boolean;
  onClose: () => void;
}

export function AddProviderDrawer({ open, onClose }: AddProviderDrawerProps) {
  const router = useRouter();
  const [step, setStep] = useState(0);
  const [kindId, setKindId] = useState<KindId>("openai");
  const [name, setName] = useState("OpenAI · production");
  const [baseUrl, setBaseUrl] = useState<string>(KINDS[0].baseUrl);
  const [authType, setAuthType] = useState<string>(KINDS[0].authType);
  const [credentialToken, setCredentialToken] = useState("");
  const [connStatus, setConnStatus] = useState<ConnStatus>("idle");
  const [connMsg, setConnMsg] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isPending, startTransition] = useTransition();
  const [isWaitingSync, setIsWaitingSync] = useState(false);

  if (!open) return null;

  const busy = isPending || isWaitingSync;

  const handleSelectKind = (id: KindId) => {
    const k = KINDS.find((k) => k.id === id)!;
    setKindId(id);
    setBaseUrl(k.baseUrl);
    setAuthType(k.authType);
    setName(`${k.label} · production`);
    setConnStatus("idle");
    setCredentialToken("");
  };

  const handleTestConnection = async () => {
    const k = KINDS.find((k) => k.id === kindId)!;
    if (!k.validationKey) {
      setConnStatus("ok");
      setConnMsg("Manual verification required for this provider type.");
      return;
    }
    setConnStatus("testing");
    try {
      await validateCredential(k.validationKey, credentialToken);
      setConnStatus("ok");
      setConnMsg("Credentials verified.");
    } catch (err) {
      setConnStatus("error");
      setConnMsg(
        err instanceof CredentialValidationError
          ? err.message
          : "Connection failed."
      );
    }
  };

  const canAdvanceStep2 =
    name.trim() !== "" &&
    baseUrl.trim() !== "" &&
    (authType !== "bearer" || credentialToken.trim() !== "");

  const handleContinue = () => {
    if (step === 1 && !canAdvanceStep2) {
      setErrorMsg("Please fill in all required fields.");
      return;
    }
    setErrorMsg(null);
    setStep((s) => s + 1);
  };

  const handleClose = () => {
    if (busy) return;
    setStep(0);
    setKindId("openai");
    setName("OpenAI · production");
    setBaseUrl(KINDS[0].baseUrl);
    setAuthType(KINDS[0].authType);
    setCredentialToken("");
    setConnStatus("idle");
    setErrorMsg(null);
    onClose();
  };

  const handleCreate = () => {
    setErrorMsg(null);
    startTransition(() => {
      void (async () => {
        try {
          const fd = new FormData();
          fd.append("name", name.trim());
          fd.append("base_url", baseUrl.trim());
          fd.append("auth_type", authType);
          if (authType === "bearer") {
            fd.append("credential_token", credentialToken.trim());
          }
          const file = fileInputRef.current?.files?.[0];
          if (file) fd.append("credential_file", file);

          await createProvider(fd);
          setIsWaitingSync(true);
          toast.message("Provider created — waiting for cache sync…");
          await waitForProviderCreationSync(name.trim(), 30_000, 1_000);
          router.refresh();
          toast.success("Provider created");
          handleClose();
        } catch (err) {
          const msg =
            err instanceof Error ? err.message : "Failed to create provider";
          setErrorMsg(msg);
          toast.error(msg);
        } finally {
          setIsWaitingSync(false);
        }
      })();
    });
  };

  return (
    <div
      className="drawer-mask"
      role="dialog"
      aria-modal="true"
      aria-label="Add provider"
      onClick={(e) => {
        if (e.target === e.currentTarget) handleClose();
      }}
    >
      <div className="drawer" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="card-head" style={{ flexShrink: 0 }}>
          <h3>Add provider</h3>
          <span className="sub">Step {step + 1} of 3</span>
          <div className="right">
            <button
              className="icon-btn"
              onClick={handleClose}
              aria-label="Close"
            >
              <CloseIcon />
            </button>
          </div>
        </div>

        {/* Body */}
        <div style={{ flex: 1, overflow: "auto", padding: 22 }}>
          {step === 0 && (
            <TypeStep kind={kindId} onSelect={handleSelectKind} />
          )}
          {step === 1 && (
            <CredentialsStep
              authType={authType}
              name={name}
              setName={setName}
              baseUrl={baseUrl}
              setBaseUrl={setBaseUrl}
              credentialToken={credentialToken}
              setCredentialToken={setCredentialToken}
              fileInputRef={fileInputRef}
              connStatus={connStatus}
              connMsg={connMsg}
              onTest={handleTestConnection}
            />
          )}
          {step === 2 && (
            <ConfirmStep name={name} baseUrl={baseUrl} kindId={kindId} />
          )}

          {errorMsg && (
            <div
              style={{
                marginTop: 14,
                padding: "8px 12px",
                background: "var(--danger-soft)",
                border: "1px solid color-mix(in oklch, var(--danger) 25%, transparent)",
                borderRadius: "var(--r-sm)",
                color: "var(--danger)",
                fontSize: 13,
              }}
            >
              {errorMsg}
            </div>
          )}
        </div>

        {/* Footer */}
        <div
          className="row between"
          style={{
            padding: "12px 16px",
            borderTop: "1px solid var(--border)",
            flexShrink: 0,
          }}
        >
          <button className="btn ghost" onClick={handleClose} disabled={busy}>
            Cancel
          </button>
          <div className="row" style={{ gap: 8 }}>
            {step > 0 && (
              <button
                className="btn"
                onClick={() => setStep((s) => s - 1)}
                disabled={busy}
              >
                Back
              </button>
            )}
            {step < 2 && (
              <button className="btn primary" data-testid="wizard-next" onClick={handleContinue}>
                Continue
              </button>
            )}
            {step === 2 && (
              <button
                className="btn primary"
                data-testid="wizard-create"
                onClick={handleCreate}
                disabled={busy}
              >
                <ZapIcon />
                {busy ? "Creating…" : "Create provider"}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
