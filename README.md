# Nmap HTML Converter

A modern, self-contained tool that converts Nmap XML scan results into beautiful, interactive HTML security reports.

[![GitHub release](https://img.shields.io/github/v/release/dl1rich/NmapHTMLConverter)](https://github.com/dl1rich/NmapHTMLConverter/releases)
[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://github.com/dl1rich/NmapHTMLConverter/workflows/Build%20and%20Test/badge.svg)](https://github.com/dl1rich/NmapHTMLConverter/actions)

## Features

✨ **Modern Design**
- Professional dark theme with glassmorphism effects
- Responsive layout that works on all devices
- Interactive search and filtering
- Smooth animations and hover effects

🔍 **Enhanced Functionality**
- Real-time search across hosts, ports, and services
- Quick statistics overlay
- Click port numbers to copy `IP:PORT` to clipboard for instant testing
- Click port rows to view detailed nmap script output in modal
- Clickable host headers for quick expand/collapse
- Export scan summary to clipboard
- Color-coded port states (🟢 open, 🔴 closed, 🟡 filtered)
- Script output indicator (📋) for ports with additional scan data

🚀 **Self-Contained**
- Single executable with embedded CSS and templates
- No external dependencies required
- Custom CSS and template support for advanced users
- Cross-platform compatibility

## Screenshots

Main  
<p align="center">
<img width="2117" height="1330" alt="image" src="https://github.com/user-attachments/assets/7f968503-b592-4a11-b237-7552f28214b4" />

Host Section  
<img width="700" alt="image" src="https://github.com/user-attachments/assets/3c1f5606-2f08-4ce7-af61-13830342b241" />

Scan Details  
<img width="700"  alt="image" src="https://github.com/user-attachments/assets/028b574b-324b-4264-981e-ece898adfbb6" />
</p>

## Quick Start

### Download Pre-built Binary

Download the latest release for your platform from the [Releases](https://github.com/dl1rich/NmapHTMLConverter/releases) page:

- **Windows**: `nmapHTMLConverter.exe` (amd64) or `nmapHTMLConverter-arm64.exe` (arm64)
- **Linux**: `nmapHTMLConverter-linux` (amd64) or `nmapHTMLConverter-linux-arm64` (arm64)
- **macOS**: `nmapHTMLConverter-mac` (Intel) or `nmapHTMLConverter-mac-arm64` (Apple Silicon)

```bash
# Make the binary executable (Linux/macOS only)
chmod +x nmapHTMLConverter-linux

# Convert your Nmap XML to HTML
./nmapHTMLConverter -xml scan-results.xml
```

That's it! Your `nmap.html` report is ready to view in any browser.

### Building from Source

```bash
# Clone the repository
git clone https://github.com/dl1rich/NmapHTMLConverter.git
cd NmapHTMLConverter

# Build the executable
go build -o nmapHTMLConverter.exe .

# Run with your Nmap XML file
./nmapHTMLConverter -xml your-scan.xml
```

## Usage

### Basic Usage
```bash
# Convert XML to HTML (uses embedded styling)
./nmapHTMLConverter -xml scan-results.xml

# Specify custom output filename
./nmapHTMLConverter -xml scan-results.xml -out security-report.html

# Use stdin (for piping)
cat scan-results.xml | ./nmapHTMLConverter
```

### Advanced Usage
```bash
# Use custom CSS styling
./nmapHTMLConverter -xml scan-results.xml -css custom-style.css

# Use custom HTML template
./nmapHTMLConverter -xml scan-results.xml -tpl custom-template.html

# Combine custom CSS and template
./nmapHTMLConverter -xml scan-results.xml -css style.css -tpl template.html
```

### Command Line Options
```
  -xml string
        input nmap XML file (default: stdin)
  -out string
        output HTML file (default "nmap.html")
  -css string
        custom CSS file (optional, uses embedded CSS by default)
  -tpl string
        custom HTML template file (optional, uses embedded template by default)
```

## Generating Nmap XML

To create XML files compatible with this converter, use the `-oX` option with Nmap:

```bash
# Basic scan with XML output
nmap -oX scan-results.xml target.com

# Comprehensive scan with service detection
nmap -sS -sV -sC -O -oX detailed-scan.xml target.com

# Scan multiple targets
nmap -oX network-scan.xml 192.168.1.0/24

# Quick scan of common ports
nmap --top-ports 1000 -oX quick-scan.xml target.com
```

## Features Overview

### 🎨 **Visual Design**
- **Dark Theme**: Professional security-focused color scheme
- **Glassmorphism**: Modern translucent card design with blur effects
- **Responsive Grid**: Adaptive layout that scales from mobile to desktop
- **Typography**: Clean, readable fonts with proper hierarchy

### 🔍 **Interactive Features**
- **Global Search**: Find hosts, ports, services, or any text instantly
- **Quick Stats**: Overlay showing scan summary and metrics
- **Expand/Collapse**: Toggle host details individually or in bulk - click anywhere on the host header!
- **Export Function**: Copy scan summary to clipboard for reports
- **Port Interaction**:
  - Click **port number** → Copies `IP:PORT` to clipboard (e.g., `10.40.18.20:22`)
  - Click **anywhere else on row** → Opens detailed modal with nmap script output
  - 📋 indicator shows ports with additional scan data (ssl-cert, http-methods, ssh-hostkey, etc.)

### 📊 **Data Presentation**
- **Host Cards**: Clean, organized display of each scanned host
- **Port Tables**: Detailed service information with proper spacing
- **Status Indicators**: Visual cues for host and port states
- **Service Details**: Product versions and additional information

### ⌨️ **Keyboard Shortcuts**
- `/` - Focus search bar
- `Escape` - Clear search or blur search bar

## Customization

### Custom CSS
Create your own styling by providing a CSS file:

```css
/* custom-style.css */
:root {
  --accent: #ff6b6b;  /* Change accent color */
  --bg: #2d3748;      /* Change background */
}

.host-card {
  border-radius: 20px;  /* More rounded corners */
}
```

### Custom Templates
The template system uses Go's `html/template` package. See the embedded template in the source code for reference structure.

Required template definitions:
- `{{define "header"}}` - Page header and opening
- `{{define "host"}}` - Individual host display
- `{{define "footer"}}` - Page footer and closing

## Security Considerations

- This tool processes XML files locally and does not transmit data
- Generated HTML reports are static files safe for sharing
- No external resources are loaded (fully offline capable)
- Input validation prevents XML-based attacks

## Browser Compatibility

- ✅ Chrome/Chromium 80+
- ✅ Firefox 75+
- ✅ Safari 13+
- ✅ Edge 80+

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Creating a Release

To create a new release with automated builds:

1. Update version in code and CHANGELOG.md
2. Commit all changes
3. Create and push a version tag:
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```
4. GitHub Actions will automatically:
   - Build executables for Windows, Linux, and macOS (amd64 and arm64)
   - Create a new release with all binaries attached
   - Generate release notes from commits

The release will appear at: `https://github.com/dl1rich/NmapHTMLConverter/releases`

## Note to self for retag (keep version 1.0.0 for now)

```sh
# Delete tag locally and remotely
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0

# Commit changes (if needed)
git add .
git commit -m "Your changes"
git push

# Recreate and push tag
git tag -a v1.0.0 -m "Release message"
git push origin v1.0.0
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Credits

**Created by Richard Jones**  
**DefenceLogic.io**

## Changelog

### v1.0.0
- Initial release with embedded assets
- Modern responsive design
- Interactive search and filtering
- Real-time statistics
- Export functionality
- Self-contained executable

---

**⚡ Convert your Nmap scans into professional security reports in seconds!**
