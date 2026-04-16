# Velox

Fast photo pipeline.

CLI tool for moving media from Android to PC, then optionally, and recommended, to a VM, and finally into Immich.

## Highlights

- Android -> PC transfer via `adb` (recommended) or `mtp`
- Transfer modes: `partial`, `semi`, `full`
- SSH + SFTP transfer to a remote VM staging folder
- Remote `immich upload` execution over SSH
- Local temp-folder cleanup after successful end-to-end transfer
- `.env` loading support for VM and Immich credentials
- File suffix filtering: `jpg`, `jpeg`, `png`, `mp4`
- Benchmarking/profiling-friendly architecture with separable stages
- It is recommended to read the Notes at the end to maximize the speed of the pipeline

## Installation

### Build from source (Go)

Requires Go (see `go.mod`):

```bash
go build -o bin/velox .
./bin/velox
```

### Run directly

```bash
go run main.go
```

## Prerequisites

<details>
<summary>Click to expand</summary>

- Go: required to build/run from source
- Linux
- Android phone
- ADB: recommended for best performance
- SSH server on VM: required for `semi` and `full`
- Immich CLI on VM: required for `full`
- Immich server: reachable from the VM
- Optional `.env` file: recommended for credentials and instance settings

</details>

## Quick Start

### 1. Create a `.env` file in the project root

```env
VM_USER=your-vm-user
VM_IP=your-vm-ip
VM_AUTH=your-vm-password
IMMICH_INSTANCE_URL=http://your-vm-ip:your-vm-port
IMMICH_API_KEY=your_immich_api_key
```

### 2. Run the full pipeline with ADB (recommended mode for faster uploads)

```bash
go run main.go --transfer full --mode adb --suffix jpg --folder /var/tmp/velox-staging
```

## Usage

```bash
velox [--transfer full|semi|partial] [--mode adb|mtp] [--ip <vm-ip>] [--folder <vm-folder>] [--suffix <ext>]
```

### Examples

```bash
go run main.go --transfer partial --mode adb --suffix jpg
go run main.go --transfer semi --mode adb --ip 192.168.1.17 --suffix png
go run main.go --transfer full --mode adb --folder /var/tmp/velox-staging --suffix mp4
```

## Configuration

Velox currently reads runtime configuration from:

- `./.env`
- CLI flags

### Current flags

- `--transfer`
- `--mode`
- `--ip`
- `--folder`
- `--suffix`

## Data Storage

Velox currently stores or uses runtime data in:

| Data | Location |
|---|---|
| Optional env values | `./.env` |
| Local temp media staging | `${XDG_CACHE_HOME}/velox/staging` or `~/.cache/velox/staging` |
| Remote VM staging folder | `/var/tmp/velox-staging` |

## Notes

- `adb` is recommended over `mtp` for speed. You will need to enable developer mode on your phone. 
In my case I followed this simple guide [Guide](https://www.android.com/intl/en_uk/articles/enable-android-developer-settings/)
- Current bottlenecks after transfer are mostly VM-side.
- VM-side bottlenecks usually come from Immich background processing and storage or NFS-backed paths. It is 
recommended to try diferent concurrency set ups for the jobs inside the Immich app and see what works best for you.
I have found that usually around 5 for most concurrency related settings is a sweet spot.
- Direct writes into Immich-managed storage paths are not recommended.

## License

MIT License. See [`LICENSE`](LICENSE).
