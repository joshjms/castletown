# castletown

## Docker quickstart

1. Install Docker (24.x or newer) with Compose v2 and ensure you can run privileged containers.
2. From the repository root run:
   ```bash
   docker compose up --build
   ```
   The first boot downloads the `gcc:15-bookworm` root filesystem via `skopeo`/`umoci`, so expect the initial `castletown` container start to take a few minutes.
3. Watch the worker logs:
   ```bash
   docker compose logs -f castletown
   ```
4. When you are done:
   ```bash
   docker compose down
   ```

The Compose stack launches:

- `castletown`: the sandbox worker. Runs privileged so it can create nested containers, exposes metrics on `:9090`, and consumes submissions from RabbitMQ.
- `postgres`: stores problems, submissions, and metadata.
- `rabbitmq`: queue that feeds submissions to the worker (management UI on <http://localhost:15672>).
- `minio`: placeholder object storage for large blobs and artifacts (console on <http://localhost:9001>).

Named Docker volumes keep the worker stateful directories (`/tmp/castletown/*`, `/var/castletown/*`) so that cached images, overlays, and problem artifacts survive container restarts.

### Useful commands

- Rebuild just the worker image: `docker compose build castletown`
- Tail only dependency logs: `docker compose logs -f postgres rabbitmq minio`
- Open a shell inside the worker: `docker compose exec castletown bash`
- Skip the automatic `gcc-15-bookworm` bootstrap if you already populated `castletown-images`: `CASTLETOWN_SKIP_ROOTFS=1 docker compose up`

## Manual setup

If you need to run Castletown directly on a host (without Docker), follow the more detailed [Getting Started guide](docs/getting-started.md) to prepare cgroup delegation, rootfs images, and prerequisites.

## Contributing

Contributions are welcome! Please open an issue or pull request with improvements. When changing the worker runtime, make sure the Docker image stays reproducible and update this README accordingly.
