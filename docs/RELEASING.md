# Releasing and Publishing credentialctl

This project uses [GoReleaser](https://goreleaser.com/) via GitHub Actions to automate building, dynamic library bundling, packaging, and publishing across distribution channels.

## What Happens on Release

When you push a Git tag starting with `v` (such as `v0.1.5`), the release workflow (`.github/workflows/release.yaml`) executes GoReleaser to perform these steps:

1. Runs `scripts/fetch-credentio-lib.sh` to stage prebuilt `libcredentio_c` shared libraries and C-ABI headers.
2. Compiles CGO binaries with dynamic RPATH embedding (`@loader_path`, `@executable_path`, `@executable_path/../lib`, `@executable_path/lib`, `$ORIGIN`, `$ORIGIN/../lib`, `$ORIGIN/lib`).
3. Injects the Git tag, commit SHA, and build timestamp into `credentialctl version` through compiler linker flags.
4. Packages binaries into `.tar.gz` archives alongside native dynamic libraries (`libcredentio_c.dylib` or `libcredentio_c.so`), `README.md`, and `LICENSE`.
5. Generates SHA-256 cryptographic signatures in `checksums.txt`.
6. Publishes artifacts and generated release notes to a new GitHub Release.
7. Pushes an updated Homebrew formula to [`ghchinoy/homebrew-tap`](https://github.com/ghchinoy/homebrew-tap).

## Required Secrets and Tokens

The release workflow requires the following GitHub repository secrets:

| Secret | Purpose | Permissions |
|---|---|---|
| `GITHUB_TOKEN` | Automatic GitHub Actions token used to publish releases and upload release assets | `contents: write` |
| `HOMEBREW_TAP_GITHUB_TOKEN` | Personal Access Token (PAT) used to push updated formulas to `ghchinoy/homebrew-tap` | `repo` scope |

## Upstream Dependency and C-ABI Alignment

`credentialctl` depends on the Google Credentio C-ABI implementation distributed via `github.com/ghchinoy/credentio-contributions`. 

Maintain strict version synchronization across the following files:
- `go.mod`: `github.com/ghchinoy/credentio-contributions/go`
- `scripts/fetch-credentio-lib.sh`: Default `VERSION` parameter
- `third_party/credentio/include/credentio_c.h`: C-ABI function declarations

Version skew between the Go wrapper and native dynamic libraries can trigger symbol resolution errors or ABI mismatches at runtime.

Before releasing, verify dependency versions:

````bash
# Check Go module dependency
grep credentio-contributions go.mod

# Check fetch script target version
grep 'VERSION=' scripts/fetch-credentio-lib.sh
````

## Dynamic Linkage and RPATH Architecture

`credentialctl` uses dynamic linking for high-performance C2PA validation. 

### macOS SIP and Runtime Search Paths
macOS System Integrity Protection (SIP) blocks runtime inheritance of `DYLD_LIBRARY_PATH` in subshells. To guarantee standalone execution without environment variable overrides, `credentialctl` embeds explicit Runpath Search Paths (RPATHs):

- `@loader_path`: Searches the folder containing the binary (used in extracted standalone tarballs).
- `@executable_path`: Searches the folder containing the running executable.
- `@executable_path/../lib`: Searches the adjacent `lib/` directory (used in Homebrew installations where binaries reside in `bin/` and shared libraries reside in `lib/`).
- `@executable_path/lib`: Searches a nested `lib/` directory.
- `@loader_path/../third_party/credentio/lib`: Searches the repository staging path during local development.

Linux builds embed matching `$ORIGIN` search paths (`$ORIGIN`, `$ORIGIN/../lib`, `$ORIGIN/lib`).

### Inspecting Binary Linkage
Verify embedded RPATHs on macOS using `otool`:

````bash
# Verify shared library dependencies
otool -L bin/credentialctl

# Verify embedded load commands and RPATHs
otool -l bin/credentialctl | grep -A2 LC_RPATH
````

## Install Docs and Channel Alignment

Installation instructions appear in multiple locations across the repository:
- `README.md` (`## Quick Installation` and Homebrew installation blocks)
- `docs/user_guide.md` (Forensic CLI workflows and troubleshooting)

When modifying packaging formats or Homebrew formula parameters, update all documentation files simultaneously to maintain consistency.

## Pre-Flight Release Checklist

Complete these pre-flight validation steps before tagging a release:

- [ ] Working tree is clean on `main` branch (`git status`)
- [ ] Test suite passes cleanly (`make test`)
- [ ] Binary compiles and links against `libcredentio_c` (`make build`)
- [ ] Version and engine metadata report correctly (`./bin/credentialctl version --json`)
- [ ] Snapshot packaging generates valid archives with bundled shared libraries (`make release-snapshot`)

## Step-by-Step Release Guide

Follow these steps to publish a new release:

### 1. Ensure `main` Is Clean and Up to Date

Check out the primary branch and verify all local modifications are committed:

````bash
git checkout main
git pull origin main
git status # Working tree must be clean
````

### 2. Run Test Suite and Local Build

Execute the full test suite and verify binary compilation:

````bash
make test
make build
````

Verify version and engine metadata output:

````bash
./bin/credentialctl version
./bin/credentialctl version --json
````

### 3. Verify Local Snapshot Packaging

Test the GoReleaser packaging pipeline locally in snapshot mode without pushing to GitHub:

````bash
make release-snapshot
````

Inspect generated archives in the `dist/` folder:

````bash
# Verify generated archive contents
tar -tzvf dist/credentialctl_*_darwin_arm64.tar.gz

# Confirm presence of executable and shared library
# Expected: credentialctl, libcredentio_c.dylib, README.md, LICENSE
````

### 4. Create and Push Git Tag

Create an annotated Git tag matching semantic versioning conventions (`vMAJOR.MINOR.PATCH`) and push it to origin:

````bash
git tag -a v0.1.5 -m "Release v0.1.5: Credentio C-ABI v0.1.5 integration and RPATH automation"
git push origin v0.1.5
````

Pushing this tag triggers `.github/workflows/release.yaml` automatically.

### 5. Monitor GitHub Actions Workflow

Track workflow execution in the GitHub Actions web interface or through GitHub CLI:

````bash
gh run watch
````

The workflow typically completes within two minutes.

### 6. Verify Distribution Channels

After the workflow succeeds, verify all distribution targets:

| Channel | Verification Procedure |
|---|---|
| **GitHub Release** | Navigate to `https://github.com/ghchinoy/credentialctl/releases/tag/v0.1.5`. Verify release notes, `credentialctl_0.1.5_darwin_arm64.tar.gz`, and `checksums.txt`. |
| **Standalone Tarball** | Download archive, extract to temporary folder, and run `./credentialctl version --json`. Verify `credentio_engine_version` reports `0.1.5`. |
| **Homebrew Tap** | Update tap and install package: `brew update && brew upgrade credentialctl` (or `brew install ghchinoy/tap/credentialctl`). |

## Homebrew Tap Verification Protocol

Verify Homebrew package installation and functionality with this sequence:

````bash
# Update tap index
brew update

# Add tap repository (if not already tapped)
brew tap ghchinoy/tap

# Install or upgrade package
brew install credentialctl

# Run Homebrew formula tests
brew test credentialctl

# Verify installed binary and C-ABI engine version
credentialctl version --json
````

Confirm that Homebrew places the binary in `$(brew --prefix)/bin/credentialctl` and the shared library in `$(brew --prefix)/lib/libcredentio_c.dylib`.

## Troubleshooting and Manual Fallback

### Failed GitHub Actions Release Job
If the GitHub Actions workflow fails:
1. Open the failed run in GitHub Actions to identify the failing step.
2. If the failure resulted from an expired or invalid `HOMEBREW_TAP_GITHUB_TOKEN`, update the repository secret under Repository Settings -> Secrets and Variables -> Actions.
3. Rerun the failed workflow jobs from the Actions interface.

### Manual Homebrew Formula Update
If GoReleaser fails to push the updated formula to `ghchinoy/homebrew-tap`:
1. Calculate the SHA-256 hash of the release archive:
   ````bash
   shasum -a 256 dist/credentialctl_0.1.5_darwin_arm64.tar.gz
   ````
2. Clone `ghchinoy/homebrew-tap` locally:
   ````bash
   git clone https://github.com/ghchinoy/homebrew-tap.git
   cd homebrew-tap
   ````
3. Edit `Formula/credentialctl.rb` with the new version URL and computed SHA-256 checksum:
   ````ruby
   class Credentialctl < Formula
     desc "C2PA Content Credentials validation and inspection tool with interactive TUI"
     homepage "https://github.com/ghchinoy/credentialctl"
     version "0.1.5"
     license "Apache-2.0"

     if OS.mac? && Hardware::CPU.arm?
       url "https://github.com/ghchinoy/credentialctl/releases/download/v0.1.5/credentialctl_0.1.5_darwin_arm64.tar.gz"
       sha256 "<computed_sha256_hash>"
     end

     def install
       bin.install "credentialctl"
       if OS.mac?
         lib.install "libcredentio_c.dylib"
       elsif OS.linux?
         lib.install "libcredentio_c.so"
       end
     end

     test do
       assert_match "credentialctl version", shell_output("#{bin}/credentialctl --version")
     end
   end
   ````
4. Commit and push changes:
   ````bash
   git add Formula/credentialctl.rb
   git commit -m "credentialctl 0.1.5"
   git push origin main
   ````

### Dynamic Linker Loading Errors
If executing the binary produces `dyld: Library not loaded: @rpath/libcredentio_c.dylib`:
- Confirm that `libcredentio_c.dylib` resides in the same directory as the standalone binary or in `../lib` relative to the binary in Homebrew installations.
- Run `otool -L <path_to_binary>` to verify reference paths.
- Run `DYLD_PRINT_RPATHS=1 <path_to_binary>` to print dynamic linker search path resolution logs.

## Rollback Procedures

If a published release contains critical defects:

### 1. Remove Git Tag
Delete the local and remote Git tag to prevent subsequent accidental builds:

````bash
# Delete local tag
git tag -d v0.1.5

# Delete remote tag
git push --delete origin v0.1.5
````

### 2. Retract or Delete GitHub Release
- Navigate to GitHub Releases, select `v0.1.5`, and click **Edit Release**.
- Select **Delete release** or check **Set as a pre-release** to prevent automated downloads.
- Alternatively, use GitHub CLI:
   ````bash
   gh release delete v0.1.5 --yes
   ````

### 3. Revert Homebrew Tap Formula
Revert the formula in `ghchinoy/homebrew-tap` to the previous stable release commit:

````bash
cd /path/to/homebrew-tap
git revert HEAD -m "Revert credentialctl to v0.1.4"
git push origin main
````

### 4. Publish Replacement Patch Release
Fix the underlying issue in `main`, verify all tests pass, and publish a new incremented patch version (such as `v0.1.6`). Do not reuse retracted version tags.

## Security and Supply Chain Assurance

Maintain these security controls across the release pipeline:

- **Cryptographic Checksums**: Every release publishes `checksums.txt` containing SHA-256 digests of all archives. Always verify archive checksums before deploying binaries in automated infrastructure.
- **Principle of Least Privilege**: GitHub Actions workflow permissions are restricted to `contents: write`. Personal access tokens for Homebrew tap updates are scoped strictly to the `ghchinoy/homebrew-tap` repository.
- **Native Dependency Provenance**: Precompiled dynamic libraries originate directly from signed releases in `ghchinoy/credentio-contributions`.
