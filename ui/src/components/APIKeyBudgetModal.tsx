"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState, useTransition } from "react";
import { toast } from "sonner";

import APIKeyBudgetForm from "@/components/APIKeyBudgetForm";
import APIKeyBudgetStatus from "@/components/APIKeyBudgetStatus";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  deleteAPIKeyBudget,
  resetAPIKeyBudget,
  upsertAPIKeyBudget,
} from "@/lib/actions";
import { APIKeyBudgetState } from "@/lib/types";

type APIKeyBudgetModalProps = {
  apiKeyID: string;
  apiKeyName: string;
  initialState: APIKeyBudgetState;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export default function APIKeyBudgetModal({
  apiKeyID,
  apiKeyName,
  initialState,
  open,
  onOpenChange,
}: APIKeyBudgetModalProps) {
  const router = useRouter();
  const [budgetState, setBudgetState] = useState<APIKeyBudgetState>(initialState);
  const [isPending, startTransition] = useTransition();

  useEffect(() => {
    setBudgetState(initialState);
  }, [initialState, open]);

  const refreshSoon = () => {
    router.refresh();
    window.setTimeout(() => router.refresh(), 1200);
  };

  const handleSave = async (formData: FormData) => {
    return new Promise<void>((resolve) => {
      startTransition(() => {
        void (async () => {
          try {
            const budget = await upsertAPIKeyBudget(apiKeyID, formData);
            setBudgetState({ status: "configured", budget });
            toast.success("Budget saved");
            refreshSoon();
          } catch (error) {
            toast.error(error instanceof Error ? error.message : "Failed to save budget");
          } finally {
            resolve();
          }
        })();
      });
    });
  };

  const handleReset = () => {
    if (budgetState.status !== "configured") {
      return;
    }
    if (!window.confirm("Reset budget spend to zero?")) {
      return;
    }

    startTransition(() => {
      void (async () => {
        try {
          const budget = await resetAPIKeyBudget(apiKeyID);
          setBudgetState({ status: "configured", budget });
          toast.success("Budget spend reset");
          refreshSoon();
        } catch (error) {
          toast.error(error instanceof Error ? error.message : "Failed to reset budget");
        }
      })();
    });
  };

  const handleRemove = () => {
    if (budgetState.status !== "configured") {
      return;
    }
    if (!window.confirm("Remove budget from this key?")) {
      return;
    }

    startTransition(() => {
      void (async () => {
        try {
          await deleteAPIKeyBudget(apiKeyID);
          setBudgetState({ status: "unbudgeted" });
          toast.success("Budget removed");
          refreshSoon();
        } catch (error) {
          toast.error(error instanceof Error ? error.message : "Failed to remove budget");
        }
      })();
    });
  };

  const configuredBudget = budgetState.status === "configured" ? budgetState.budget : undefined;
  const controlsDisabled = isPending || budgetState.status !== "configured";

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Manage budget: {apiKeyName}</DialogTitle>
          <DialogDescription>
            Set, update, reset, or remove budget controls for this key.
          </DialogDescription>
        </DialogHeader>

        <div className="rounded-md border p-3">
          <APIKeyBudgetStatus state={budgetState} />
        </div>

        <APIKeyBudgetForm
          initialBudget={configuredBudget}
          disabled={isPending || budgetState.status === "unavailable"}
          onSubmit={handleSave}
        />

        <div className="flex gap-2">
          <Button
            type="button"
            variant="outline"
            disabled={controlsDisabled}
            onClick={handleReset}
          >
            Reset spend
          </Button>
          <Button
            type="button"
            variant="destructive"
            disabled={controlsDisabled}
            onClick={handleRemove}
          >
            Remove budget
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
