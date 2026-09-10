# dictionaries

`dictionaries` is a v0 runtime-backed string key/value map.

## Implemented API

- `dictionaries.new(): dictionary`
- `dictionaries.set(dictionary: dictionary, key: string, value: string): void`
- `dictionaries.get(dictionary: dictionary, key: string): string`
- `dictionaries.contains(dictionary: dictionary, key: string): bool`
- `dictionaries.length(dictionary: dictionary): int`

## Current Limits

- Keys and values are strings only.
- Missing keys return an empty string until Qwic has a standard result/error
  model.
- Hash-table storage is deferred; v0 uses a small linear map.
