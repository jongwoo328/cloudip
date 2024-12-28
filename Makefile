APP_NAME := cloudip
PLATFORMS := linux/386 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
BUILD_DIR := build
VERSION := 0.4.0

build:
	@for platform in $(PLATFORMS); do \
		platform_split=($${platform//\// }); \
		GOOS=$${platform_split[0]} GOARCH=$${platform_split[1]} \
		go build \
			-ldflags "-X 'cloudip/cmd.Version=$(VERSION)'" \
			-o $(BUILD_DIR)/$(APP_NAME)-$${platform_split[0]}-$${platform_split[1]}-$(VERSION) .; \
	done

# Clean build directory
clean:
	rm -rf $(BUILD_DIR)

.PHONY: build clean
