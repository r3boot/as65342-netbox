TARGET = as65342-netbox

BUILD_DIR = ./build
LIB_DIR = ./lib

RELEASE_NAME = $(TARGET)-$(shell uname -sm | tr A-Z a-z | sed -e 's, ,-,g')
RELEASE_DIR = ./$(RELEASE_NAME)

all: $(BUILD_DIR)/$(TARGET)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)

$(LIB_DIR)/netbox:
	mkdir -p $(LIB_DIR)/netbox
	swagger generate client --target=$(LIB_DIR)/netbox --spec=./swagger.json

fetch_dependencies:
	go get -v ./...

$(BUILD_DIR)/$(TARGET): $(BUILD_DIR) $(LIB_DIR)/netbox fetch_dependencies
	go build -v -o "${BUILD_DIR}/${TARGET}" "./cmd/${TARGET}/main.go"

release: $(RELEASE_DIR)
	strip -v $(BUILD_DIR)/$(TARGET)
	install -m 0755 $(BUILD_DIR)/$(TARGET) $(RELEASE_DIR)/$(TARGET)
	tar cvzf $(RELEASE_NAME).tar.gz $(RELEASE_DIR)

install:
	strip -v $(BUILD_DIR)/$(TARGET)
	install -o root -g root -m 0755 $(BUILD_DIR)/$(TARGET) \
		/usr/local/bin/$(TARGET)

clean:
	[[ -d "${BUILD_DIR}" ]] && rm -rf "${BUILD_DIR}" || true
	[[ -d "${RELEASE_DIR}" ]] && rm -rf "${RELEASE_DIR}" || true
