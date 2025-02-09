APP_NAME = gsl
SRC = main.go
OUTPUT_DIR = build

PLATFORMS = \
    linux/amd64 \
    linux/arm64 \
    darwin/amd64 \
    darwin/arm64

BUILD_TARGETS = $(PLATFORMS:%=$(OUTPUT_DIR)/$(APP_NAME)-%)

all: $(BUILD_TARGETS)

$(OUTPUT_DIR)/$(APP_NAME)-%: $(SRC)
	@mkdir -p $(OUTPUT_DIR)
	@GOOS=$(word 1,$(subst /, ,$*)) GOARCH=$(word 2,$(subst /, ,$*)) go build -o $@ $(SRC) &
	@echo "Built: $@"

clean:
	@rm -rf $(OUTPUT_DIR)
	@echo "Cleaned build files."

.PHONY: all clean
