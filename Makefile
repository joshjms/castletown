all:
	echo "what"

.PHONY: prepare-dirs
prepare-dirs:
	@sudo mkdir -p /tmp/castletown/storage
	@sudo mkdir -p /tmp/castletown/images
	@sudo mkdir -p /tmp/castletown/libcontainer
	@sudo mkdir -p /tmp/castletown/overlayfs
	@sudo mkdir -p /tmp/castletown/rootfs

.PHONY: make-rootfs
make-rootfs: prepare-dirs
	sudo bash scripts/rootfs.sh

.PHONY: test-sandbox
test-sandbox: make-rootfs
	@echo "Running sandbox tests..."
	sudo env "PATH=$$PATH:/usr/local/go/bin" go test github.com/joshjms/castletown/sandbox -v

.PHONY: test-job
test-job: make-rootfs
	@echo "Running job tests..."
	sudo env "PATH=$$PATH:/usr/local/go/bin" go test github.com/joshjms/castletown/job -v


.PHONY: build
build:
	bash scripts/build.sh

.PHONY: dev
dev:
	sudo env "PATH=$$PATH:/usr/local/go/bin" go run main.go server

.PHONY: e2e
e2e: prepare-dirs make-rootfs
	@echo "Running end-to-end tests..."
	@echo "Building castletown..."
	@sudo env "PATH=$$PATH:/usr/local/go/bin" go build -o tmp/castletown main.go
	@echo "Starting castletown server..."
	@sudo tmp/castletown server &
	@sleep 2
	sudo env "PATH=$$PATH:/usr/local/go/bin" go test -v ./tests/e2e -timeout 2m
	@sudo pkill castletown
	@sudo rm -rf tmp

