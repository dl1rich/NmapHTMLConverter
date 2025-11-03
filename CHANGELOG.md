# Changelog

All notable changes to the Nmap HTML Converter project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-11-03

### Added
- ✨ Modern, responsive HTML report generation from Nmap XML
- 🎨 Professional dark theme with glassmorphism effects
- 🔍 Real-time search and filtering across hosts, ports, and services
- 📊 Interactive statistics overlay with scan metrics
- 🚀 Self-contained executable with embedded CSS and templates
- 📱 Mobile-responsive design that works on all devices
- ⌨️ Keyboard shortcuts for improved usability
- 📋 Export functionality to copy scan summaries
- 🎛️ Expandable/collapsible host details
- 🟢 Color-coded port states (open/closed/filtered)
- 🔧 Custom CSS and template support for advanced users
- 📝 Comprehensive documentation and examples
- 🏗️ Cross-platform build system with Makefile
- 📄 MIT license for open source distribution

### Features
- **Interactive Elements**
  - Global search across all scan data
  - Quick stats overlay with host/port/service counts
  - Bulk expand/collapse functionality
  - Smooth animations and hover effects
  
- **Modern Design**
  - Dark theme optimized for security professionals
  - Glassmorphism cards with backdrop blur effects
  - Responsive grid layout (mobile to desktop)
  - Professional typography with proper hierarchy
  
- **Technical Excellence**
  - Single executable with no external dependencies
  - Embedded CSS and HTML templates
  - Efficient XML streaming for large files
  - Cross-platform compatibility (Windows, Linux, macOS)
  - Input validation and security considerations

### Command Line Interface
- `-xml` - Input Nmap XML file (or stdin)
- `-out` - Output HTML filename (default: nmap.html)
- `-css` - Custom CSS override (optional)
- `-tpl` - Custom template override (optional)
- `-version` - Show version information
- `-h` - Show help and usage examples

### Browser Support
- Chrome/Chromium 80+
- Firefox 75+
- Safari 13+
- Edge 80+

### Security
- Local processing only (no data transmission)
- Static HTML output safe for sharing
- No external resource dependencies
- XML input validation

---

**Created by Richard Jones - DefenceLogic.io**

For the latest updates and releases, visit: [GitHub Repository](https://github.com/defencelogic/nmap-html-converter)