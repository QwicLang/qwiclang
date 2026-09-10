# lists

`lists` is a v0 runtime-backed sequence package.

```qwic
import lists

public func main() {
    const names: list = lists.new()
    lists.push(names, "Qwic")
    print(lists.get(names, 0))
}
```

## Implemented API

- `lists.new(): list`
- `lists.push(list: list, value: string): void`
- `lists.get(list: list, index: int): string`
- `lists.length(list: list): int`
- `lists.contains(list: list, value: string): bool`

## Current Limits

- Lists store strings only in this bootstrap implementation.
- Out-of-range reads return an empty string until Qwic has a standard
  result/error model.
- Memory lifetime is runtime-owned in v0.
