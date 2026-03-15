"use client";

import { useState, useTransition } from "react";
import { toast } from "sonner";

import { createAPIKey } from "@/lib/actions";
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

export default function CreateAPIKeyModal() {
  const [open, setOpen] = useState(false);
  const [isPending, startTransition] = useTransition();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [createdKey, setCreatedKey] = useState<string | null>(null);

  const handleSubmit = (formData: FormData) => {
    setErrorMessage(null);
    setCreatedKey(null);

    startTransition(() => {
      void (async () => {
        try {
          const key = await createAPIKey(formData);
          setCreatedKey(key);
          toast.success("API key created");
        } catch (error) {
          const message = error instanceof Error ? error.message : "Failed to create API key";
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
        setOpen(nextOpen);
        if (!nextOpen) {
          setErrorMessage(null);
          setCreatedKey(null);
        }
      }}
    >
      <DialogTrigger render={<Button />}>Create Key</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create API Key</DialogTitle>
          <DialogDescription>
            Generated secrets are shown once. Copy and store the key before closing this dialog.
          </DialogDescription>
        </DialogHeader>

        <form action={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="api-key-name">Name</Label>
            <Input id="api-key-name" name="name" required placeholder="Production Service Key" />
          </div>

          <div className="space-y-2">
            <Label htmlFor="api-key-rate-limit">Rate Limit (RPS)</Label>
            <Input
              id="api-key-rate-limit"
              name="rate_limit_rps"
              type="number"
              min={0}
              defaultValue={100}
              required
            />
          </div>

          {createdKey ? (
            <div className="space-y-2 rounded-md border bg-muted/30 p-3">
              <Label htmlFor="generated-api-key">Generated Key</Label>
              <Input id="generated-api-key" readOnly value={createdKey} className="font-mono text-xs" />
            </div>
          ) : null}

          <DialogFooter>
            <Button type="submit" disabled={isPending}>
              {isPending ? "Saving..." : "Create API Key"}
            </Button>
          </DialogFooter>
          {errorMessage ? <p className="text-sm text-destructive">{errorMessage}</p> : null}
        </form>
      </DialogContent>
    </Dialog>
  );
}
