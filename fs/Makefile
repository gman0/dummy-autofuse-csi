BUILD_DIR ?= build
CFLAGS := $(shell pkg-config fuse3 --cflags) -Wall -s -O2
LIBS := $(shell pkg-config fuse3 --libs)

VERSION ?= $(shell git describe --long --tags --dirty --always)

$(shell mkdir -p $(BUILD_DIR))

all: dummy-fuse

dummy-fuse: dummy-fuse.c $(BUILD_DIR)/version.o
	$(CC) $(CFLAGS) $(LIBS) $^ -o $(BUILD_DIR)/$@

$(BUILD_DIR)/version.c:
	@printf "%s" 'const char *dummy_version = "$(VERSION)";' > $(BUILD_DIR)/version.c

$(BUILD_DIR)/version.o: $(BUILD_DIR)/version.c
	$(CC) -c -o $(@:.c=.o) $<

clean:
	rm -rf $(BUILD_DIR)

.PHONY: clean
