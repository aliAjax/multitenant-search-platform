# Bug Reproduction

Batch WAL writes defer cleanup inside a loop and a finalizer can replace the operation error. Reproduce with the two targeted platform tests in `collection.json`.
