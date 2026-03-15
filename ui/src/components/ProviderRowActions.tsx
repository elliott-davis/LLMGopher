"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { toast } from "sonner";
import { Loader2, MoreHorizontal, Pencil, Trash2 } from "lucide-react";

import { deleteProvider, waitForProviderDeletionSync } from "@/lib/actions";
import { Provider } from "@/lib/types";
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
import EditProviderModal from "@/components/EditProviderModal";

type ProviderRowActionsProps = {
  provider: Provider;
};

export default function ProviderRowActions({ provider }: ProviderRowActionsProps) {
  const router = useRouter();
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [isDeleting, startDeleteTransition] = useTransition();
  const [isWaitingForSync, setIsWaitingForSync] = useState(false);

  const handleDeleteConfirm = () => {
    startDeleteTransition(() => {
      void (async () => {
        try {
          setIsWaitingForSync(true);
          const response = await deleteProvider(provider.id);
          if (!response.success) {
            toast.error("Deletion Failed", { description: response.error });
            return;
          }

          toast.message("Delete accepted. Waiting for cache sync...");
          const synced = await waitForProviderDeletionSync(provider.id, 30000, 1000);
          router.refresh();

          if (!synced) {
            toast.warning("Provider deleted, but cache sync is taking longer than expected.");
            return;
          }

          toast.success("Provider deleted");
          setDeleteOpen(false);
        } catch (error) {
          const message =
            error instanceof Error ? error.message : "Failed to delete provider";
          toast.error(message);
        } finally {
          setIsWaitingForSync(false);
        }
      })();
    });
  };

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger render={<Button variant="ghost" size="icon" />}>
          <MoreHorizontal className="size-4" aria-hidden />
          <span className="sr-only">Open actions</span>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem
            onClick={() => {
              setEditOpen(true);
            }}
          >
            <Pencil className="mr-2 size-4" />
            Edit
          </DropdownMenuItem>
          <DropdownMenuItem
            variant="destructive"
            onClick={() => {
              setDeleteOpen(true);
            }}
          >
            <Trash2 className="mr-2 size-4" />
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <EditProviderModal
        provider={provider}
        open={editOpen}
        onOpenChange={setEditOpen}
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
            <AlertDialogTitle>Delete provider</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this provider? This action cannot
              be undone.
            </AlertDialogDescription>
            {isWaitingForSync ? (
              <p className="mt-2 flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="size-4 animate-spin" aria-hidden />
                Deleting provider and waiting for cache sync...
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
              {isDeleting || isWaitingForSync ? "Deleting..." : "Delete"}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
