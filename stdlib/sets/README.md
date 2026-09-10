# sets

`sets` is a v0 runtime-backed unique string collection.

## Implemented API

- `sets.new(): set`
- `sets.add(set: set, value: string): void`
- `sets.contains(set: set, value: string): bool`
- `sets.length(set: set): int`

## Current Limits

- Sets store strings only.
- Hash-table storage is deferred; v0 uses a small linear collection.
- Memory lifetime is runtime-owned in v0.
