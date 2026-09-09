APP_NAME = desktop-tasks
INSTALL_DIR = $(HOME)/.local/bin

all: build

build:
	go build -ldflags="-s -w" -o $(APP_NAME) .

install: build
	install -d $(INSTALL_DIR)
	install -m 755 $(APP_NAME) $(INSTALL_DIR)/$(APP_NAME)

clean:
	rm -f $(APP_NAME)

.PHONY: all build install clean
