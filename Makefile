BUILD_IMAGE = action-update-go
CURRENT_DIR = $(shell pwd)

dist/update-go:
	docker build -t "$(BUILD_IMAGE)" .
	docker run --rm -v "$(CURRENT_DIR)/dist:/out" --entrypoint cp "$(BUILD_IMAGE)" /update-go /out/update-go

.PHONY: clean
clean:
	rm dist/update-go
