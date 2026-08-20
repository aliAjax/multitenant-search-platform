# Bug Reproduction

The zero-value quota manager writes to a nil map and a disabled limiter is represented by a nil pointer that is called. Reproduce with the three targeted quota tests in `collection.json`.
