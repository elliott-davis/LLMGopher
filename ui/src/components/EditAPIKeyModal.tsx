"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { toast } from "sonner";

import { updateAPIKey } from "@/lib/actions";
import { APIKey, Model } from "@/lib/types";
import APIKeyLifecycleForm from "@/components/APIKeyLifecycleForm";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

type EditAPIKeyModalProps = {
  apiKey: APIKey;
  models: Model[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export default function EditAPIKeyModal({
  apiKey,
  models,
  open,
  onOpenChange,
}: EditAPIKeyModalProps) {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const refreshKeysSoon = () => {
    router.refresh();
    window.setTimeout(() => router.refresh(), 1200);
    window.setTimeout(() => router.refresh(), 3000);
  };

  const handleSubmit = (formData: FormData) => {
    setErrorMessage(null);

    startTransition(() => {
      void (async () => {
        try {
          await updateAPIKey(apiKey.id, formData);
          toast.success("API key updated", {
            description: "The gateway cache may take a few seconds to reflect changes.",
          });
          onOpenChange(false);
          refreshKeysSoon();
        } catch (error) {
          const message =
            error instanceof Error ? error.message : "Failed to update API key";
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
          <DialogTitle>Edit API Key</DialogTitle>
          <DialogDescription>
            Update lifecycle fields without exposing raw key material.
          </DialogDescription>
        </DialogHeader>

        <form action={handleSubmit} className="space-y-4">
          <APIKeyLifecycleForm
            idPrefix={`edit-api-key-${apiKey.id}`}
            apiKey={apiKey}
            models={models}
            disabled={isPending}
          />

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? "Saving..." : "Save changes"}
            </Button>
          </DialogFooter>
          {errorMessage ? (
            <p className="text-sm text-destructive">{errorMessage}</p>
          ) : null}
        </form>
      </DialogContent>
    </Dialog>
  );
}
