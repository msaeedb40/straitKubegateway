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
    <title>StraitKubeGateway Helm Charts Repository</title>
    <meta name="description" content="Official Helm chart repository for StraitKubeGateway: Kubernetes-native eBPF Transit Gateway, CNI, and Multi-Cluster Service Networking.">
    <!-- Google Fonts -->
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800;900&family=JetBrains+Mono:ital,wght@0,400;0,500;0,600;0,700;1,400&display=swap" rel="stylesheet">
    <!-- Tailwind CSS v4 Browser CDN -->
    <script src="https://unpkg.com/@tailwindcss/browser@4"></script>
    <style type="text/tailwindcss">
        @theme {
            --font-sans: 'Inter', system-ui, -apple-system, sans-serif;
            --font-mono: 'JetBrains Mono', monospace;
            --color-brand-cyan: #38bdf8;
            --color-brand-blue: #3b82f6;
            --color-brand-indigo: #6366f1;
            --color-brand-emerald: #10b981;
        }
    </style>
    <style>
        body {
            font-family: 'Inter', system-ui, -apple-system, sans-serif;
        }
        code, pre {
            font-family: 'JetBrains Mono', monospace;
        }
        .bg-grid-pattern {
            background-size: 40px 40px;
            background-image: 
                linear-gradient(to right, rgba(255, 255, 255, 0.03) 1px, transparent 1px),
                linear-gradient(to bottom, rgba(255, 255, 255, 0.03) 1px, transparent 1px);
        }
    </style>
</head>
<body class="bg-slate-950 text-slate-100 min-h-screen flex flex-col items-center justify-center p-4 sm:p-6 md:p-10 relative overflow-x-hidden selection:bg-cyan-500/30 selection:text-cyan-200">
    
    <!-- Background Glow Effects & Grid Pattern -->
    <div class="fixed inset-0 bg-grid-pattern pointer-events-none -z-20 opacity-70"></div>
    <div class="fixed top-0 left-1/2 -translate-x-1/2 w-[700px] h-[400px] bg-gradient-to-br from-cyan-500/15 via-blue-600/10 to-transparent blur-[120px] rounded-full pointer-events-none -z-10"></div>
    <div class="fixed bottom-0 right-10 w-[500px] h-[350px] bg-gradient-to-tl from-indigo-500/10 via-emerald-500/10 to-transparent blur-[100px] rounded-full pointer-events-none -z-10"></div>

    <div class="max-w-4xl w-full my-auto">
        <!-- Header -->
        <header class="text-center mb-8 sm:mb-10">
            <!-- Badge -->
            <div class="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-cyan-500/10 border border-cyan-500/30 text-cyan-400 text-xs sm:text-sm font-semibold mb-5 shadow-sm shadow-cyan-950">
                <span class="relative flex h-2 w-2">
                    <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                    <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                </span>
                <span>Official Helm Repository Active</span>
                <span class="text-slate-600">•</span>
                <span class="font-mono text-cyan-300">v${CHART_VERSION}</span>
            </div>

            <!-- Title -->
            <h1 class="text-3xl sm:text-4xl md:text-5xl font-black tracking-tight mb-4 bg-gradient-to-r from-white via-slate-100 to-cyan-400 bg-clip-text text-transparent">
                StraitKubeGateway Charts
            </h1>

            <!-- Subtitle -->
            <p class="text-slate-400 text-sm sm:text-base md:text-lg max-w-2xl mx-auto leading-relaxed">
                Kubernetes-native eBPF Transit Gateway, Production CNI, and Multi-Cluster Service Mesh with zero kube-proxy overhead.
            </p>
        </header>

        <!-- Main Card Container -->
        <main class="bg-slate-900/80 backdrop-blur-xl border border-slate-800/90 rounded-2xl p-6 sm:p-8 md:p-10 shadow-2xl shadow-black/60 relative overflow-hidden">
            <!-- Card accent top highlight -->
            <div class="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-cyan-500 via-blue-500 to-indigo-500"></div>

            <!-- Step 1 -->
            <section class="mb-8">
                <div class="flex items-center gap-3 mb-3">
                    <span class="flex items-center justify-center w-7 h-7 rounded-full bg-cyan-500/15 border border-cyan-500/40 text-cyan-400 font-mono text-xs font-bold">1</span>
                    <h2 class="text-base sm:text-lg font-bold text-white tracking-wide">Add Helm Repository</h2>
                </div>
                <div class="relative group bg-slate-950/90 border border-slate-800 rounded-xl p-4 sm:p-4.5 overflow-x-auto shadow-inner">
                    <button class="copy-btn absolute top-3 right-3 bg-slate-800/90 hover:bg-cyan-500/20 hover:text-cyan-300 text-slate-400 border border-slate-700/80 hover:border-cyan-500/50 rounded-lg px-2.5 py-1 text-xs font-medium transition-all duration-200 flex items-center gap-1.5 cursor-pointer z-10" onclick="copyCode(this, 'helm repo add straitkubegateway ${REPO_URL}\nhelm repo update')">
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path></svg>
                        <span>Copy</span>
                    </button>
                    <pre class="font-mono text-xs sm:text-sm text-cyan-400 leading-relaxed pr-16 whitespace-pre"><code>helm repo add straitkubegateway ${REPO_URL}
helm repo update</code></pre>
                </div>
            </section>

            <!-- Step 2 -->
            <section class="mb-8">
                <div class="flex items-center gap-3 mb-3">
                    <span class="flex items-center justify-center w-7 h-7 rounded-full bg-cyan-500/15 border border-cyan-500/40 text-cyan-400 font-mono text-xs font-bold">2</span>
                    <h2 class="text-base sm:text-lg font-bold text-white tracking-wide">Install StraitKubeGateway Chart</h2>
                </div>
                <div class="relative group bg-slate-950/90 border border-slate-800 rounded-xl p-4 sm:p-4.5 overflow-x-auto shadow-inner">
                    <button class="copy-btn absolute top-3 right-3 bg-slate-800/90 hover:bg-cyan-500/20 hover:text-cyan-300 text-slate-400 border border-slate-700/80 hover:border-cyan-500/50 rounded-lg px-2.5 py-1 text-xs font-medium transition-all duration-200 flex items-center gap-1.5 cursor-pointer z-10" onclick="copyCode(this, 'helm install straitkubegateway straitkubegateway/straitkubegateway \\\n  --namespace kube-system \\\n  --create-namespace')">
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path></svg>
                        <span>Copy</span>
                    </button>
                    <pre class="font-mono text-xs sm:text-sm text-cyan-400 leading-relaxed pr-16 whitespace-pre"><code>helm install straitkubegateway straitkubegateway/straitkubegateway \\
  --namespace kube-system \\
  --create-namespace</code></pre>
                </div>
            </section>

            <!-- Metadata Grid -->
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4 pt-4 border-t border-slate-800/80">
                <div class="bg-slate-950/60 border border-slate-800/60 rounded-xl p-3.5 transition-colors hover:border-slate-700">
                    <div class="text-[11px] font-semibold uppercase tracking-wider text-slate-400 mb-1">Chart Name</div>
                    <div class="text-sm sm:text-base font-bold text-white font-mono truncate">${CHART_NAME}</div>
                </div>
                <div class="bg-slate-950/60 border border-slate-800/60 rounded-xl p-3.5 transition-colors hover:border-slate-700">
                    <div class="text-[11px] font-semibold uppercase tracking-wider text-slate-400 mb-1">Chart Version</div>
                    <div class="text-sm sm:text-base font-bold text-cyan-400 font-mono">v${CHART_VERSION}</div>
                </div>
                <div class="bg-slate-950/60 border border-slate-800/60 rounded-xl p-3.5 transition-colors hover:border-slate-700">
                    <div class="text-[11px] font-semibold uppercase tracking-wider text-slate-400 mb-1">App Version</div>
                    <div class="text-sm sm:text-base font-bold text-emerald-400 font-mono">v${APP_VERSION}</div>
                </div>
                <div class="bg-slate-950/60 border border-slate-800/60 rounded-xl p-3.5 transition-colors hover:border-slate-700">
                    <div class="text-[11px] font-semibold uppercase tracking-wider text-slate-400 mb-1">Package (.tgz)</div>
                    <div class="text-sm sm:text-base font-bold truncate">
                        <a href="${CHART_NAME}-${CHART_VERSION}.tgz" class="text-cyan-400 hover:text-cyan-300 hover:underline font-mono inline-flex items-center gap-1">
                            <span>Download</span>
                            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path></svg>
                        </a>
                    </div>
                </div>
            </div>

            <!-- Feature Tags -->
            <div class="mt-6 pt-4 border-t border-slate-800/50 flex flex-wrap items-center justify-center gap-2 text-xs">
                <span class="px-2.5 py-1 rounded-md bg-slate-800/80 border border-slate-700/60 text-slate-300">⚡ eBPF CO-RE Dataplane</span>
                <span class="px-2.5 py-1 rounded-md bg-slate-800/80 border border-slate-700/60 text-slate-300">🛡️ Built-in CNI & NetKit</span>
                <span class="px-2.5 py-1 rounded-md bg-slate-800/80 border border-slate-700/60 text-slate-300">🚀 Kube-Proxy Replacement</span>
                <span class="px-2.5 py-1 rounded-md bg-slate-800/80 border border-slate-700/60 text-slate-300">🌐 Gateway API v1.6.1</span>
                <span class="px-2.5 py-1 rounded-md bg-slate-800/80 border border-slate-700/60 text-slate-300">🔒 WireGuard & IPsec</span>
            </div>
        </main>

        <!-- Navigation Links -->
        <nav class="flex flex-wrap items-center justify-center gap-6 sm:gap-8 mt-8 text-sm font-medium">
            <a href="https://github.com/msaeedb40/straitKubegateway" target="_blank" rel="noopener noreferrer" class="text-cyan-400 hover:text-white transition-colors flex items-center gap-1.5 group">
                <span>GitHub Repository</span>
                <span class="transition-transform group-hover:translate-x-0.5">→</span>
            </a>
            <a href="index.yaml" target="_blank" class="text-cyan-400 hover:text-white transition-colors flex items-center gap-1.5 group">
                <span>Raw index.yaml</span>
                <span class="transition-transform group-hover:translate-x-0.5">→</span>
            </a>
            <a href="https://github.com/msaeedb40/straitKubegateway/tree/main/docs" target="_blank" rel="noopener noreferrer" class="text-slate-400 hover:text-white transition-colors flex items-center gap-1.5 group">
                <span>Documentation</span>
                <span class="transition-transform group-hover:translate-x-0.5">→</span>
            </a>
        </nav>

        <!-- Footer -->
        <footer class="mt-8 text-center text-xs text-slate-400">
            &copy; $(date +%Y) StraitKubeGateway Project. Distributed under Apache 2.0 / Open Source License.
        </footer>
    </div>

    <!-- Script for copy to clipboard -->
    <script>
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
