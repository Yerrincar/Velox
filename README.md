### Velox, fast photo pipeline.
A totally bespoke CLI tool for moving media from Android to PC, then optionally (and recommended) to a VM, and finally into Immich.

## Highlights
- Android -> PC transfer via `adb` (recommended) or `mtp`
- Transfer modes:
  - `partial`: mobile -> PC
  - `semi`: mobile -> PC -> VM
  - `full`: mobile -> PC -> VM -> Immich
- SSH + SFTP transfer to a remote VM staging folder
- Remote `immich upload` execution over SSH
- Local temp-folder cleanup after successful end-to-end transfer
- `.env` loading support for VM and Immich credentials
- File suffix filtering:
  - `jpg`
  - `jpeg`
  - `png`
  - `mp4`
- Benchmarking/profiling-friendly architecture with separable stages
## Installation
### Build from source (Go)
Requires Go (see `go.mod`):
```bash
go build -o bin/velox .
./bin/velox
Run directly
go run main.go
Prerequisites
<details>
<summary>Click to expand</summary>
- Go: required to build/run from source
- Linux 
- Android phone
- ADB: recommended for best performance. You 
- SSH server on VM: required for semi and full
- Immich CLI on VM: required for full
- Immich server: reachable from the VM
- Optional .env file: recommended for credentials and instance settings
</details>
Quick Start
1. Create a .env file in the project root:
VM_USER=your-vm-user
VM_IP=your-vm-ip
VM_AUTH=your-vm-password
IMMICH_INSTANCE_URL=http://your-vm-ip:your-vm-port
IMMICH_API_KEY=your_immich_api_key
2. Run the full pipeline with ADB:
go run main.go --transfer full --mode adb --suffix jpg --folder /var/tmp/velox-staging
Usage
velox [--transfer full|semi|partial] [--mode adb|mtp] [--ip <vm-ip>] [--folder <vm-folder>] [--suffix <ext>]
Examples:
go run main.go --transfer partial --mode adb --suffix jpg
go run main.go --transfer semi --mode adb --ip 192.168.1.17 --suffix png
go run main.go --transfer full --mode adb --folder /var/tmp/velox-staging --suffix mp4
Configuration
Velox currently reads runtime configuration from:
- ./.env
- CLI flags
Current flags include:
- --transfer
- --mode
- --ip
- --folder
- --suffix
Data Storage
Velox currently stores/uses local runtime data in:
Data
Optional env values
Local temp media staging
Remote VM staging folder
Notes
- adb is recommended over mtp for speed.
- Current bottlenecks after transfer are mostly VM-side:
  - Immich background processing
  - storage / NFS-backed paths
- Direct writes into Immich-managed storage paths are not recommended.
- Preferred flow is:
  - upload to VM staging folder
  - run immich upload from there
License
MIT License. See LICENSE (LICENSE).
