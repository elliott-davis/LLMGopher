import { NextResponse } from "next/server";

const GATEWAY_VALIDATE_ENDPOINT = "http://gateway:8080/api/v1/credentials/validate";

type ValidateCredentialBody = {
  provider?: string;
  apiKey?: string;
};

export async function POST(request: Request) {
  let payload: ValidateCredentialBody;
  try {
    payload = (await request.json()) as ValidateCredentialBody;
  } catch {
    return NextResponse.json(
      { valid: false, code: "invalid_request", error: "Invalid JSON payload" },
      { status: 400 }
    );
  }

  const provider = payload.provider?.trim().toLowerCase() ?? "";
  const apiKey = payload.apiKey?.trim() ?? "";
  if (!provider || !apiKey) {
    return NextResponse.json(
      { valid: false, code: "invalid_request", error: "provider and apiKey are required" },
      { status: 400 }
    );
  }

  try {
    const response = await fetch(GATEWAY_VALIDATE_ENDPOINT, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ provider, apiKey }),
      cache: "no-store",
    });

    let body: unknown = null;
    try {
      body = (await response.json()) as unknown;
    } catch {
      body = null;
    }

    return NextResponse.json(
      body ??
        (response.ok
          ? { valid: true }
          : { valid: false, code: "provider_error", error: "Validation failed" }),
      { status: response.status }
    );
  } catch {
    return NextResponse.json(
      { valid: false, code: "network_error", error: "Network Error" },
      { status: 503 }
    );
  }
}
