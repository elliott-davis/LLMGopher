import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import ModelRateLimitStatus from "@/components/ModelRateLimitStatus";

describe("ModelRateLimitStatus", () => {
  it("renders positive model-level limits in requests per second", () => {
    render(<ModelRateLimitStatus rateLimitRPS={25} />);

    expect(screen.getByText("25 requests/sec")).toBeInTheDocument();
  });

  it("renders zero and missing values as no model-level limit", () => {
    const { rerender } = render(<ModelRateLimitStatus rateLimitRPS={0} />);

    expect(screen.getByText("No model-level limit")).toBeInTheDocument();

    rerender(<ModelRateLimitStatus />);

    expect(screen.getByText("No model-level limit")).toBeInTheDocument();
  });
});
