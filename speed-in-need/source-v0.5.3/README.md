# SPEED IN NEED v0.5.3 Go source snapshot

This branch adds a selected Go source snapshot from the locally retained editable tree. The source folder keeps a historical `v0.5.2` label, while its project README identifies the code as updated to v0.5.3.

Only Go source, Go tests, `go.mod`, and this note are included. Assets and car/profile models, world data, native/runtime binaries, generated files, local logs/reports, build/asset-generation scripts, and source files containing machine-specific absolute paths are omitted. This avoids publishing unverified assets or local machine details.

This is not a complete playable release and cannot be built on its own because required assets and runtimes are intentionally absent. The original source tree passed its Go test suites locally; this filtered snapshot has not been built independently.

No project license file was present in the source tree. No open-source license is granted by this snapshot.
