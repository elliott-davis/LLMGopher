import { seed, type SeedSnapshot } from "../fixtures/seed";

let _store: SeedSnapshot = seed();

export function getStore(): SeedSnapshot {
  return _store;
}

// Restore to seed defaults. Called between test files.
export function reset(): void {
  _store = seed();
}
