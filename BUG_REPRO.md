# Bug Reproduction

Segment filtering, sorting, and metadata copying reuse caller-backed slices, so later operations mutate input ordering or contents. Reproduce with the three targeted index tests in `collection.json`.
