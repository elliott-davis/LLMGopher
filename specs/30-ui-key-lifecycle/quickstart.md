# Quickstart: UI API Key Lifecycle Controls

## Prerequisites

- Docker Compose development stack is available.
- Local dev admin/client key is `sk-test-key-1:key-001`.
- Gateway and UI are built from the current branch.

## Start the Stack

```bash
make dev
```

If the stack is already running after code changes:

```bash
make dev-restart
```

Open the admin UI and navigate to `/keys`.

## Smoke Checks

### 1. Create a Key With Lifecycle Fields

1. Open the create key dialog.
2. Enter a name, non-negative rate limit, optional expiration, metadata JSON, and one or more allowed models.
3. Save the key.
4. Verify the generated raw key is displayed once.
5. Close and reopen the key inventory.
6. Verify the raw key is no longer displayed and lifecycle fields are visible in the inventory.

Expected result: the key appears with active status, rate limit, expiration state, metadata summary, and model allowlist or unrestricted state.

### 2. Update an Existing Key

1. Choose an existing key and open edit.
2. Change the name, rate limit, expiration, metadata, and model allowlist.
3. Save.
4. Refresh the inventory if needed.

Expected result: the updated values appear after the gateway synchronization window.

### 3. Deactivate and Reactivate

1. Deactivate an active test key.
2. Verify the inventory shows inactive state and a reactivation option.
3. Reactivate the same key.
4. Verify the inventory shows active state.

Expected result: no new key is created and the same key ID remains visible.

### 4. Delete a Test Key

1. Choose a non-production test key.
2. Click delete.
3. Confirm the destructive action.
4. Wait for the inventory refresh.

Expected result: the key is removed from the inventory and the UI explains any synchronization delay.

### 5. Validate Error Handling

1. Attempt to submit invalid metadata.
2. Attempt to submit a negative rate limit.
3. Stop the gateway or point the UI at an unavailable gateway and attempt a save.

Expected result: the UI preserves form state, explains the failed action, and does not expose any raw key material.

## Verification Commands

```bash
cd ui
npm run lint
npm run build
```

Run backend API key tests only if backend contract changes are made:

```bash
go test ./internal/api/... -run 'Test.*APIKey' -v
```

## Rollback Notes

This feature is UI-only when implemented as planned. Reverting the UI changes removes lifecycle controls from the admin UI but does not change gateway key enforcement or existing admin API behavior.
