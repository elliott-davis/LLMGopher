"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { toast } from "sonner";
import { Loader2, MoreHorizontal, Pencil, Trash2 } from "lucide-react";

import { deleteModel, waitForModelDeletionSync } from "@/lib/actions";
import { Model, Provider } from "@/lib/types";
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
import EditModelModal from "@/components/EditModelModal";

type ModelRowActionsProps = {
  model: Model;
  providers: Provider[];
};

export default function ModelRowActions({ model, providers }: ModelRowActionsProps) {
  const router = useRouter();
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [isWaitingForSync, setIsWaitingForSync] = useState(false);
  const [isDeleting, startDeleteTransition] = useTransition();

  const refreshModelsSoon = () => {
    router.refresh();
    window.setTimeout(() => router.refresh(), 1200);
    window.setTimeout(() => router.refresh(), 3000);
  };

  const handleDeleteConfirm = () => {
    startDeleteTransition(() => {
      void (async () => {
        try {
          setIsWaitingForSync(true);
          await deleteModel(model.id);
          toast.message("Delete accepted. Waiting for cache sync...");

          const synced = await waitForModelDeletionSync(model.id, 30000, 1000);
          refreshModelsSoon();

          if (synced) {
            toast.success("Model deleted");
            setDeleteOpen(false);
            return;
          }

          toast.warning(
            "Delete is taking longer than expected to sync. The page will keep refreshing."
          );
          setDeleteOpen(false);
        } catch (error) {
          const message =
            error instanceof Error ? error.message : "Failed to delete model";
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

      <EditModelModal
        model={model}
        providers={providers}
        open={editOpen}
        onOpenChange={setEditOpen}
      />

      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete model</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this model? This action cannot be
              undone and traffic will fail once the cache syncs.
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
              {isDeleting || isWaitingForSync ? "Deleting..." : "Delete"}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
