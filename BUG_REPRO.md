# Bug Reproduction

Mapping compatibility wraps `ErrConflict` with `%v`, so callers cannot classify a type conflict through `errors.Is`. Reproduce with the three targeted collection tests in `collection.json`.
