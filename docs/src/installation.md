# Install Mergeway

Pick the method that fits your setup. Each installs a single binary named `mw`.

## Option 1 – Download a Release (macOS, Linux)

```bash
curl -L https://github.com/mergewayhq/mergeway-cli/releases/download/v0.11.0/mw-
$(uname | tr '[:upper:]' '[:lower:]')-amd64.tar.gz \
  | tar -xz
sudo mv mw /usr/local/bin/
```

Check the published SHA-256 checksum before moving the binary if you operate in a locked-down environment.

## Option 2 – Go Install (for contributors)

```bash
go install github.com/mergewayhq/mergeway-cli@latest
```

This drops the binary in `$GOPATH/bin`. Prefer tagged versions in production.

## Verify

```bash
mw --version
```

You should see something similar to:

```
Mergeway CLI v0.11.0 (commit abc1234)
```

If the command is missing, confirm that the installation path is on your `PATH`.

Move on to the [Getting Started](getting-started.md) guide once the binary is available.
