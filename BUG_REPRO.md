# Bug Reproduction

Disabling the analyzer stores a typed-nil implementation in an interface and the registry calls it, producing a nil-pointer panic. Reproduce with the two targeted analysis tests in `collection.json`.
