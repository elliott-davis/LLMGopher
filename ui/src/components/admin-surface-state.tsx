import * as React from "react";
import { Button } from "@/components/ui/button";

export function LoadingState() {
  return (
    <div data-testid="surface-loading" aria-busy="true">
      <p>Loading…</p>
    </div>
  );
}

export interface EmptyStateProps {
  message?: string;
  onClear?: () => void;
}

export function EmptyState({ message = "No items found.", onClear }: EmptyStateProps) {
  return (
    <div data-testid="surface-empty">
      <p>{message}</p>
      {onClear && (
        <Button variant="ghost" size="sm" data-testid="clear-filters" onClick={onClear}>
          Clear filters
        </Button>
      )}
    </div>
  );
}

export interface UnavailableStateProps {
  message?: string;
}

export function UnavailableState({ message }: UnavailableStateProps) {
  return (
    <div data-testid="surface-unavailable">
      <p>Service unavailable.</p>
      {message && <p>{message}</p>}
    </div>
  );
}

export interface ErrorStateProps {
  message: string;
}

export function ErrorState({ message }: ErrorStateProps) {
  return (
    <div data-testid="surface-error">
      <p>{message}</p>
    </div>
  );
}
