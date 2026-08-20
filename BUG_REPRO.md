# Bug Reproduction

Missing tenant and collection errors lose their sentinel chain, and retry loops ignore cancellation. Reproduce with the four targeted tenant tests in `collection.json`.
