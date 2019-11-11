TARGET = as65342-netbox

BUILD_DIR = ./build
LIB_DIR = ./lib

all: $(BUILD_DIR)/$(TARGET)

$(LIB_DIR)/netbox:
	mkdir -p $(LIB_DIR)/netbox
	swagger generate client --target=$(LIB_DIR)/netbox --spec=./swagger.json

fetch_dependencies:
	go get -v ./...

$(BUILD_DIR)/$(TARGET): $(LIB_DIR)/netbox fetch_dependencies
	go build -v -o "${BUILD_DIR}/${TARGET}" "./cmd/${TARGET}/main.go"

install:
	strip -v $(BUILD_DIR)/$(TARGET)
	install -o root -g root -m 0755 $(BUILD_DIR)/$(TARGET) \
		/usr/local/bin/$(TARGET)

clean:
	[[ -d "${BUILD_DIR}" ]] && rm -rf "${BUILD_DIR}" || true
