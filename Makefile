BINARY_NAME=warpp
INSTALL_PATH=/usr/local/bin

.PHONY: build install uninstall clean

build:
	go build -o $(BINARY_NAME) .

install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@# Create wp symlink only if wp doesn't exist
	@if ! command -v wp &> /dev/null; then \
		echo "Creating wp symlink..."; \
		ln -sf $(INSTALL_PATH)/$(BINARY_NAME) $(INSTALL_PATH)/wp; \
	else \
		echo "wp command already exists, skipping symlink"; \
	fi
	@echo "Done! Run 'warpp' to start"

uninstall:
	@echo "Uninstalling..."
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@# Only remove wp if it's our symlink
	@if [ -L $(INSTALL_PATH)/wp ] && [ "$$(readlink $(INSTALL_PATH)/wp)" = "$(INSTALL_PATH)/$(BINARY_NAME)" ]; then \
		rm -f $(INSTALL_PATH)/wp; \
	fi
	@echo "Done!"

clean:
	rm -f $(BINARY_NAME)
