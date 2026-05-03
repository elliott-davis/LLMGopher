import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import AuditLogTable from "@/components/usage/AuditLogTable";

describe("AuditLogTable", () => {
  it("renders audit rows, error styling fields, streaming, and pagination links", () => {
    render(
      <AuditLogTable
        rows={[
          {
            id: 1,
            request_id: "req-1",
            api_key_id: "key-001",
            model: "gpt-4o",
            provider: "openai",
            prompt_tokens: 100,
            output_tokens: 50,
            total_tokens: 150,
            cost_usd: 0.01,
            status_code: 500,
            latency_ms: 2500,
            streaming: true,
            error_message: "provider timeout",
            created_at: "2026-04-01T00:00:00Z",
          },
        ]}
        total={100}
        limit={25}
        offset={25}
        filters={{
          group_by: "model",
          limit: 25,
          offset: 25,
          status: "error",
        }}
      />
    );

    expect(screen.getByText("req-1")).toBeInTheDocument();
    expect(screen.getByText("500")).toBeInTheDocument();
    expect(screen.getByText("yes")).toBeInTheDocument();
    expect(screen.getByText("provider timeout")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Previous" })).toHaveAttribute(
      "href",
      expect.stringContaining("offset=0")
    );
    expect(screen.getByRole("link", { name: "Next" })).toHaveAttribute(
      "href",
      expect.stringContaining("offset=50")
    );
  });
});
