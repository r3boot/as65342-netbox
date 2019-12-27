TARGET = ansible-generator

BUILD_DIR = ./build
LIB_DIR = ./lib
CMD_DIR = ./cmd

RELEASE_NAME = $(TARGET)-$(shell uname -sm | tr A-Z a-z | sed -e 's, ,-,g')
RELEASE_DIR = ./$(RELEASE_NAME)

PREFIX = /usr/local

TARGETS = ansible-generator \
		  backup-generator \
		  dns-generator \
		  icinga2-generator \
		  mailman-generator

all: $(TARGETS)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)

$(LIB_DIR)/netbox:
	mkdir -p $(LIB_DIR)/netbox
	swagger generate client --target=$(LIB_DIR)/netbox --spec=./swagger.json

fetch_dependencies:
	go get -v ./...

$(TARGETS):
	go build -v -o $(BUILD_DIR)/$@ $(CMD_DIR)/$@/main.go

release: $(RELEASE_DIR)
	strip -v $(BUILD_DIR)/{ansible,backup,dns,icinga2}-generator
	install -m 0755 $(BUILD_DIR)/ansible-generator \
		$(RELEASE_DIR)/ansible-generator
	install -m 0755 $(BUILD_DIR)/backup-generator \
		$(RELEASE_DIR)/backup-generator
	install -m 0755 $(BUILD_DIR)/dns-generator \
		$(RELEASE_DIR)/dns-generator
	install -m 0755 $(BUILD_DIR)/icinga2-generator \
		$(RELEASE_DIR)/icinga2-generator
	tar cvzf $(RELEASE_NAME).tar.gz $(RELEASE_DIR)

install:
	strip -v $(BUILD_DIR)/{ansible,backup,dns,icinga2}-generator
	install -m 0755 $(BUILD_DIR)/ansible-generator \
		$(PREFIX)/bin/ansible-generator
	install -m 0755 $(BUILD_DIR)/backup-generator \
		$(PREFIX)/bin/backup-generator
	install -m 0755 $(BUILD_DIR)/dns-generator \
		$(PREFIX)/bin/dns-generator
	install -m 0755 $(BUILD_DIR)/icinga2-generator \
		$(PREFIX)/bin/icinga2-generator

clean:
	[[ -d "${BUILD_DIR}" ]] && rm -rf "${BUILD_DIR}" || true
	[[ -d "${RELEASE_DIR}" ]] && rm -rf "${RELEASE_DIR}" || true
