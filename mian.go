/*
Nmap HTML Converter
A modern, self-contained tool that converts Nmap XML scan results
into beautiful, interactive HTML security reports.

Created by: Richard Jones
Company: DefenceLogic.io
Version: 1.0.0
License: MIT

Build: go build -o nmapHTMLConverter.exe .
Usage: ./nmapHTMLConverter -xml scan-results.xml
*/
package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// Minimal structs for decoding <host> elements we care about
type NmapRunInfo struct {
	XMLName   xml.Name `xml:"nmaprun"`
	Scanner   string   `xml:"scanner,attr"`
	StartStr  string   `xml:"startstr,attr"`
	Args      string   `xml:"args,attr"`
	StartTime string   `xml:"start,attr"`
}

type Host struct {
	XMLName   xml.Name  `xml:"host"`
	Addresses []Address `xml:"address"`
	Hostnames Hostnames `xml:"hostnames"`
	Ports     Ports     `xml:"ports"`
	Status    Status    `xml:"status"`
}

type Address struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

type Hostnames struct {
	Names []Hostname `xml:"hostname"`
}

type Hostname struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type Ports struct {
	Ports []Port `xml:"port"`
}

type Port struct {
	Protocol string   `xml:"protocol,attr"`
	PortId   int      `xml:"portid,attr"`
	State    State    `xml:"state"`
	Service  Service  `xml:"service"`
	Scripts  []Script `xml:"script"`
}

type State struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
}

type Script struct {
	ID     string `xml:"id,attr"`
	Output string `xml:"output,attr"`
}

type Service struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
	Extras  string `xml:"extrainfo,attr"`
}

type Status struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
}

type TemplateData struct {
	Info      NmapRunInfo
	CSS       template.CSS
	Generated time.Time
}

// Embedded default CSS
const defaultCSS = `:root{
  --bg:#0f1720;
  --card:#0b1220;
  --muted:#9aa4b2;
  --accent:#38bdf8; /* sky-400 */
  --accent-2:#60a5fa;
  --danger:#fb7185;
  --success:#10b981;
  --warning:#f59e0b;
  --glass: rgba(255,255,255,0.03);
  --glass-strong: rgba(255,255,255,0.08);
  --radius:12px;
  --text: #e6eef8;
  --surface: #081020;
  --border: rgba(255,255,255,0.08);
  font-family: "SF Pro Display", Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial;
  font-size: 15px;
  line-height: 1.6;
  -webkit-font-smoothing:antialiased;
  -moz-osx-font-smoothing:grayscale;
}

*{box-sizing:border-box}
html,body{height:100%;margin:0;background:linear-gradient(135deg,#071021 0%, #0a1628 25%, #071725 60%, #0c1b2e 100%);color:var(--text);scroll-behavior: smooth}
body::before{content:"";position:fixed;top:0;left:0;right:0;bottom:0;background:radial-gradient(circle at 30% 20%, rgba(56,189,248,0.03) 0%, transparent 50%), radial-gradient(circle at 70% 80%, rgba(168,85,247,0.02) 0%, transparent 50%);pointer-events:none;z-index:-1}
.container{max-width:1200px;margin:20px auto;padding:0 18px}

/* Topbar */
.topbar{background:linear-gradient(90deg, rgba(255,255,255,0.04), rgba(255,255,255,0.02));backdrop-filter: blur(12px);border-radius:16px;padding:16px;margin:12px 0;border:1px solid var(--border);box-shadow:0 8px 32px rgba(0,0,0,0.4)}
.topbar .container{display:flex;gap:16px;align-items:center;justify-content:space-between;flex-wrap:wrap}
.brand{display:flex;gap:16px;align-items:center}
.logo{width:48px;height:48px;fill:none;stroke:var(--accent);stroke-width:1.8;filter:drop-shadow(0 0 8px rgba(56,189,248,0.3))}
.brand h1{margin:0;font-size:22px;letter-spacing:-0.3px;font-weight:700;background:linear-gradient(135deg, var(--accent), var(--accent-2));-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text}
.muted{color:var(--muted);font-size:13px;font-weight:500}
.controls{display:flex;gap:12px;align-items:center;flex-wrap:wrap}

/* inputs & buttons */
.search{padding:12px 16px;border-radius:12px;border:1px solid var(--border);background:var(--glass-strong);color:var(--text);min-width:280px;font-size:14px;transition:all 0.3s ease;backdrop-filter:blur(8px)}
.search:focus{outline:none;border-color:var(--accent);box-shadow:0 0 0 3px rgba(56,189,248,0.1), 0 4px 12px rgba(0,0,0,0.3);background:var(--glass)}
.search::placeholder{color:var(--muted)}
.btn{background:linear-gradient(135deg,var(--accent),var(--accent-2));border:none;color:#022;padding:12px 16px;border-radius:12px;cursor:pointer;box-shadow:0 4px 12px rgba(56,189,248,0.3);font-weight:600;transition:all 0.2s ease;font-size:14px}
.btn.small{padding:8px 12px;font-size:13px;border-radius:10px}
.btn:hover{transform:translateY(-1px);box-shadow:0 6px 20px rgba(56,189,248,0.4)}
.btn:active{transform:translateY(0);box-shadow:0 2px 8px rgba(56,189,248,0.3)}
.btn.secondary{background:var(--glass-strong);color:var(--text);border:1px solid var(--border)}
.btn.secondary:hover{background:rgba(255,255,255,0.1)}

/* summary */
.summary{display:flex;justify-content:space-between;align-items:center;margin:20px 0;padding:16px 20px;background:var(--glass-strong);border-radius:14px;border:1px solid var(--border);backdrop-filter:blur(8px)}
.summary-stats{display:flex;gap:12px;flex-wrap:wrap}
.summary-stats .pill{background:var(--glass);padding:8px 14px;border-radius:999px;border:1px solid var(--border);font-weight:500;transition:all 0.2s ease}
.summary-stats .pill:hover{background:var(--glass-strong);transform:translateY(-1px)}

/* hosts grid */
.hosts-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(520px,1fr));gap:20px;margin-top:24px}
.host-card{background:linear-gradient(145deg,rgba(255,255,255,0.02),rgba(255,255,255,0.04));border-radius:16px;padding:24px;border:1px solid var(--border);box-shadow:0 8px 32px rgba(0,0,0,0.3);backdrop-filter:blur(8px);transition:all 0.3s ease;position:relative;overflow:hidden}
.host-card::before{content:"";position:absolute;top:0;left:0;right:0;height:3px;background:linear-gradient(90deg,var(--accent),var(--accent-2));opacity:0;transition:opacity 0.3s ease}
.host-card:hover{transform:translateY(-2px);box-shadow:0 12px 40px rgba(0,0,0,0.4);border-color:rgba(56,189,248,0.2)}
.host-card:hover::before{opacity:1}
.host-card[data-status="down"]{opacity:0.6;filter:grayscale(20%)}
.host-head{display:flex;justify-content:space-between;align-items:flex-start;gap:16px;cursor:pointer;user-select:none;padding:4px;margin:-4px;border-radius:12px;transition:background 0.2s ease}
.host-head:hover{background:rgba(56,189,248,0.05)}
.host-title{display:flex;flex-direction:column;gap:10px;flex:1;pointer-events:none}
.host-name{display:flex;gap:12px;align-items:baseline;flex-wrap:wrap}
.ip{font-family: "SF Mono",Menlo,"Monaco","Cascadia Code","Roboto Mono",Courier New,monospace;background:linear-gradient(135deg,#071a2b,#062033);padding:10px 14px;border-radius:10px;border:1px solid var(--border);font-size:15px;font-weight:600;letter-spacing:0.5px}
.hostname{font-size:14px;color:var(--muted);font-weight:500}
.host-badges{display:flex;gap:10px;align-items:center;flex-wrap:wrap}
.badge{padding:7px 12px;border-radius:20px;font-weight:600;background:var(--glass);font-size:12px;border:1px solid var(--border);transition:all 0.2s ease}
.state-up{color:#10b981;background:linear-gradient(90deg, rgba(16,185,129,0.08), rgba(16,185,129,0.04));border-color:rgba(16,185,129,0.2)}
.state-down{color:var(--danger);background:linear-gradient(90deg, rgba(251,113,133,0.08), rgba(251,113,133,0.04));border-color:rgba(251,113,133,0.2)}
.ports-count{background:var(--glass-strong);color:var(--accent);border-color:rgba(56,189,248,0.2)}

/* host body */
.host-body{margin-top:20px;padding-top:20px;border-top:1px solid var(--border);animation:slideDown 0.3s ease-out}
.host-meta{margin-bottom:20px}
.host-meta dl{display:grid;grid-template-columns:110px 1fr;gap:10px 20px;margin:0;background:var(--glass);padding:16px;border-radius:10px;border:1px solid var(--border)}
.host-meta dt{color:var(--muted);font-size:12px;font-weight:600;text-transform:uppercase;letter-spacing:0.5px}
.host-meta dd{margin:0;font-family:"SF Mono",monospace;color:var(--text);word-break:break-word;font-size:13px;line-height:1.5}
.host-meta code{background:rgba(56,189,248,0.1);color:var(--accent);padding:3px 7px;border-radius:4px;font-size:12px}

/* ports table */
.ports-table-wrap{overflow:auto;border-radius:12px;background:var(--glass);padding:0;border:1px solid var(--border);backdrop-filter:blur(8px)}
.ports-table{width:100%;border-collapse:collapse;font-size:14px;min-width:650px}
.ports-table thead th{font-size:12px;text-align:left;padding:16px 20px;border-bottom:1px solid var(--border);color:var(--muted);font-weight:600;text-transform:uppercase;letter-spacing:0.5px;background:rgba(255,255,255,0.02);white-space:nowrap}
.ports-table tbody td{padding:16px 20px;border-bottom:1px solid rgba(255,255,255,0.03);transition:background 0.2s ease;vertical-align:top}
.ports-table tbody tr{cursor:pointer}
.ports-table tbody tr:hover{background:rgba(56,189,248,0.08)}
.ports-table tbody tr:last-child td{border-bottom:none}
.p-port{font-weight:700;color:var(--accent);font-family:"SF Mono",monospace;min-width:90px;font-size:15px}
.p-proto{color:var(--muted);text-transform:uppercase;font-size:11px;font-weight:600;min-width:90px}
.p-state{font-weight:600;min-width:110px}
.p-state[data-state="open"]{color:var(--success)}
.p-state[data-state="closed"]{color:var(--danger)}
.p-state[data-state="filtered"]{color:var(--warning)}
.p-service{color:var(--text);font-weight:500;min-width:130px}
.p-product{color:var(--muted);font-size:13px;line-height:1.5;max-width:350px;word-wrap:break-word}

/* footer */
.footer{margin-top:40px;padding:20px;text-align:center;color:var(--muted);font-size:13px;background:var(--glass);border-radius:12px;border:1px solid var(--border)}

/* animations */
@keyframes slideDown{
  from{opacity:0;transform:translateY(-10px)}
  to{opacity:1;transform:translateY(0)}
}

@keyframes fadeIn{
  from{opacity:0}
  to{opacity:1}
}

.host-card{animation:fadeIn 0.5s ease-out}

/* loading states */
.loading{position:relative;overflow:hidden}
.loading::after{content:"";position:absolute;top:0;left:-100%;right:100%;height:100%;background:linear-gradient(90deg,transparent,rgba(255,255,255,0.1),transparent);animation:shimmer 1.5s infinite}

@keyframes shimmer{
  to{left:100%;right:-100%}}

@keyframes slideIn{
  from{opacity:0;transform:translateX(20px)}
  to{opacity:1;transform:translateX(0)}
}

@keyframes slideOut{
  from{opacity:1;transform:translateX(0)}
  to{opacity:0;transform:translateX(20px)}
}

/* utility classes */
.hidden{display:none !important}
.text-center{text-align:center}
.mt-auto{margin-top:auto}
.flex{display:flex}
.items-center{align-items:center}
.justify-between{justify-content:space-between}
.gap-2{gap:8px}
.gap-4{gap:16px}

/* scrollbar styling */
::-webkit-scrollbar{width:8px;height:8px}
::-webkit-scrollbar-track{background:var(--glass)}
::-webkit-scrollbar-thumb{background:var(--border);border-radius:4px}
::-webkit-scrollbar-thumb:hover{background:rgba(255,255,255,0.2)}

/* Advanced filtering */
.filter-bar{display:flex;gap:8px;margin:16px 0;flex-wrap:wrap;align-items:center}
.filter-chip{background:var(--glass);border:1px solid var(--border);padding:6px 12px;border-radius:20px;font-size:12px;cursor:pointer;transition:all 0.2s ease;user-select:none}
.filter-chip.active{background:var(--accent);color:#022;border-color:var(--accent)}
.filter-chip:hover{background:var(--glass-strong)}

/* Risk scoring */
.risk-critical{background:linear-gradient(90deg, rgba(239,68,68,0.1), rgba(239,68,68,0.05));border-color:rgba(239,68,68,0.3);color:#ef4444}
.risk-high{background:linear-gradient(90deg, rgba(251,113,133,0.1), rgba(251,113,133,0.05));border-color:rgba(251,113,133,0.3);color:#fb7185}
.risk-medium{background:linear-gradient(90deg, rgba(245,158,11,0.1), rgba(245,158,11,0.05));border-color:rgba(245,158,11,0.3);color:#f59e0b}
.risk-low{background:linear-gradient(90deg, rgba(34,197,94,0.1), rgba(34,197,94,0.05));border-color:rgba(34,197,94,0.3);color:#22c55e}
.risk-info{background:linear-gradient(90deg, rgba(59,130,246,0.1), rgba(59,130,246,0.05));border-color:rgba(59,130,246,0.3);color:#3b82f6}

/* Service icons */
.service-icon{width:16px;height:16px;margin-right:6px;vertical-align:middle}

/* Advanced stats */
.stats-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:12px;margin:16px 0}
.stat-card{background:var(--glass);border:1px solid var(--border);border-radius:10px;padding:12px;text-align:center}
.stat-number{font-size:24px;font-weight:700;color:var(--accent)}
.stat-label{font-size:12px;color:var(--muted);text-transform:uppercase;letter-spacing:0.5px}

/* Timeline */
.timeline{position:relative;margin:20px 0}
.timeline-item{position:relative;padding:10px 0 10px 30px;border-left:2px solid var(--border)}
.timeline-item::before{content:"";position:absolute;left:-5px;top:15px;width:8px;height:8px;background:var(--accent);border-radius:50%}
.timeline-time{font-size:11px;color:var(--muted);font-weight:600}

/* Port details modal */
.modal{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,0.8);z-index:10000;display:none;align-items:center;justify-content:center}
.modal.show{display:flex}
.modal-content{background:var(--card);border-radius:16px;padding:24px;max-width:800px;width:90%;max-height:80vh;overflow:auto;border:1px solid var(--border);box-shadow:0 20px 60px rgba(0,0,0,0.6)}
.modal-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:20px}
.modal-close{background:none;border:none;color:var(--muted);font-size:24px;cursor:pointer;padding:0;width:30px;height:30px;display:flex;align-items:center;justify-content:center}
.modal-close:hover{color:var(--text)}

/* Network topology */
.topology{margin:20px 0}
.network-segment{background:var(--glass);border:1px solid var(--border);border-radius:10px;padding:12px;margin:8px 0}
.network-title{font-weight:600;color:var(--accent);margin-bottom:8px}
.host-list{display:flex;flex-wrap:wrap;gap:6px}
.host-mini{background:var(--glass-strong);padding:4px 8px;border-radius:6px;font-size:11px;border:1px solid var(--border)}

/* Export options */
.export-menu{position:absolute;top:100%;right:0;background:var(--card);border:1px solid var(--border);border-radius:8px;padding:8px;min-width:160px;box-shadow:0 8px 32px rgba(0,0,0,0.4);z-index:1000}
.export-item{display:block;width:100%;padding:8px 12px;background:none;border:none;color:var(--text);text-align:left;border-radius:6px;cursor:pointer;font-size:13px}
.export-item:hover{background:var(--glass)}

/* responsive tweaks */
@media (max-width:720px){
  .topbar .container{flex-direction:column;align-items:stretch}
  .controls{justify-content:space-between}
  .search{width:100%}
  .hosts-grid{grid-template-columns:1fr}
  .ports-table{min-width:auto;font-size:12px}
  .ports-table thead th, .ports-table tbody td{padding:10px 14px}
}

@media (min-width:1500px){
  .container{max-width:1500px}
  .hosts-grid{grid-template-columns:repeat(auto-fill,minmax(650px,1fr))}
  .host-card{padding:28px}
  .ports-table{min-width:700px}
}

/* Print styles */
@media print{
  body{background:white !important;color:#000}
  body::before{display:none}
  .topbar .controls, .btn, #statsOverlay{display:none !important}
  .host-card{page-break-inside:avoid;border:1px solid #ddd;box-shadow:none}
  .host-body{display:block !important}
  .host-body[hidden]{display:block !important}
  .footer{page-break-before:always}
  .search{display:none}
}`

// Embedded default template
const defaultTemplate = `{{define "header"}}
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width,initial-scale=1"/>
  <title>Nmap Security Report</title>
  <meta name="description" content="Network security scan report generated by Nmap"/>
  
  <style>{{.CSS}}</style>
</head>
<body>
  <header class="topbar">
    <div class="container">
      <div class="brand">
        <svg class="logo" viewBox="0 0 24 24" aria-hidden="true">
          <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
        </svg>
        <div>
          <h1>Network Security Report</h1>
          <p class="muted">Scanner: {{.Info.Scanner}} • Generated: {{.Generated.Format "Jan 2, 2006 15:04"}}</p>
        </div>
      </div>

      <div class="controls">
        <input id="globalSearch" class="search" placeholder="🔍 Search hosts, ports, services..." />
        <div style="position:relative;">
          <button id="exportMenu" class="btn small secondary">📊 Export ⏷</button>
          <div id="exportDropdown" class="export-menu" style="display:none;">
            <button class="export-item" onclick="exportCSV()">📄 Export CSV</button>
            <button class="export-item" onclick="exportJSON()">📋 Export JSON</button>
            <button class="export-item" onclick="exportPDF()">📑 Save as PDF</button>
            <button class="export-item" onclick="exportSummary()">📊 Copy Summary</button>
          </div>
        </div>
        <button id="collapseAll" class="btn small secondary">Collapse All</button>
        <button id="expandAll" class="btn small">Expand All</button>
      </div>
    </div>
  </header>

  <main class="container">
    <!-- Advanced Filter Bar -->
    <div class="filter-bar">
      <span style="color:var(--muted);font-size:13px;font-weight:600;">Filter by:</span>
      <div class="filter-chip active" data-filter="all">🌐 All Hosts</div>
      <div class="filter-chip" data-filter="up">🟢 Up</div>
      <div class="filter-chip" data-filter="down">🔴 Down</div>
      <div class="filter-chip" data-filter="critical">🚨 Critical</div>
      <div class="filter-chip" data-filter="web">🌍 Web Services</div>
      <div class="filter-chip" data-filter="database">🗄️ Databases</div>
      <div class="filter-chip" data-filter="ssh">🔑 SSH</div>
      <div class="filter-chip" data-filter="windows">🪟 Windows</div>
    </div>

    <!-- Enhanced Statistics -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-number" id="totalHostsStat">0</div>
        <div class="stat-label">Total Hosts</div>
      </div>
      <div class="stat-card">
        <div class="stat-number" id="vulnerableHostsStat">0</div>
        <div class="stat-label">Vulnerable Hosts</div>
      </div>
      <div class="stat-card">
        <div class="stat-number" id="criticalPortsStat">0</div>
        <div class="stat-label">Critical Services</div>
      </div>
      <div class="stat-card">
        <div class="stat-number" id="riskScoreStat">0</div>
        <div class="stat-label">Risk Score</div>
      </div>
    </div>

    <section class="summary">
      <div>
        <strong>Scan Command:</strong><br/>
        <code style="font-size:12px;color:var(--muted);">{{.Info.Args}}</code>
      </div>
      <div class="summary-stats">
        <span class="pill">📊 Hosts: <strong id="hostCount">—</strong></span>
        <span class="pill">🔓 Open Ports: <strong id="openPortCount">—</strong></span>
        <span class="pill">🛡️ Services: <strong id="serviceCount">—</strong></span>
        <span class="pill">⚠️ Risk Level: <strong id="riskLevel">—</strong></span>
      </div>
    </section>

    <section id="hosts" class="hosts-grid" aria-live="polite">
{{end}}

{{define "host"}}
  <article class="host-card" data-host="{{range .Addresses}}{{.Addr}} {{end}}" data-status="{{.Status.State}}">
    <header class="host-head">
      <div class="host-title">
        <div class="host-name">
          <strong class="ip">
            {{with .Addresses}}{{if gt (len .) 0}}{{(index . 0).Addr}}{{end}}{{end}}
          </strong>
          {{with .Hostnames.Names}}{{if gt (len .) 0}}
          <span class="hostname">
            {{(index . 0).Name}}
          </span>
          {{end}}{{end}}
        </div>
        <div class="host-badges">
          <span class="badge state-{{.Status.State}}">
            {{if eq .Status.State "up"}}🟢{{else}}🔴{{end}} {{.Status.State}}
          </span>
          <span class="badge ports-count">
            📊 {{len .Ports.Ports}} port{{if ne (len .Ports.Ports) 1}}s{{end}}
          </span>
        </div>
      </div>

      <div class="host-actions">
        <button class="btn toggle small secondary" aria-expanded="false">
          <span class="toggle-text">Expand</span>
        </button>
      </div>
    </header>

    <div class="host-body" hidden>
      {{if or .Addresses .Hostnames.Names}}
      <div class="host-meta">
        <dl>
          {{if .Addresses}}
          <dt>Addresses</dt>
          <dd>
            {{range .Addresses}}<code>{{.Addr}}</code> <small class="muted">({{.AddrType}})</small><br/>{{end}}
          </dd>
          {{end}}

          {{if .Hostnames.Names}}
          <dt>Hostnames</dt>
          <dd>
            {{range .Hostnames.Names}}<code>{{.Name}}</code> <small class="muted">({{.Type}})</small><br/>{{end}}
          </dd>
          {{end}}

          <dt>Status</dt>
          <dd>{{.Status.State}} <small class="muted">({{.Status.Reason}})</small></dd>
        </dl>
      </div>
      {{end}}

      {{if .Ports.Ports}}
      <div class="ports-table-wrap">
        <table class="ports-table" role="grid" aria-label="Open ports and services">
          <thead>
            <tr>
              <th>Port</th>
              <th>Protocol</th>
              <th>State</th>
              <th>Service</th>
              <th>Product / Version</th>
            </tr>
          </thead>
          <tbody>
            {{range .Ports.Ports}}
            <tr data-port="{{.PortId}}" 
                data-service="{{.Service.Name}}" 
                data-product="{{.Service.Product}}" 
                data-version="{{.Service.Version}}"
                data-extras="{{.Service.Extras}}"
                data-state="{{.State.State}}"
                data-reason="{{.State.Reason}}"
                data-has-scripts="{{if .Scripts}}true{{else}}false{{end}}"
                onclick="showPortDetails(event, this)">
              <td class="p-port" onclick="event.stopPropagation(); copyPortToClipboard(event, this)">{{.PortId}}{{if .Scripts}}<span style="margin-left:4px;font-size:10px;color:var(--accent)">📋</span>{{end}}</td>
              <td class="p-proto">{{.Protocol}}</td>
              <td class="p-state" data-state="{{.State.State}}">
                {{if eq .State.State "open"}}🟢{{else if eq .State.State "closed"}}🔴{{else}}🟡{{end}} {{.State.State}}
              </td>
              <td class="p-service">
                {{if .Service.Name}}
                  <span class="service-icon">{{if eq .Service.Name "http"}}🌐{{else if eq .Service.Name "https"}}🔒{{else if eq .Service.Name "ssh"}}🔑{{else if eq .Service.Name "ftp"}}📁{{else if eq .Service.Name "mysql"}}🗄️{{else if eq .Service.Name "postgresql"}}🗄️{{else if eq .Service.Name "smtp"}}📧{{else if eq .Service.Name "dns"}}🌐{{else if eq .Service.Name "telnet"}}⚠️{{else if eq .Service.Name "rdp"}}🖥️{{else}}⚙️{{end}}</span>
                  {{.Service.Name}}
                  {{if eq .Service.Name "telnet"}} <span class="badge risk-critical">CRITICAL</span>{{end}}
                  {{if eq .Service.Name "ftp"}} <span class="badge risk-high">HIGH</span>{{end}}
                  {{if and (eq .Service.Name "http") (not .Service.Product)}} <span class="badge risk-medium">MEDIUM</span>{{end}}
                {{else}}-{{end}}
              </td>
              <td class="p-product">
                {{if .Service.Product}}{{.Service.Product}}{{if .Service.Version}} {{.Service.Version}}{{end}}{{if .Service.Extras}} ({{.Service.Extras}}){{end}}{{else}}-{{end}}
              </td>
            </tr>
            {{end}}
          </tbody>
        </table>
      </div>
      
      <!-- Hidden script data for JavaScript access -->
      <div class="port-scripts-data" style="display:none;">
        {{range .Ports.Ports}}
        {{if .Scripts}}
        <div data-port="{{.PortId}}">
          {{range .Scripts}}
          <div class="script-item" data-id="{{.ID}}">{{.Output}}</div>
          {{end}}
        </div>
        {{end}}
        {{end}}
      </div>
      
      {{else}}
      <p class="text-center muted">No open ports detected</p>
      {{end}}
    </div>
  </article>
{{end}}

{{define "footer"}}
    </section>
    
    <div id="statsOverlay" class="hidden" style="position:fixed;top:20px;right:20px;background:var(--glass-strong);backdrop-filter:blur(12px);border:1px solid var(--border);border-radius:12px;padding:16px;z-index:1000;">
      <h4 style="margin:0 0 8px 0;font-size:14px;">Quick Stats</h4>
      <div style="font-size:12px;color:var(--muted);">
        <div>🖥️ Hosts: <span id="totalHosts">0</span></div>
        <div>🟢 Up: <span id="upHosts">0</span></div>
        <div>🔴 Down: <span id="downHosts">0</span></div>
        <div>🔓 Open Ports: <span id="totalOpenPorts">0</span></div>
      </div>
    </div>

    <footer class="footer">
      <div style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:16px;">
        <div>
          <small class="muted">
            🛡️ Network Security Report generated by <strong>DefenceLogic.io</strong><br/>
            📅 Created by Richard Jones • Nmap HTML Converter v1.0.0<br/>
            📅 Exported on {{.Generated.Format "Monday, January 2, 2006 at 15:04:05"}}
          </small>
        </div>
        <div style="display:flex;gap:8px;">
          <button id="toggleStats" class="btn small secondary">📊 Show Stats</button>
          <button id="printReport" class="btn small secondary">🖨️ Print Report</button>
        </div>
      </div>
    </footer>

    <script>
      (function(){
        const search = document.getElementById('globalSearch');
        const hosts = Array.from(document.querySelectorAll('.host-card'));
        const hostCount = document.getElementById('hostCount');
        const openPortCount = document.getElementById('openPortCount');
        const serviceCount = document.getElementById('serviceCount');
        const statsOverlay = document.getElementById('statsOverlay');
        const toggleStatsBtn = document.getElementById('toggleStats');
        const filterChips = document.querySelectorAll('.filter-chip');
        const exportMenu = document.getElementById('exportMenu');
        const exportDropdown = document.getElementById('exportDropdown');

        // Service categorization
        const webServices = ['http', 'https', 'nginx', 'apache', 'iis'];
        const databaseServices = ['mysql', 'postgresql', 'mongodb', 'redis', 'oracle'];
        const criticalServices = ['telnet', 'rlogin', 'rsh'];
        const windowsServices = ['msrpc', 'netbios-ssn', 'microsoft-ds', 'rdp'];

        // Risk scoring function
        function calculateRiskScore(service, version) {
          let score = 0;
          if(criticalServices.includes(service)) score += 50;
          if(service === 'ftp') score += 30;
          if(service === 'ssh' && !version) score += 20;
          if(webServices.includes(service) && !version) score += 15;
          if(service === 'smtp') score += 10;
          return Math.min(score, 100);
        }

        function updateStats(){
          const visibleHosts = hosts.filter(h => !h.classList.contains('hidden-by-filter'));
          const upHosts = hosts.filter(h => h.getAttribute('data-status') === 'up');
          const downHosts = hosts.filter(h => h.getAttribute('data-status') !== 'up');
          
          let totalOpenPorts = 0;
          let vulnerableHosts = 0;
          let criticalPorts = 0;
          let totalRiskScore = 0;
          let uniqueServices = new Set();
          
          hosts.forEach(host => {
            let hostRisk = 0;
            let hostVulnerable = false;
            const portRows = host.querySelectorAll('.ports-table tbody tr');
            
            portRows.forEach(row => {
              const state = row.querySelector('.p-state').textContent.toLowerCase();
              const serviceEl = row.querySelector('.p-service');
              const service = serviceEl ? serviceEl.textContent.trim().replace(/[🌐🔒🔑📁🗄️📧⚠️🖥️⚙️]/g, '').trim() : '';
              const product = row.querySelector('.p-product').textContent.trim();
              
              if(state.includes('open')){
                totalOpenPorts++;
                if(service && service !== '-') {
                  uniqueServices.add(service);
                  const risk = calculateRiskScore(service, product);
                  hostRisk += risk;
                  if(risk >= 30) {
                    criticalPorts++;
                    hostVulnerable = true;
                  }
                }
              }
            });
            
            if(hostVulnerable) vulnerableHosts++;
            totalRiskScore += hostRisk;
          });

          // Update counters
          hostCount.textContent = visibleHosts.length;
          openPortCount.textContent = totalOpenPorts;
          serviceCount.textContent = uniqueServices.size;
          
          // Update advanced stats
          const totalHostsStat = document.getElementById('totalHostsStat');
          const vulnerableHostsStat = document.getElementById('vulnerableHostsStat');
          const criticalPortsStat = document.getElementById('criticalPortsStat');
          const riskScoreStat = document.getElementById('riskScoreStat');
          
          if(totalHostsStat) totalHostsStat.textContent = hosts.length;
          if(vulnerableHostsStat) vulnerableHostsStat.textContent = vulnerableHosts;
          if(criticalPortsStat) criticalPortsStat.textContent = criticalPorts;
          if(riskScoreStat) riskScoreStat.textContent = Math.round(totalRiskScore / hosts.length) || 0;
          
          // Risk level assessment
          const avgRisk = totalRiskScore / hosts.length || 0;
          const riskLevel = document.getElementById('riskLevel');
          if(riskLevel) {
            if(avgRisk >= 40) {
              riskLevel.textContent = '🚨 Critical';
              riskLevel.style.color = '#ef4444';
            } else if(avgRisk >= 25) {
              riskLevel.textContent = '⚠️ High';
              riskLevel.style.color = '#f59e0b';
            } else if(avgRisk >= 10) {
              riskLevel.textContent = '🟡 Medium';
              riskLevel.style.color = '#eab308';
            } else {
              riskLevel.textContent = '🟢 Low';
              riskLevel.style.color = '#22c55e';
            }
          }
          
          // Update overlay stats
          const totalHostsEl = document.getElementById('totalHosts');
          const upHostsEl = document.getElementById('upHosts');
          const downHostsEl = document.getElementById('downHosts');
          const totalOpenPortsEl = document.getElementById('totalOpenPorts');
          
          if(totalHostsEl) totalHostsEl.textContent = hosts.length;
          if(upHostsEl) upHostsEl.textContent = upHosts.length;
          if(downHostsEl) downHostsEl.textContent = downHosts.length;
          if(totalOpenPortsEl) totalOpenPortsEl.textContent = totalOpenPorts;
        }

        function matchesHost(host, query){
          if(!query) return true;
          query = query.toLowerCase();
          
          const hostData = host.getAttribute('data-host') + ' ' + 
                          (host.querySelector('.hostname')?.textContent || '');
          if(hostData.toLowerCase().includes(query)) return true;
          
          const rows = Array.from(host.querySelectorAll('.ports-table tbody tr'));
          return rows.some(row => row.textContent.toLowerCase().includes(query));
        }

        function doFilter(){
          const query = search.value.trim();
          hosts.forEach(host => {
            const matches = matchesHost(host, query);
            host.style.display = matches ? '' : 'none';
            host.classList.toggle('hidden-by-filter', !matches);
          });
          updateStats();
        }

        search.addEventListener('input', doFilter);
        search.addEventListener('keydown', function(e){
          if(e.key === 'Escape') {
            search.value = '';
            doFilter();
          }
        });

        document.body.addEventListener('click', function(e){
          // Toggle on button click
          if(e.target.matches('.toggle') || e.target.closest('.toggle')){
            e.stopPropagation(); // Prevent host-head click from also firing
            const button = e.target.matches('.toggle') ? e.target : e.target.closest('.toggle');
            const card = button.closest('.host-card');
            const body = card.querySelector('.host-body');
            const text = button.querySelector('.toggle-text');
            const isHidden = body.hasAttribute('hidden');
            
            if(isHidden) {
              body.removeAttribute('hidden');
              text.textContent = 'Collapse';
              button.setAttribute('aria-expanded', 'true');
            } else {
              body.setAttribute('hidden','');
              text.textContent = 'Expand';
              button.setAttribute('aria-expanded', 'false');
            }
          }
          
          // Toggle on header click
          if(e.target.matches('.host-head') || e.target.closest('.host-head')){
            const header = e.target.matches('.host-head') ? e.target : e.target.closest('.host-head');
            const card = header.closest('.host-card');
            const body = card.querySelector('.host-body');
            const button = card.querySelector('.toggle');
            const text = button.querySelector('.toggle-text');
            const isHidden = body.hasAttribute('hidden');
            
            if(isHidden) {
              body.removeAttribute('hidden');
              text.textContent = 'Collapse';
              button.setAttribute('aria-expanded', 'true');
            } else {
              body.setAttribute('hidden','');
              text.textContent = 'Expand';
              button.setAttribute('aria-expanded', 'false');
            }
          }
        });

        document.getElementById('collapseAll').addEventListener('click', () => {
          hosts.forEach(host => {
            const body = host.querySelector('.host-body');
            const button = host.querySelector('.toggle');
            const text = button.querySelector('.toggle-text');
            body.setAttribute('hidden','');
            text.textContent = 'Expand';
            button.setAttribute('aria-expanded', 'false');
          });
        });

        document.getElementById('expandAll').addEventListener('click', () => {
          hosts.forEach(host => {
            const body = host.querySelector('.host-body');
            const button = host.querySelector('.toggle');
            const text = button.querySelector('.toggle-text');
            body.removeAttribute('hidden');
            text.textContent = 'Collapse';
            button.setAttribute('aria-expanded', 'true');
          });
        });

        // Advanced filtering
        filterChips.forEach(chip => {
          chip.addEventListener('click', () => {
            filterChips.forEach(c => c.classList.remove('active'));
            chip.classList.add('active');
            
            const filter = chip.dataset.filter;
            hosts.forEach(host => {
              let show = true;
              
              if(filter === 'up') {
                show = host.getAttribute('data-status') === 'up';
              } else if(filter === 'down') {
                show = host.getAttribute('data-status') !== 'up';
              } else if(filter === 'critical') {
                const services = Array.from(host.querySelectorAll('.p-service')).map(el => el.textContent.toLowerCase());
                show = services.some(s => criticalServices.some(cs => s.includes(cs))) || services.includes('telnet');
              } else if(filter === 'web') {
                const services = Array.from(host.querySelectorAll('.p-service')).map(el => el.textContent.toLowerCase());
                show = services.some(s => webServices.some(ws => s.includes(ws)));
              } else if(filter === 'database') {
                const services = Array.from(host.querySelectorAll('.p-service')).map(el => el.textContent.toLowerCase());
                show = services.some(s => databaseServices.some(ds => s.includes(ds)));
              } else if(filter === 'ssh') {
                const services = Array.from(host.querySelectorAll('.p-service')).map(el => el.textContent.toLowerCase());
                show = services.some(s => s.includes('ssh'));
              } else if(filter === 'windows') {
                const services = Array.from(host.querySelectorAll('.p-service')).map(el => el.textContent.toLowerCase());
                show = services.some(s => windowsServices.some(ws => s.includes(ws)));
              }
              
              host.style.display = show ? '' : 'none';
              host.classList.toggle('hidden-by-filter', !show);
            });
            
            updateStats();
          });
        });

        // Export menu toggle
        if(exportMenu) {
          exportMenu.addEventListener('click', (e) => {
            e.stopPropagation();
            exportDropdown.style.display = exportDropdown.style.display === 'none' ? 'block' : 'none';
          });
        }

        // Close export menu when clicking outside
        document.addEventListener('click', () => {
          if(exportDropdown) exportDropdown.style.display = 'none';
        });

        // Copy IP:PORT to clipboard when clicking port number
        window.copyPortToClipboard = function(event, portCell) {
          const row = portCell.closest('tr');
          const port = row.dataset.port;
          const hostCard = row.closest('.host-card');
          const ip = hostCard.querySelector('.ip').textContent.trim();
          const target = ` + "`" + `${ip}:${port}` + "`" + `;
          
          // Copy to clipboard
          navigator.clipboard.writeText(target).then(() => {
            // Show success notification
            const notification = document.createElement('div');
            notification.style.cssText = 'position:fixed;top:20px;right:20px;background:var(--accent);color:#fff;padding:12px 20px;border-radius:8px;box-shadow:0 4px 16px rgba(0,0,0,0.3);z-index:10000;font-size:14px;font-weight:600;animation:slideIn 0.3s ease;';
            notification.textContent = ` + "`" + `✓ Copied: ${target}` + "`" + `;
            document.body.appendChild(notification);
            
            setTimeout(() => {
              notification.style.animation = 'slideOut 0.3s ease';
              setTimeout(() => notification.remove(), 300);
            }, 2000);
          }).catch(err => {
            console.error('Failed to copy:', err);
          });
        };

        // Show port details modal with script output
        window.showPortDetails = function(event, row) {
          const port = row.dataset.port;
          const hostCard = row.closest('.host-card');
          const ip = hostCard.querySelector('.ip').textContent.trim();
          
          // Check if port has script data
          const hasScripts = row.dataset.hasScripts === 'true';
          if (!hasScripts) return; // No modal needed if no scripts
          
          // Find script data for this port
          const scriptContainer = hostCard.querySelector(` + "`" + `.port-scripts-data [data-port="${port}"]` + "`" + `);
          if (!scriptContainer) return;
          
          const scripts = scriptContainer.querySelectorAll('.script-item');
          if (scripts.length === 0) return;
          
          // Build script output HTML
          let scriptsHTML = '';
          scripts.forEach(script => {
            const scriptId = script.dataset.id;
            const output = script.textContent;
            scriptsHTML += ` + "`" + `
              <div style="margin-bottom:16px;">
                <div style="font-weight:600;color:var(--accent);margin-bottom:6px;font-size:13px;">
                  📋 ${scriptId}
                </div>
                <pre style="background:var(--glass);padding:12px;border-radius:6px;overflow-x:auto;font-size:12px;line-height:1.5;margin:0;border:1px solid var(--border);white-space:pre-wrap;word-wrap:break-word;">${output}</pre>
              </div>
            ` + "`" + `;
          });
          
          // Create modal
          const modal = document.createElement('div');
          modal.style.cssText = 'position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,0.8);display:flex;align-items:center;justify-content:center;z-index:10000;padding:20px;animation:fadeIn 0.2s ease;';
          modal.innerHTML = ` + "`" + `
            <div style="background:var(--card);border:1px solid var(--border);border-radius:12px;max-width:800px;width:100%;max-height:80vh;overflow:hidden;display:flex;flex-direction:column;box-shadow:0 8px 32px rgba(0,0,0,0.4);">
              <div style="padding:20px;border-bottom:1px solid var(--border);display:flex;justify-content:space-between;align-items:center;">
                <h3 style="margin:0;font-size:18px;">
                  🔍 Port ${port} Details - ${ip}:${port}
                </h3>
                <button onclick="this.closest('div[style*=fixed]').remove()" style="background:none;border:none;color:var(--muted);font-size:24px;cursor:pointer;padding:0;width:32px;height:32px;display:flex;align-items:center;justify-content:center;border-radius:6px;transition:all 0.2s;" onmouseover="this.style.background='var(--glass)';this.style.color='var(--text)'" onmouseout="this.style.background='none';this.style.color='var(--muted)'">×</button>
              </div>
              <div style="padding:20px;overflow-y:auto;">
                <div style="margin-bottom:16px;padding:12px;background:var(--glass);border-radius:8px;border:1px solid var(--border);">
                  <div style="display:grid;grid-template-columns:120px 1fr;gap:8px;font-size:13px;">
                    <div style="color:var(--muted);">Service:</div>
                    <div style="font-weight:600;">${row.dataset.service || 'Unknown'}</div>
                    <div style="color:var(--muted);">Product:</div>
                    <div>${row.dataset.product || '-'}${row.dataset.version ? ' ' + row.dataset.version : ''}</div>
                    <div style="color:var(--muted);">State:</div>
                    <div style="color:var(--success);font-weight:600;">${row.dataset.state}</div>
                    ${row.dataset.extras ? ` + "`" + `<div style="color:var(--muted);">Extra Info:</div><div>${row.dataset.extras}</div>` + "`" + ` : ''}
                  </div>
                </div>
                <h4 style="margin:0 0 12px 0;font-size:14px;color:var(--muted);text-transform:uppercase;letter-spacing:0.5px;">Script Output</h4>
                ${scriptsHTML}
              </div>
            </div>
          ` + "`" + `;
          
          document.body.appendChild(modal);
          
          // Close on background click
          modal.addEventListener('click', (e) => {
            if (e.target === modal) modal.remove();
          });
        };

        // Export functions
        window.exportCSV = function() {
          const csvData = [];
          csvData.push(['IP Address', 'Hostname', 'Port', 'Protocol', 'State', 'Service', 'Product']);
          
          hosts.forEach(host => {
            const ip = host.querySelector('.ip').textContent;
            const hostname = host.querySelector('.hostname')?.textContent || '';
            const rows = host.querySelectorAll('.ports-table tbody tr');
            
            rows.forEach(row => {
              const cells = row.querySelectorAll('td');
              csvData.push([
                ip, hostname, 
                cells[0].textContent, cells[1].textContent, 
                cells[2].textContent, cells[3].textContent, cells[4].textContent
              ]);
            });
          });
          
          const csv = csvData.map(row => row.map(cell => ` + "`" + `"${cell}"` + "`" + `).join(',')).join('\\n');
          downloadFile(csv, 'nmap-scan-results.csv', 'text/csv');
        };

        window.exportJSON = function() {
          const jsonData = {
            scan_info: {
              scanner: 'nmap',
              generated: new Date().toISOString(),
              total_hosts: hosts.length
            },
            hosts: []
          };
          
          hosts.forEach(host => {
            const ip = host.querySelector('.ip').textContent;
            const hostname = host.querySelector('.hostname')?.textContent || '';
            const status = host.getAttribute('data-status');
            const ports = [];
            
            const rows = host.querySelectorAll('.ports-table tbody tr');
            rows.forEach(row => {
              const cells = row.querySelectorAll('td');
              ports.push({
                port: cells[0].textContent,
                protocol: cells[1].textContent,
                state: cells[2].textContent,
                service: cells[3].textContent,
                product: cells[4].textContent
              });
            });
            
            jsonData.hosts.push({ ip, hostname, status, ports });
          });
          
          downloadFile(JSON.stringify(jsonData, null, 2), 'nmap-scan-results.json', 'application/json');
        };

        window.exportPDF = function() {
          window.print();
        };

        window.exportSummary = function() {
          const totalHosts = hosts.length;
          const upHosts = hosts.filter(h => h.getAttribute('data-status') === 'up').length;
          const openPorts = document.getElementById('openPortCount').textContent;
          const services = document.getElementById('serviceCount').textContent;
          const riskLevel = document.getElementById('riskLevel').textContent;
          
          const summary = ` + "`" + `Network Security Scan Summary
==========================================

📊 Total Hosts Scanned: ${totalHosts}
🟢 Hosts Online: ${upHosts}
🔴 Hosts Offline: ${totalHosts - upHosts}
🔓 Open Ports Found: ${openPorts}
🛡️ Unique Services: ${services}
⚠️ Overall Risk Level: ${riskLevel}

Generated: ${new Date().toLocaleString()}
Tool: Nmap HTML Converter v1.0.0
Created by: Richard Jones @ DefenceLogic.io` + "`" + `;
          
          navigator.clipboard.writeText(summary).then(() => {
            alert('Summary copied to clipboard!');
          });
        };

        function downloadFile(content, filename, contentType) {
          const blob = new Blob([content], { type: contentType });
          const url = URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = filename;
          a.click();
          URL.revokeObjectURL(url);
        }

        toggleStatsBtn.addEventListener('click', () => {
          const isHidden = statsOverlay.classList.contains('hidden');
          statsOverlay.classList.toggle('hidden');
          toggleStatsBtn.textContent = isHidden ? '📊 Hide Stats' : '📊 Show Stats';
        });

        // Print button - expand all sections before printing
        const printBtn = document.getElementById('printReport');
        if(printBtn) {
          printBtn.addEventListener('click', () => {
            // Expand all host details
            hosts.forEach(host => {
              const body = host.querySelector('.host-body');
              const button = host.querySelector('.toggle');
              const text = button?.querySelector('.toggle-text');
              if(body) {
                body.removeAttribute('hidden');
                if(text) text.textContent = 'Collapse';
                if(button) button.setAttribute('aria-expanded', 'true');
              }
            });
            // Small delay to let the DOM update before printing
            setTimeout(() => window.print(), 100);
          });
        }

        document.addEventListener('keydown', function(e){
          if(e.key === '/' && !e.target.matches('input')) {
            e.preventDefault();
            search.focus();
          }
          if(e.key === 'Escape' && document.activeElement === search) {
            search.blur();
          }
        });

        updateStats();
        
        window.addEventListener('load', () => {
          document.body.classList.remove('loading');
        });

        search.setAttribute('aria-label', 'Search hosts and services');
        setTimeout(() => search.focus(), 100);
      })();
    </script>
  </body>
</html>
{{end}}`

func main() {
	var xmlPath, outPath, tplPath, cssPath string
	var showVersion bool

	flag.StringVar(&xmlPath, "xml", "", "input nmap XML file (default: stdin)")
	flag.StringVar(&outPath, "out", "nmap.html", "output HTML file")
	flag.StringVar(&tplPath, "tpl", "", "custom HTML template file (optional, uses embedded template by default)")
	flag.StringVar(&cssPath, "css", "", "custom CSS file (optional, uses embedded CSS by default)")
	flag.BoolVar(&showVersion, "version", false, "show version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Nmap HTML Converter v1.0.0\n")
		fmt.Fprintf(os.Stderr, "Created by Richard Jones - DefenceLogic.io\n\n")
		fmt.Fprintf(os.Stderr, "Converts Nmap XML scan results into beautiful, interactive HTML reports.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -xml scan-results.xml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -xml scan.xml -out report.html\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  cat scan.xml | %s\n", os.Args[0])
	}

	flag.Parse()

	// Show help if no arguments provided
	if flag.NFlag() == 0 && xmlPath == "" {
		flag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Println("Nmap HTML Converter v1.0.0")
		fmt.Println("Created by Richard Jones - DefenceLogic.io")
		fmt.Println("A modern tool for converting Nmap XML to interactive HTML reports")
		os.Exit(0)
	}

	// input reader
	var in io.Reader
	if xmlPath == "" {
		in = os.Stdin
	} else {
		f, err := os.Open(xmlPath)
		if err != nil {
			log.Fatalf("open xml: %v", err)
		}
		defer f.Close()
		in = f
	}

	// output file
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("create output: %v", err)
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	// load template - use embedded by default or custom if provided
	var tpl *template.Template
	if tplPath != "" {
		tpl, err = template.ParseFiles(tplPath)
		if err != nil {
			log.Fatalf("parse custom template: %v", err)
		}
	} else {
		tpl, err = template.New("embedded").Parse(defaultTemplate)
		if err != nil {
			log.Fatalf("parse embedded template: %v", err)
		}
	}

	// read css - use embedded by default or custom if provided
	var cssContent string
	if cssPath != "" {
		b, err := os.ReadFile(cssPath)
		if err != nil {
			log.Fatalf("read custom css: %v", err)
		}
		cssContent = string(b)
	} else {
		cssContent = defaultCSS
	}

	decoder := xml.NewDecoder(in)

	// read root <nmaprun> attributes for header
	var info NmapRunInfo
	// move decoder until we hit nmaprun start element and decode into info
	for {
		tok, err := decoder.Token()
		if err != nil {
			log.Fatalf("reading xml: %v", err)
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "nmaprun" {
			if err := decoder.DecodeElement(&info, &se); err != nil {
				// NOTE: we decoded the whole nmaprun which includes hosts; that's not ideal for streaming
				// Instead we'll just extract attributes if available by reading se.Attr. Simpler: rebuild info from se.Attr.
				info = NmapRunInfo{}
				for _, a := range se.Attr {
					switch a.Name.Local {
					case "scanner":
						info.Scanner = a.Value
					case "startstr":
						info.StartStr = a.Value
					case "args":
						info.Args = a.Value
					case "start":
						info.StartTime = a.Value
					}
				}
			}
			break
		}
	}

	// execute header template
	data := TemplateData{
		Info:      info,
		CSS:       template.CSS(cssContent),
		Generated: time.Now(),
	}
	if err := tpl.ExecuteTemplate(writer, "header", data); err != nil {
		log.Fatalf("execute header: %v", err)
	}

	// Rewind: create a new decoder (can't rewind stdin; if stdin used and we already consumed, user should pass file)
	// To keep streaming safe, reopen the file if xmlPath provided. If stdin used, we assume we have full file piped.
	if xmlPath != "" {
		f2, err := os.Open(xmlPath)
		if err != nil {
			log.Fatalf("reopen xml: %v", err)
		}
		defer f2.Close()
		decoder = xml.NewDecoder(f2)
	} else {
		// stdin: we already consumed root; to keep code simple, re-open os.Stdin is not possible.
		// For stdin use, recommend piping from file or use -xml.
		log.Println("stdin streaming may not support large files correctly; prefer -xml file for streaming.")
	}

	// stream hosts and render host template per host
	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("xml token: %v", err)
		}
		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "host" {
				var h Host
				if err := decoder.DecodeElement(&h, &se); err != nil {
					log.Fatalf("decode host: %v", err)
				}
				// execute host template with h as context
				if err := tpl.ExecuteTemplate(writer, "host", h); err != nil {
					log.Fatalf("execute host template: %v", err)
				}
			}
		}
	}

	// footer
	if err := tpl.ExecuteTemplate(writer, "footer", data); err != nil {
		// footer optional: ignore if not defined
		if !strings.Contains(err.Error(), "no template") {
			log.Fatalf("execute footer: %v", err)
		}
	}
}
