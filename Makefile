.PHONY: make-rootfs
make-rootfs:
	bash scripts/rootfs.sh

.PHONY: test-sandbox
test-sandbox: make-rootfs
	go test github.com/joshjms/castletown/sandbox -v
