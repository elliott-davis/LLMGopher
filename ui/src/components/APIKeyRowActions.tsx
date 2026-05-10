"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { toast } from "sonner";
import {
  Loader2,
  MoreHorizontal,
  Pencil,
  Power,
  PowerOff,
  Trash2,
} from "lucide-react";

import {
  deleteAPIKey,
  setAPIKeyActiveState,
  waitForAPIKeyDeletionSync,
} from "@/lib/actions";
import { APIKey, APIKeyBudgetState, Model } from "@/lib/types";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import EditAPIKeyModal from "@/components/EditAPIKeyModal";
import APIKeyBudgetModal from "@/components/APIKeyBudgetModal";

type APIKeyRowActionsProps = {
  apiKey: APIKey;
  models: Model[];
  budgetState: APIKeyBudgetState;
};

export default function APIKeyRowActions({ apiKey, models, budgetState }: APIKeyRowActionsProps) {
  const router = useRouter();
  const [editOpen, setEditOpen] = useState(false);
  const [budgetOpen, setBudgetOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [isWaitingForSync, setIsWaitingForSync] = useState(false);
  const [isToggling, startToggleTransition] = useTransition();
  const [isDeleting, startDeleteTransition] = useTransition();

  const refreshKeysSoon = () => {
    router.refresh();
    window.setTimeout(() => router.refresh(), 1200);
    window.setTimeout(() => router.refresh(), 3000);
  };

  const handleActiveToggle = () => {
    const nextState = !apiKey.is_active;
    const actionLabel = nextState ? "reactivated" : "deactivated";

    startToggleTransition(() => {
      void (async () => {
        try {
          await setAPIKeyActiveState(apiKey.id, nextState);
          toast.success(`API key ${actionLabel}`, {
            description: "The gateway cache may take a few seconds to sync.",
          });
          refreshKeysSoon();
        } catch (error) {
          const message =
            error instanceof Error ? error.message : `Failed to ${actionLabel} API key`;
          toast.error(message);
        }
      })();
    });
  };

  const handleDeleteConfirm = () => {
    startDeleteTransition(() => {
      void (async () => {
        try {
          setIsWaitingForSync(true);
          await deleteAPIKey(apiKey.id);
          toast.message("Delete accepted. Waiting for cache sync...");

          const synced = await waitForAPIKeyDeletionSync(apiKey.id, 30000, 1000);
          refreshKeysSoon();

          if (synced) {
            toast.success("API key deleted");
            setDeleteOpen(false);
            return;
          }

          toast.warning(
            "Delete is taking longer than expected to sync. The page will keep refreshing."
          );
          setDeleteOpen(false);
        } catch (error) {
          const message =
            error instanceof Error ? error.message : "Failed to delete API key";
          toast.error(message);
        } finally {
          setIsWaitingForSync(false);
        }
      })();
    });
  };

  return (
    <>
      <div className="flex items-center gap-1">
        <Button size="sm" variant="outline" onClick={() => setBudgetOpen(true)}>
          Manage budget
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={<Button variant="ghost" size="icon" disabled={isToggling} data-testid="key-actions-menu" />}
          >
            <MoreHorizontal className="size-4" aria-hidden />
            <span className="sr-only">Open actions</span>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => setEditOpen(true)}>
              <Pencil className="mr-2 size-4" />
              Edit
            </DropdownMenuItem>
            <DropdownMenuItem
              data-testid={apiKey.is_active ? "key-deactivate" : "key-reactivate"}
              onClick={handleActiveToggle}
              disabled={isToggling}
            >
              {apiKey.is_active ? (
                <PowerOff className="mr-2 size-4" />
              ) : (
                <Power className="mr-2 size-4" />
              )}
              {apiKey.is_active ? "Deactivate" : "Reactivate"}
            </DropdownMenuItem>
            <DropdownMenuItem
              data-testid="key-delete"
              variant="destructive"
              onClick={() => setDeleteOpen(true)}
            >
              <Trash2 className="mr-2 size-4" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <EditAPIKeyModal
        apiKey={apiKey}
        models={models}
        open={editOpen}
        onOpenChange={setEditOpen}
      />
      <APIKeyBudgetModal
        apiKeyID={apiKey.id}
        apiKeyName={apiKey.name}
        initialState={budgetState}
        open={budgetOpen}
        onOpenChange={setBudgetOpen}
      />

      <AlertDialog
        open={deleteOpen}
        onOpenChange={(nextOpen) => {
          if (isDeleting || isWaitingForSync) {
            return;
          }
          setDeleteOpen(nextOpen);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete API key</AlertDialogTitle>
            <AlertDialogDescription>
              Permanently delete <span className="font-medium">{apiKey.name}</span>? Clients using this key will
              stop authenticating after gateway synchronization. This action
              cannot be undone.
            </AlertDialogDescription>
            {isWaitingForSync ? (
              <p className="mt-2 flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="size-4 animate-spin" aria-hidden />
                Deletion requested. Waiting for cache sync...
              </p>
            ) : null}
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting || isWaitingForSync}>
              Cancel
            </AlertDialogCancel>
            <Button
              variant="destructive"
              onClick={handleDeleteConfirm}
              disabled={isDeleting || isWaitingForSync}
            >
              {isDeleting || isWaitingForSync ? "Deleting..." : "Delete API key"}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
