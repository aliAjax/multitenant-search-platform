# Bug Reproduction

Concurrent `EnsureIndex` calls race on the document service index map, while replacing a document leaves old postings searchable. Reproduce with:

`go test -race ./internal/document -run '^TestConcurrentEnsureIndex$' -count=1`

`go test -race ./internal/document -run '^TestIndexReplacesTerms$' -count=1`
