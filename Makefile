
.PHONY: test
test:
	go test -v -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt ./attacker/... ./gui/...

.PHONY: mockgen
mockgen:
	mockgen -source=./gui/widgets.go -destination=./gui/widgets.mock.go -package=gui
