
.PHONY: test
test:
	go test -race -v -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt ./...

.PHONY: mockgen
mockgen:
	mockgen -source=./gui/widgets.go -destination=./gui/widgets.mock.go -package=gui

.PHONY: release-dry-run
release-dry-run:
	goreleaser release --snapshot --clean
