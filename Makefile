.PHONY: build install clean

BINARY_NAME=gobuild
INSTALL_DIR=/usr/local/bin

build:
	go build -o $(BINARY_NAME) ./cmd/gobuild

install: build
	sudo mv $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)"

clean:
	rm -f $(BINARY_NAME)
	@echo "Cleaned build artifacts"
