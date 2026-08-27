#!/usr/bin/env bash
# ==============================================================================
# StraitKubeGateway Helm Chart Repository Hosting & Publishing Script
# Generates static assets for GitHub Pages (https://msaeedb40.github.io/straitKubegateway)
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
CHART_DIR="${ROOT_DIR}/straitKubegateway-helm-repo"
OUTPUT_DIR="${1:-${ROOT_DIR}/dist/charts}"
REPO_URL="${HELM_REPO_URL:-https://msaeedb40.github.io/straitKubegateway}"
CUSTOM_DOMAIN="${CUSTOM_DOMAIN:-}"

echo "==> Preparing StraitKubeGateway Helm Chart Repository"
echo "    Chart directory: ${CHART_DIR}"
echo "    Output directory: ${OUTPUT_DIR}"
echo "    Repository URL:   ${REPO_URL}"
echo "    Custom Domain:    ${CUSTOM_DOMAIN:-none (GitHub Pages default)}"

mkdir -p "${OUTPUT_DIR}"

# 1. Lint Chart
echo "==> Linting Helm Chart..."
helm lint "${CHART_DIR}"

# 2. Package Chart
echo "==> Packaging Helm Chart to ${OUTPUT_DIR}..."
helm package "${CHART_DIR}" --destination "${OUTPUT_DIR}"

# 3. Generate / Update index.yaml
echo "==> Indexing Helm Chart repository..."
if [ -f "${OUTPUT_DIR}/index.yaml" ]; then
    helm repo index "${OUTPUT_DIR}" --url "${REPO_URL}" --merge "${OUTPUT_DIR}/index.yaml"
else
    helm repo index "${OUTPUT_DIR}" --url "${REPO_URL}"
fi

# 4. Create CNAME file only if CUSTOM_DOMAIN is provided
if [ -n "${CUSTOM_DOMAIN}" ]; then
    echo "==> Writing CNAME file (${CUSTOM_DOMAIN})..."
    echo "${CUSTOM_DOMAIN}" > "${OUTPUT_DIR}/CNAME"
else
    echo "==> No custom domain set; using GitHub Pages URL: ${REPO_URL}"
    rm -f "${OUTPUT_DIR}/CNAME"
fi

# 5. Create .nojekyll to prevent GitHub Pages Jekyll processing
touch "${OUTPUT_DIR}/.nojekyll"

CHART_VERSION=$(grep '^version:' "${CHART_DIR}/Chart.yaml" | awk '{print $2}')
APP_VERSION=$(grep '^appVersion:' "${CHART_DIR}/Chart.yaml" | awk '{print $2}' | tr -d '"')
CHART_NAME=$(grep '^name:' "${CHART_DIR}/Chart.yaml" | awk '{print $2}')

# 7. Generate beautiful modern index.html landing page with Tailwind CSS v4
cat <<EOF > "${OUTPUT_DIR}/index.html"
<!DOCTYPE html>
<html lang="en" class="scroll-smooth dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>StraitKubeGateway | Official Helm Charts Repository</title>
    <meta name="description" content="Official Helm chart repository for StraitKubeGateway: Kubernetes-native eBPF Transit Gateway, CNI, and Multi-Cluster Service Mesh.">
    <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%2338bdf8'><path d='M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5'/></svg>">
    <!-- Google Fonts -->
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700;800;900&family=JetBrains+Mono:ital,wght@0,400;0,500;0,600;0,700;1,400&display=swap" rel="stylesheet">
    <!-- Tailwind CSS v4 Browser CDN -->
    <script src="https://unpkg.com/@tailwindcss/browser@4"></script>
    <style>
        :root {
            --font-sans: 'Plus Jakarta Sans', system-ui, -apple-system, sans-serif;
            --font-mono: 'JetBrains Mono', monospace;
            --color-brand-cyan: #38bdf8;
            --color-brand-blue: #3b82f6;
            --color-brand-indigo: #6366f1;
            --color-brand-emerald: #10b981;
        }
        body {
            font-family: 'Plus Jakarta Sans', system-ui, -apple-system, sans-serif;
        }
        code, pre, .font-mono {
            font-family: 'JetBrains Mono', monospace;
        }
        .bg-grid-pattern {
            background-size: 32px 32px;
            background-image: 
                linear-gradient(to right, rgba(255, 255, 255, 0.04) 1px, transparent 1px),
                linear-gradient(to bottom, rgba(255, 255, 255, 0.04) 1px, transparent 1px);
        }
        .glow-radial-cyan {
            background: radial-gradient(circle at 50% 0%, rgba(56, 189, 248, 0.18) 0%, rgba(59, 130, 246, 0.08) 45%, transparent 75%);
        }
        .glass-panel {
            background: rgba(15, 23, 42, 0.75);
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            border: 1px solid rgba(255, 255, 255, 0.08);
        }
        .glass-card {
            background: rgba(30, 41, 59, 0.45);
            backdrop-filter: blur(12px);
            -webkit-backdrop-filter: blur(12px);
            border: 1px solid rgba(255, 255, 255, 0.06);
        }
        .glass-card:hover {
            border-color: rgba(56, 189, 248, 0.35);
            background: rgba(30, 41, 59, 0.7);
        }
        /* Custom scrollbar */
        ::-webkit-scrollbar {
            width: 6px;
            height: 6px;
        }
        ::-webkit-scrollbar-track {
            background: #090d16;
        }
        ::-webkit-scrollbar-thumb {
            background: #1e293b;
            border-radius: 9999px;
        }
        ::-webkit-scrollbar-thumb:hover {
            background: #38bdf8;
        }
    </style>
</head>
<body class="bg-[#070a12] text-slate-100 min-h-screen relative overflow-x-hidden selection:bg-cyan-500/30 selection:text-cyan-200 antialiased">
    
    <!-- Background Ambient Glow & Grid Pattern -->
    <div class="fixed inset-0 bg-grid-pattern pointer-events-none -z-20 opacity-80"></div>
    <div class="fixed top-0 left-0 right-0 h-[650px] glow-radial-cyan pointer-events-none -z-10"></div>
    <div class="fixed bottom-0 right-0 w-[550px] h-[450px] bg-gradient-to-tl from-indigo-500/10 via-emerald-500/5 to-transparent blur-[140px] pointer-events-none -z-10"></div>

    <!-- Navigation Bar -->
    <nav class="sticky top-0 z-50 glass-panel border-b border-slate-800/80 px-4 sm:px-8 py-3.5 transition-all">
        <div class="max-w-6xl mx-auto flex items-center justify-between">
            <!-- Brand Logo -->
            <a href="#" class="flex items-center gap-3 group">
                <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-cyan-500 via-sky-500 to-indigo-600 p-[1.5px] shadow-lg shadow-cyan-500/20 group-hover:scale-105 transition-transform">
                    <div class="w-full h-full bg-slate-950 rounded-[10px] flex items-center justify-center">
                        <svg class="w-5 h-5 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
                            <polyline points="2 17 12 22 22 17"></polyline>
                            <polyline points="2 12 12 17 22 12"></polyline>
                        </svg>
                    </div>
                </div>
                <div>
                    <span class="font-extrabold text-base tracking-tight text-white flex items-center gap-1.5">
                        StraitKubeGateway
                        <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-400 border border-cyan-500/30">HELM</span>
                    </span>
                    <span class="block text-[11px] text-slate-400 font-medium -mt-0.5">eBPF Transit Gateway & CNI</span>
                </div>
            </a>

            <!-- Quick Links -->
            <div class="flex items-center gap-3 sm:gap-4 text-xs font-medium">
                <a href="https://github.com/msaeedb40/straitKubegateway" target="_blank" rel="noopener noreferrer" class="hidden sm:flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800/60 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-700/60 transition-all">
                    <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path fill-rule="evenodd" clip-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/></svg>
                    <span>GitHub</span>
                </a>
                <a href="index.yaml" target="_blank" class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-400 border border-cyan-500/30 transition-all font-mono">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"></path></svg>
                    <span>index.yaml</span>
                </a>
            </div>
        </div>
    </nav>

    <!-- Main Container -->
    <div class="max-w-6xl mx-auto px-4 sm:px-6 py-10 sm:py-14 space-y-12">
        
        <!-- Hero Header -->
        <header class="text-center max-w-3xl mx-auto space-y-4">
            <!-- Pill Status -->
            <div class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-slate-900/90 border border-cyan-500/30 shadow-lg shadow-cyan-950/40 text-xs font-semibold">
                <span class="relative flex h-2.5 w-2.5">
                    <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-80"></span>
                    <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500 shadow-sm shadow-emerald-400"></span>
                </span>
                <span class="text-slate-300">Helm Repository Active</span>
                <span class="text-slate-600">•</span>
                <span class="font-mono text-cyan-400 font-bold">v${CHART_VERSION}</span>
                <span class="text-slate-600">•</span>
                <span class="text-slate-400 text-[11px]">Kubernetes 1.28+</span>
            </div>

            <!-- Main Heading -->
            <h1 class="text-3xl sm:text-5xl md:text-6xl font-black tracking-tight leading-[1.1] bg-gradient-to-b from-white via-slate-100 to-cyan-300 bg-clip-text text-transparent">
                Kernel-Accelerated Kubernetes Networking
            </h1>

            <!-- Subtitle -->
            <p class="text-slate-400 text-sm sm:text-base md:text-lg leading-relaxed pt-1 max-w-2xl mx-auto">
                Deploy StraitKubeGateway with Helm: unified eBPF dataplane, complete kube-proxy replacement, high-performance CNI, and cross-cluster transit routing.
            </p>

            <!-- Metrics Pills -->
            <div class="flex flex-wrap items-center justify-center gap-2 pt-2 text-xs font-medium text-slate-300">
                <span class="px-3 py-1 rounded-full bg-slate-900/80 border border-slate-800 flex items-center gap-1.5">
                    <span class="text-cyan-400">⚡</span> Linux Kernel 6.7+ CO-RE
                </span>
                <span class="px-3 py-1 rounded-full bg-slate-900/80 border border-slate-800 flex items-center gap-1.5">
                    <span class="text-emerald-400">🛡️</span> Zero Kube-Proxy Overhead
                </span>
                <span class="px-3 py-1 rounded-full bg-slate-900/80 border border-slate-800 flex items-center gap-1.5">
                    <span class="text-blue-400">🌐</span> Gateway API v1.6.1 Ready
                </span>
            </div>
        </header>

        <!-- Interactive Installation Center -->
        <section class="glass-panel rounded-3xl p-6 sm:p-8 md:p-10 shadow-2xl shadow-black/80 relative overflow-hidden border border-slate-800">
            <!-- Glow Highlights -->
            <div class="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-transparent via-cyan-500 to-transparent"></div>

            <div class="flex flex-col lg:flex-row gap-8 items-start">
                
                <!-- Left: Interactive Options & Toggles -->
                <div class="w-full lg:w-5/12 space-y-6">
                    <div>
                        <h2 class="text-lg font-bold text-white flex items-center gap-2">
                            <span class="w-2 h-2 rounded-full bg-cyan-400"></span>
                            Installation Configurator
                        </h2>
                        <p class="text-xs text-slate-400 mt-1">Select feature presets or customize flags to generate your deployment command.</p>
                    </div>

                    <!-- Mode Selector Tabs -->
                    <div class="grid grid-cols-2 gap-2 p-1.5 bg-slate-950/80 rounded-xl border border-slate-800">
                        <button id="tab-quick" onclick="switchTab('quick')" class="tab-btn px-3 py-2 rounded-lg text-xs font-semibold transition-all bg-cyan-500/20 text-cyan-300 border border-cyan-500/40">
                            🚀 Default Quickstart
                        </button>
                        <button id="tab-custom" onclick="switchTab('custom')" class="tab-btn px-3 py-2 rounded-lg text-xs font-semibold text-slate-400 hover:text-white transition-all">
                            ⚙️ Feature Toggles
                        </button>
                    </div>

                    <!-- Interactive Toggles (Shown in custom mode) -->
                    <div id="toggles-section" class="space-y-3 hidden bg-slate-950/60 p-4 rounded-2xl border border-slate-800/80">
                        <span class="text-[11px] font-bold tracking-wider uppercase text-slate-400 block mb-2">Enable Capabilities</span>
                        
                        <label class="flex items-center justify-between p-2.5 rounded-xl hover:bg-slate-900/60 cursor-pointer transition-colors border border-transparent hover:border-slate-800">
                            <div class="flex items-center gap-2.5">
                                <span class="text-cyan-400 text-sm">🔒</span>
                                <div>
                                    <div class="text-xs font-semibold text-white">WireGuard Mesh Encryption</div>
                                    <div class="text-[10px] text-slate-400">Hardware-accelerated node crypto</div>
                                </div>
                            </div>
                            <input type="checkbox" id="opt-wireguard" onchange="updateCommand()" class="w-4 h-4 rounded text-cyan-500 bg-slate-900 border-slate-700 focus:ring-cyan-500">
                        </label>

                        <label class="flex items-center justify-between p-2.5 rounded-xl hover:bg-slate-900/60 cursor-pointer transition-colors border border-transparent hover:border-slate-800">
                            <div class="flex items-center gap-2.5">
                                <span class="text-blue-400 text-sm">🌐</span>
                                <div>
                                    <div class="text-xs font-semibold text-white">Gateway API Support</div>
                                    <div class="text-[10px] text-slate-400">HTTPRoute, TCPRoute, UDPRoute</div>
                                </div>
                            </div>
                            <input type="checkbox" id="opt-gwapi" checked onchange="updateCommand()" class="w-4 h-4 rounded text-cyan-500 bg-slate-900 border-slate-700 focus:ring-cyan-500">
                        </label>

                        <label class="flex items-center justify-between p-2.5 rounded-xl hover:bg-slate-900/60 cursor-pointer transition-colors border border-transparent hover:border-slate-800">
                            <div class="flex items-center gap-2.5">
                                <span class="text-emerald-400 text-sm">🔀</span>
                                <div>
                                    <div class="text-xs font-semibold text-white">Multi-Cluster Transit Gateway</div>
                                    <div class="text-[10px] text-slate-400">Segment routing & peering</div>
                                </div>
                            </div>
                            <input type="checkbox" id="opt-transit" onchange="updateCommand()" class="w-4 h-4 rounded text-cyan-500 bg-slate-900 border-slate-700 focus:ring-cyan-500">
                        </label>

                        <label class="flex items-center justify-between p-2.5 rounded-xl hover:bg-slate-900/60 cursor-pointer transition-colors border border-transparent hover:border-slate-800">
                            <div class="flex items-center gap-2.5">
                                <span class="text-indigo-400 text-sm">📊</span>
                                <div>
                                    <div class="text-xs font-semibold text-white">Prometheus & Tracing</div>
                                    <div class="text-[10px] text-slate-400">eBPF flow telemetry export</div>
                                </div>
                            </div>
                            <input type="checkbox" id="opt-metrics" checked onchange="updateCommand()" class="w-4 h-4 rounded text-cyan-500 bg-slate-900 border-slate-700 focus:ring-cyan-500">
                        </label>
                    </div>

                    <!-- Metadata Overview -->
                    <div class="grid grid-cols-2 gap-3 pt-2">
                        <div class="glass-card rounded-xl p-3.5">
                            <div class="text-[10px] font-bold uppercase tracking-wider text-slate-400 mb-1">Chart Package</div>
                            <div class="text-xs font-mono font-bold text-white truncate">${CHART_NAME}</div>
                        </div>
                        <div class="glass-card rounded-xl p-3.5">
                            <div class="text-[10px] font-bold uppercase tracking-wider text-slate-400 mb-1">Latest Version</div>
                            <div class="text-xs font-mono font-bold text-cyan-400">v${CHART_VERSION}</div>
                        </div>
                        <div class="glass-card rounded-xl p-3.5">
                            <div class="text-[10px] font-bold uppercase tracking-wider text-slate-400 mb-1">App Version</div>
                            <div class="text-xs font-mono font-bold text-emerald-400">v${APP_VERSION}</div>
                        </div>
                        <div class="glass-card rounded-xl p-3.5">
                            <div class="text-[10px] font-bold uppercase tracking-wider text-slate-400 mb-1">Direct Archive</div>
                            <a href="${CHART_NAME}-${CHART_VERSION}.tgz" class="text-xs font-mono text-cyan-400 hover:text-cyan-300 underline inline-flex items-center gap-1">
                                <span>Download .tgz</span>
                                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path></svg>
                            </a>
                        </div>
                    </div>
                </div>

                <!-- Right: Code Terminal & Execution -->
                <div class="w-full lg:w-7/12 space-y-5">
                    
                    <!-- Step 1: Add Repo -->
                    <div>
                        <div class="flex items-center justify-between mb-2">
                            <div class="flex items-center gap-2">
                                <span class="flex items-center justify-center w-6 h-6 rounded-full bg-cyan-500/20 text-cyan-300 font-mono text-xs font-bold border border-cyan-500/30">1</span>
                                <span class="text-xs font-bold uppercase tracking-wider text-slate-300">Add & Update Repository</span>
                            </div>
                        </div>
                        <div class="relative group bg-[#090d16] border border-slate-800 rounded-2xl p-4 overflow-x-auto shadow-inner">
                            <button class="copy-btn absolute top-3 right-3 bg-slate-800/90 hover:bg-cyan-500/20 hover:text-cyan-300 text-slate-400 border border-slate-700/80 hover:border-cyan-500/50 rounded-lg px-2.5 py-1 text-xs font-medium transition-all duration-200 flex items-center gap-1.5 cursor-pointer z-10" onclick="copyCode(this, 'helm repo add straitkubegateway ${REPO_URL}\nhelm repo update')">
                                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path></svg>
                                <span>Copy</span>
                            </button>
                            <pre class="font-mono text-xs sm:text-sm text-cyan-400 leading-relaxed pr-16 whitespace-pre"><code>helm repo add straitkubegateway ${REPO_URL}
helm repo update</code></pre>
                        </div>
                    </div>

                    <!-- Step 2: Install Command (Dynamic) -->
                    <div>
                        <div class="flex items-center justify-between mb-2">
                            <div class="flex items-center gap-2">
                                <span class="flex items-center justify-center w-6 h-6 rounded-full bg-cyan-500/20 text-cyan-300 font-mono text-xs font-bold border border-cyan-500/30">2</span>
                                <span class="text-xs font-bold uppercase tracking-wider text-slate-300">Install StraitKubeGateway</span>
                            </div>
                            <span class="text-[11px] font-mono text-slate-400">namespace: kube-system</span>
                        </div>
                        <div class="relative group bg-[#090d16] border border-slate-800 rounded-2xl p-4 overflow-x-auto shadow-inner">
                            <button id="btn-copy-install" class="copy-btn absolute top-3 right-3 bg-slate-800/90 hover:bg-cyan-500/20 hover:text-cyan-300 text-slate-400 border border-slate-700/80 hover:border-cyan-500/50 rounded-lg px-2.5 py-1 text-xs font-medium transition-all duration-200 flex items-center gap-1.5 cursor-pointer z-10" onclick="copyDynamicInstall(this)">
                                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path></svg>
                                <span>Copy</span>
                            </button>
                            <pre id="code-install" class="font-mono text-xs sm:text-sm text-cyan-300 leading-relaxed pr-16 whitespace-pre"><code>helm install straitkubegateway straitkubegateway/straitkubegateway \
  --namespace kube-system \
  --create-namespace</code></pre>
                        </div>
                    </div>

                    <!-- Verification step -->
                    <div class="pt-1">
                        <div class="text-[11px] font-bold uppercase tracking-wider text-slate-400 mb-1.5 flex items-center gap-1.5">
                            <span>🔍 Verify Deployment</span>
                        </div>
                        <div class="relative group bg-[#090d16]/80 border border-slate-800/80 rounded-xl p-3 text-xs font-mono text-slate-300 overflow-x-auto">
                            <code>kubectl get pods -n kube-system -l app.kubernetes.io/name=straitkubegateway</code>
                        </div>
                    </div>

                </div>
            </div>
        </section>

        <!-- Architecture & Dataplane Highlights Grid -->
        <section class="space-y-6">
            <div class="text-center max-w-2xl mx-auto">
                <h2 class="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Core Architecture & Dataplane</h2>
                <p class="text-sm text-slate-400 mt-1">Built specifically for high-throughput, zero-copy Kubernetes infrastructure.</p>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-5">
                
                <!-- Card 1 -->
                <div class="glass-card rounded-2xl p-6 transition-all duration-300 hover:-translate-y-1">
                    <div class="w-10 h-10 rounded-xl bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center text-cyan-400 text-lg mb-4">
                        ⚡
                    </div>
                    <h3 class="text-base font-bold text-white mb-2">Native eBPF Dataplane</h3>
                    <p class="text-xs text-slate-400 leading-relaxed">
                        TCX, NetKit, and XDP hooks in Linux kernel 6.7+ bypassing iptables and conntrack bottlenecks entirely.
                    </p>
                </div>

                <!-- Card 2 -->
                <div class="glass-card rounded-2xl p-6 transition-all duration-300 hover:-translate-y-1">
                    <div class="w-10 h-10 rounded-xl bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400 text-lg mb-4">
                        🚀
                    </div>
                    <h3 class="text-base font-bold text-white mb-2">Kube-Proxy Replacement</h3>
                    <p class="text-xs text-slate-400 leading-relaxed">
                        Default in-kernel Maglev hash and consistent load balancing for NodePort, ClusterIP, and LoadBalancer services.
                    </p>
                </div>

                <!-- Card 3 -->
                <div class="glass-card rounded-2xl p-6 transition-all duration-300 hover:-translate-y-1">
                    <div class="w-10 h-10 rounded-xl bg-blue-500/10 border border-blue-500/30 flex items-center justify-center text-blue-400 text-lg mb-4">
                        🌐
                    </div>
                    <h3 class="text-base font-bold text-white mb-2">Multi-Cluster Transit</h3>
                    <p class="text-xs text-slate-400 leading-relaxed">
                        Hub-and-spoke and mesh transit gateway with 32-bit segment isolation and cross-cluster routing policies.
                    </p>
                </div>

                <!-- Card 4 -->
                <div class="glass-card rounded-2xl p-6 transition-all duration-300 hover:-translate-y-1">
                    <div class="w-10 h-10 rounded-xl bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400 text-lg mb-4">
                        🛡️
                    </div>
                    <h3 class="text-base font-bold text-white mb-2">Built-in CNI & NetKit</h3>
                    <p class="text-xs text-slate-400 leading-relaxed">
                        CNI Spec 1.1+ compliant pod allocation with NetKit veth replacement, dynamic IPAM, and systemd cgroup v2 control.
                    </p>
                </div>

                <!-- Card 5 -->
                <div class="glass-card rounded-2xl p-6 transition-all duration-300 hover:-translate-y-1">
                    <div class="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/30 flex items-center justify-center text-purple-400 text-lg mb-4">
                        🔒
                    </div>
                    <h3 class="text-base font-bold text-white mb-2">Zero-Trust Security</h3>
                    <p class="text-xs text-slate-400 leading-relaxed">
                        eBPF NetworkPolicy enforcement with pod/namespace identity, stateful firewalling, and automated WireGuard/IPsec.
                    </p>
                </div>

                <!-- Card 6 -->
                <div class="glass-card rounded-2xl p-6 transition-all duration-300 hover:-translate-y-1">
                    <div class="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400 text-lg mb-4">
                        🧭
                    </div>
                    <h3 class="text-base font-bold text-white mb-2">Gateway API v1.6.1</h3>
                    <p class="text-xs text-slate-400 leading-relaxed">
                        Native controller support for GatewayClass, Gateway, HTTPRoute, TCPRoute, UDPRoute, GRPCRoute, and TLSRoute.
                    </p>
                </div>

            </div>
        </section>

        <!-- Footer -->
        <footer class="pt-8 border-t border-slate-800/80 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-slate-400">
            <div class="flex items-center gap-2">
                <span class="font-semibold text-slate-300">StraitKubeGateway</span>
                <span>•</span>
                <span>Open Source Project (Apache 2.0)</span>
            </div>

            <div class="flex items-center gap-6">
                <a href="https://github.com/msaeedb40/straitKubegateway" target="_blank" rel="noopener noreferrer" class="hover:text-cyan-400 transition-colors">GitHub</a>
                <a href="https://github.com/msaeedb40/straitKubegateway/tree/main/docs" target="_blank" rel="noopener noreferrer" class="hover:text-cyan-400 transition-colors">Documentation</a>
                <a href="index.yaml" target="_blank" class="hover:text-cyan-400 transition-colors font-mono">index.yaml</a>
            </div>
        </footer>

    </div>

    <!-- Interactive Script for Tabs, Toggles, and Copy Actions -->
    <script>
        let currentTab = 'quick';

        function switchTab(tab) {
            currentTab = tab;
            const tabQuick = document.getElementById('tab-quick');
            const tabCustom = document.getElementById('tab-custom');
            const togglesSection = document.getElementById('toggles-section');

            if (tab === 'quick') {
                tabQuick.className = 'tab-btn px-3 py-2 rounded-lg text-xs font-semibold transition-all bg-cyan-500/20 text-cyan-300 border border-cyan-500/40';
                tabCustom.className = 'tab-btn px-3 py-2 rounded-lg text-xs font-semibold text-slate-400 hover:text-white transition-all';
                togglesSection.classList.add('hidden');
            } else {
                tabCustom.className = 'tab-btn px-3 py-2 rounded-lg text-xs font-semibold transition-all bg-cyan-500/20 text-cyan-300 border border-cyan-500/40';
                tabQuick.className = 'tab-btn px-3 py-2 rounded-lg text-xs font-semibold text-slate-400 hover:text-white transition-all';
                togglesSection.classList.remove('hidden');
            }
            updateCommand();
        }

        function updateCommand() {
            const codeElem = document.getElementById('code-install');
            if (currentTab === 'quick') {
                codeElem.innerHTML = '<code>helm install straitkubegateway straitkubegateway/straitkubegateway \\\n  --namespace kube-system \\\n  --create-namespace</code>';
                return;
            }

            const wireguard = document.getElementById('opt-wireguard').checked;
            const gwapi = document.getElementById('opt-gwapi').checked;
            const transit = document.getElementById('opt-transit').checked;
            const metrics = document.getElementById('opt-metrics').checked;

            let lines = [
                'helm install straitkubegateway straitkubegateway/straitkubegateway \\\\',
                '  --namespace kube-system \\\\',
                '  --create-namespace \\\\'
            ];

            if (wireguard) {
                lines.push('  --set straitd.wireguard.enabled=true \\\\');
            }
            if (!gwapi) {
                lines.push('  --set gatewayAPI.enabled=false \\\\');
            }
            if (transit) {
                lines.push('  --set transit.enabled=true \\\\');
            }
            if (!metrics) {
                lines.push('  --set observability.prometheus.enabled=false \\\\');
            }

            // Remove trailing backslash from last line
            const lastIdx = lines.length - 1;
            lines[lastIdx] = lines[lastIdx].replace(/ \\\\$/, '');

            codeElem.innerHTML = '<code>' + lines.join('\n') + '</code>';
        }

        function copyDynamicInstall(btn) {
            const codeElem = document.getElementById('code-install');
            const rawText = codeElem.innerText || codeElem.textContent;
            copyCode(btn, rawText);
        }

        function copyCode(btn, text) {
            navigator.clipboard.writeText(text).then(() => {
                const originalHtml = btn.innerHTML;
                btn.innerHTML = '<svg class="w-3.5 h-3.5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg><span class="text-emerald-400">Copied!</span>';
                btn.classList.add('border-emerald-500/50', 'bg-emerald-500/10');
                setTimeout(() => {
                    btn.innerHTML = originalHtml;
                    btn.classList.remove('border-emerald-500/50', 'bg-emerald-500/10');
                }, 2000);
            });
        }
    </script>
</body>
</html>
EOF

chmod +x "${SCRIPT_DIR}/publish-helm-repo.sh" 2>/dev/null || true

echo "==> Chart repository generated successfully at: ${OUTPUT_DIR}"
