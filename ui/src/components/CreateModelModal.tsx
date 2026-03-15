"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { toast } from "sonner";

import { createModel } from "@/lib/actions";
import { Provider } from "@/lib/types";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

type CreateModelModalProps = {
  providers: Provider[];
};

export default function CreateModelModal({ providers }: CreateModelModalProps) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [isPending, startTransition] = useTransition();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const hasProviders = providers.length > 0;
  const submitDisabled = !hasProviders || isPending;

  const refreshModelsSoon = () => {
    router.refresh();
    window.setTimeout(() => router.refresh(), 1200);
    window.setTimeout(() => router.refresh(), 3000);
  };

  const handleSubmit = (formData: FormData) => {
    setErrorMessage(null);

    startTransition(() => {
      void (async () => {
        try {
          await createModel(formData);
          toast.success("Model created");
          setOpen(false);
          refreshModelsSoon();
        } catch (error) {
          const message = error instanceof Error ? error.message : "Failed to create model";
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
        }
      }}
    >
      <DialogTrigger render={<Button />}>Add Model</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Model</DialogTitle>
          <DialogDescription>
            Add a model alias and attach it to a configured provider instance.
          </DialogDescription>
        </DialogHeader>

        <form action={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="model-alias">Alias</Label>
            <Input id="model-alias" name="alias" required placeholder="gpt-4o" />
          </div>

          <div className="space-y-2">
            <Label htmlFor="model-name">Name</Label>
            <Input id="model-name" name="name" required placeholder="gpt-4o-2024-11-20" />
          </div>

          <div className="space-y-2">
            <Label htmlFor="provider-select">Provider</Label>
            <Select name="provider_id" disabled={!hasProviders} required>
              <SelectTrigger id="provider-select" className="w-full">
                <SelectValue
                  placeholder={
                    hasProviders ? "Select a provider" : "No providers found - create one first"
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
            <Label htmlFor="context-window">Context Window</Label>
            <Input
              id="context-window"
              name="context_window"
              type="number"
              min={1}
              defaultValue={128000}
              required
            />
          </div>

          <DialogFooter>
            <Button type="submit" disabled={submitDisabled}>
              {isPending ? "Saving..." : !hasProviders ? "Create a provider first" : "Create Model"}
            </Button>
          </DialogFooter>
          {errorMessage ? <p className="text-sm text-destructive">{errorMessage}</p> : null}
        </form>
      </DialogContent>
    </Dialog>
  );
}
