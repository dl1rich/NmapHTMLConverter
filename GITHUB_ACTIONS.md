# GitHub Actions Workflows

This repository includes automated CI/CD workflows for building and releasing the Nmap HTML Converter.

## Workflows

### 1. Build and Test (`build.yml`)
**Triggers:** Push to main/master branch, Pull Requests

**Purpose:** Validates that the code compiles across all platforms

**Matrix Build:**
- **OS:** Windows, Linux, macOS
- **Architecture:** amd64, arm64

**Actions:**
- Checks out code
- Sets up Go 1.21
- Compiles for all platform combinations
- Runs tests (Linux amd64 only)

### 2. Build and Release (`release.yml`)
**Triggers:** Push of version tags (e.g., `v1.0.0`, `v2.1.3`)

**Purpose:** Automatically builds and publishes releases

**Build Matrix:**
| Platform | Architecture | Output File |
|----------|-------------|-------------|
| Windows | amd64 | `nmapHTMLConverter.exe` |
| Windows | arm64 | `nmapHTMLConverter-arm64.exe` |
| Linux | amd64 | `nmapHTMLConverter-linux` |
| Linux | arm64 | `nmapHTMLConverter-linux-arm64` |
| macOS | amd64 (Intel) | `nmapHTMLConverter-mac` |
| macOS | arm64 (Apple Silicon) | `nmapHTMLConverter-mac-arm64` |

**Build Optimizations:**
- `CGO_ENABLED=0` - Pure Go, no C dependencies
- `-ldflags="-s -w"` - Strips debug info, reduces binary size

**Release Process:**
1. Builds all platform binaries in parallel
2. Uploads artifacts
3. Creates GitHub release
4. Attaches all binaries to the release
5. Auto-generates release notes from commits

## Usage

### Creating a New Release

1. **Update version references:**
   ```bash
   # Update CHANGELOG.md with new version
   # Update version constant in code if applicable
   ```

2. **Commit changes:**
   ```bash
   git add .
   git commit -m "Prepare v1.0.0 release"
   git push origin main
   ```

3. **Create and push tag:**
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

4. **Wait for automation:**
   - GitHub Actions will automatically build all binaries
   - Release will be created at: `https://github.com/dl1rich/NmapHTMLConverter/releases`
   - All 6 platform binaries will be attached

### Version Tag Format

Tags must follow the pattern: `v*` (e.g., `v1.0.0`, `v2.1.3`, `v1.0.0-beta`)

Examples:
- ✅ `v1.0.0` - Stable release
- ✅ `v2.1.3` - Patch release
- ✅ `v1.0.0-rc1` - Release candidate
- ✅ `v1.5.0-beta` - Beta release
- ❌ `1.0.0` - Missing 'v' prefix
- ❌ `release-1.0.0` - Wrong format

### Monitoring Builds

1. Go to **Actions** tab in GitHub repository
2. Click on the workflow run (triggered by your tag push)
3. View build progress for each platform
4. Download artifacts manually if needed (before release is published)

### Troubleshooting

**Build fails on specific platform:**
- Check the Actions logs for that platform matrix job
- Verify Go code is platform-compatible
- Ensure no OS-specific dependencies are missing

**Release not created:**
- Verify tag format matches `v*` pattern
- Check GitHub token permissions (GITHUB_TOKEN needs `contents: write`)
- Review release workflow logs

**Binary doesn't work on target platform:**
- Verify correct GOOS/GOARCH combination
- Ensure CGO_ENABLED=0 for pure Go binaries
- Check if any C dependencies were accidentally included

## Security Notes

- Workflows use official GitHub Actions (`actions/checkout@v4`, `actions/setup-go@v5`)
- GITHUB_TOKEN is automatically provided by GitHub (no secrets needed)
- Binaries are built in isolated GitHub-hosted runners
- No external dependencies or custom scripts executed

## Customization

### Adding New Platforms

Edit `.github/workflows/release.yml` matrix to add platforms:

```yaml
- goos: freebsd
  goarch: amd64
  output: nmapHTMLConverter-freebsd
```

### Changing Go Version

Update in both workflow files:

```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.22'  # Change here
```

### Custom Build Flags

Modify the build step:

```yaml
- name: Build binary
  env:
    GOOS: ${{ matrix.goos }}
    GOARCH: ${{ matrix.goarch }}
    CGO_ENABLED: 0
  run: |
    go build -ldflags="-s -w -X main.version=${{ github.ref_name }}" -o ${{ matrix.output }} .
```

## File Locations

```
.github/
└── workflows/
    ├── build.yml      # CI build and test workflow
    └── release.yml    # Release automation workflow
```

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Cross Compilation](https://go.dev/doc/install/source#environment)
- [Semantic Versioning](https://semver.org/)
