"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState, useTransition } from "react";
import { toast } from "sonner";

import { updateModel } from "@/lib/actions";
import { Model, Provider } from "@/lib/types";
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

type EditModelModalProps = {
  model: Model;
  providers: Provider[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export default function EditModelModal({
  model,
  providers,
  open,
  onOpenChange,
}: EditModelModalProps) {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [providerId, setProviderId] = useState(model.provider_id);
  const hasProviders = providers.length > 0;
  const submitDisabled = !hasProviders || isPending;

  const refreshModelsSoon = () => {
    router.refresh();
    window.setTimeout(() => router.refresh(), 1200);
    window.setTimeout(() => router.refresh(), 3000);
  };

  useEffect(() => {
    if (open) {
      setProviderId(model.provider_id);
    }
  }, [open, model.provider_id]);

  const handleSubmit = (formData: FormData) => {
    setErrorMessage(null);

    startTransition(() => {
      void (async () => {
        try {
          await updateModel(model.id, formData);
          toast.success("Model updated");
          onOpenChange(false);
          refreshModelsSoon();
        } catch (error) {
          const message =
            error instanceof Error ? error.message : "Failed to update model";
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
          <DialogTitle>Edit Model</DialogTitle>
          <DialogDescription>
            Update the model alias, name, provider, or context window.
          </DialogDescription>
        </DialogHeader>

        <form action={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-model-alias">Alias</Label>
            <Input
              id="edit-model-alias"
              name="alias"
              required
              placeholder="gpt-4o"
              defaultValue={model.alias}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-model-name">Name</Label>
            <Input
              id="edit-model-name"
              name="name"
              required
              placeholder="gpt-4o-2024-11-20"
              defaultValue={model.name}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-provider-select">Provider</Label>
            <input type="hidden" name="provider_id" value={providerId} />
            <Select
              disabled={!hasProviders}
              value={providerId}
              onValueChange={(v) => setProviderId(v ?? "")}
            >
              <SelectTrigger id="edit-provider-select" className="w-full">
                <SelectValue
                  placeholder={
                    hasProviders
                      ? "Select a provider"
                      : "No providers found - create one first"
                  }
                />
              </SelectTrigger>
              <SelectContent>
                {hasProviders ? (
                  providers.map((provider) => (
                    <SelectItem key={provider.id} value={provider.id}>
                      {provider.name}
                    </SelectItem>
                  ))
                ) : (
                  <SelectItem value="__no_provider__" disabled>
                    No providers found - create one first
                  </SelectItem>
                )}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-context-window">Context Window</Label>
            <Input
              id="edit-context-window"
              name="context_window"
              type="number"
              min={1}
              defaultValue={model.context_window}
              required
            />
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={submitDisabled}>
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
