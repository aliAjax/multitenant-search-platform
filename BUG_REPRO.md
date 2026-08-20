# Bug Reproduction

RPC handlers discard the request context and snapshot calls reuse a stale context, so cancellation and deadlines cross request boundaries. Reproduce with the two API tests and the snapshot test in `collection.json`.
