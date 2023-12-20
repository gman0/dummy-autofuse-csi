GOLDFLAGS := "-w -s"
BUILD_DIR ?= build

$(shell mkdir -p $(BUILD_DIR))

dummy-fuse-workload:
	CGO_ENABLED=0 go build -ldflags $(GOLDFLAGS) -o $(BUILD_DIR)/$@ cmd/main.go

clean:
	rm -rf $(BUILD_DIR)

.PHONY: clean
