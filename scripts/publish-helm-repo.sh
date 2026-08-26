#!/usr/bin/env bash
# ==============================================================================
# StraitKubeGateway Helm Chart Repository Hosting & Publishing Script
# Generates static assets for https://charts.straitkubegateway.io
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
CHART_DIR="${ROOT_DIR}/straitKubegateway-helm-repo"
OUTPUT_DIR="${1:-${ROOT_DIR}/dist/charts}"
REPO_URL="${HELM_REPO_URL:-https://charts.straitkubegateway.io}"
CUSTOM_DOMAIN="${CUSTOM_DOMAIN:-charts.straitkubegateway.io}"

echo "==> Preparing StraitKubeGateway Helm Chart Repository"
echo "    Chart directory: ${CHART_DIR}"
echo "    Output directory: ${OUTPUT_DIR}"
echo "    Repository URL:   ${REPO_URL}"
echo "    Custom Domain:    ${CUSTOM_DOMAIN}"

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

# 4. Create CNAME file for GitHub Pages / DNS routing
echo "==> Writing CNAME file (${CUSTOM_DOMAIN})..."
echo "${CUSTOM_DOMAIN}" > "${OUTPUT_DIR}/CNAME"

# 5. Create .nojekyll to prevent GitHub Pages Jekyll processing
touch "${OUTPUT_DIR}/.nojekyll"

# 6. Extract chart version and metadata for the landing page
CHART_VERSION=$(grep '^version:' "${CHART_DIR}/Chart.yaml" | awk '{print $2}')
APP_VERSION=$(grep '^appVersion:' "${CHART_DIR}/Chart.yaml" | awk '{print $2}' | tr -d '"')
CHART_NAME=$(grep '^name:' "${CHART_DIR}/Chart.yaml" | awk '{print $2}')

# 7. Generate beautiful modern index.html landing page
cat <<EOF > "${OUTPUT_DIR}/index.html"
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>StraitKubeGateway Helm Charts Repository</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-primary: #0a0d14;
            --bg-card: rgba(18, 24, 38, 0.85);
            --bg-code: #0d111a;
            --border-color: rgba(56, 189, 248, 0.2);
            --border-glow: rgba(56, 189, 248, 0.4);
            --text-main: #f1f5f9;
            --text-muted: #94a3b8;
            --accent-cyan: #38bdf8;
            --accent-blue: #3b82f6;
            --accent-purple: #818cf8;
            --accent-green: #10b981;
            --radius-lg: 16px;
            --radius-md: 10px;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
            background-color: var(--bg-primary);
            color: var(--text-main);
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 2rem 1.5rem;
            position: relative;
            overflow-x: hidden;
        }

        body::before {
            content: '';
            position: absolute;
            top: -20%;
            left: 50%;
            transform: translateX(-50%);
            width: 800px;
            height: 500px;
            background: radial-gradient(circle, rgba(56, 189, 248, 0.15) 0%, rgba(59, 130, 246, 0.08) 50%, transparent 80%);
            z-index: -1;
            pointer-events: none;
        }

        .container {
            max-width: 880px;
            width: 100%;
        }

        .header {
            text-align: center;
            margin-bottom: 2.5rem;
        }

        .badge-pill {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 6px 14px;
            border-radius: 9999px;
            background: rgba(56, 189, 248, 0.1);
            border: 1px solid rgba(56, 189, 248, 0.3);
            color: var(--accent-cyan);
            font-size: 0.85rem;
            font-weight: 600;
            margin-bottom: 1.25rem;
            letter-spacing: 0.02em;
        }

        .badge-pill::before {
            content: '';
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background-color: var(--accent-green);
            box-shadow: 0 0 10px var(--accent-green);
        }

        h1 {
            font-size: 2.5rem;
            font-weight: 800;
            line-height: 1.2;
            background: linear-gradient(135deg, #ffffff 30%, var(--accent-cyan) 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 0.75rem;
        }

        p.subtitle {
            color: var(--text-muted);
            font-size: 1.1rem;
            line-height: 1.6;
            max-width: 640px;
            margin: 0 auto;
        }

        .card {
            background: var(--bg-card);
            backdrop-filter: blur(16px);
            -webkit-backdrop-filter: blur(16px);
            border: 1px solid var(--border-color);
            border-radius: var(--radius-lg);
            padding: 2rem;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4), 0 0 0 1px rgba(255, 255, 255, 0.05);
            margin-bottom: 2rem;
        }

        .card h2 {
            font-size: 1.25rem;
            font-weight: 700;
            color: #ffffff;
            margin-bottom: 1rem;
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .step-num {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            width: 28px;
            height: 28px;
            background: rgba(56, 189, 248, 0.15);
            color: var(--accent-cyan);
            border-radius: 50%;
            font-size: 0.85rem;
            font-weight: 700;
        }

        .code-block {
            position: relative;
            background: var(--bg-code);
            border: 1px solid rgba(255, 255, 255, 0.08);
            border-radius: var(--radius-md);
            padding: 1.1rem 1.25rem;
            margin-bottom: 1.5rem;
            overflow-x: auto;
        }

        .code-block code {
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.95rem;
            color: #38bdf8;
            white-space: pre-wrap;
            word-break: break-all;
        }

        .copy-btn {
            position: absolute;
            top: 10px;
            right: 10px;
            background: rgba(255, 255, 255, 0.08);
            border: 1px solid rgba(255, 255, 255, 0.12);
            border-radius: 6px;
            color: var(--text-muted);
            padding: 4px 10px;
            font-size: 0.75rem;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.2s ease;
        }

        .copy-btn:hover {
            background: rgba(56, 189, 248, 0.2);
            color: #ffffff;
            border-color: var(--accent-cyan);
        }

        .meta-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-top: 1.5rem;
        }

        .meta-item {
            background: rgba(255, 255, 255, 0.03);
            border: 1px solid rgba(255, 255, 255, 0.06);
            border-radius: var(--radius-md);
            padding: 1rem;
        }

        .meta-label {
            font-size: 0.75rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: var(--text-muted);
            margin-bottom: 4px;
        }

        .meta-value {
            font-size: 1rem;
            font-weight: 600;
            color: #ffffff;
        }

        .links-row {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 1.5rem;
            margin-top: 1.5rem;
        }

        .links-row a {
            color: var(--accent-cyan);
            text-decoration: none;
            font-weight: 500;
            font-size: 0.95rem;
            display: flex;
            align-items: center;
            gap: 6px;
            transition: color 0.2s ease;
        }

        .links-row a:hover {
            color: #ffffff;
            text-decoration: underline;
        }

        footer {
            margin-top: 2rem;
            text-align: center;
            color: #64748b;
            font-size: 0.85rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="badge-pill">Official Helm Repository Active</div>
            <h1>StraitKubeGateway Charts</h1>
            <p class="subtitle">
                Kubernetes-native eBPF Transit Gateway, CNI, and Multi-Cluster Service Networking.
            </p>
        </div>

        <div class="card">
            <h2><span class="step-num">1</span> Add Helm Repository</h2>
            <div class="code-block">
                <button class="copy-btn" onclick="copyCode(this, 'helm repo add straitkubegateway https://charts.straitkubegateway.io\nhelm repo update')">Copy</button>
                <code>helm repo add straitkubegateway https://charts.straitkubegateway.io
helm repo update</code>
            </div>

            <h2><span class="step-num">2</span> Install StraitKubeGateway Chart</h2>
            <div class="code-block">
                <button class="copy-btn" onclick="copyCode(this, 'helm install straitkubegateway straitkubegateway/straitkubegateway --namespace kube-system --create-namespace')">Copy</button>
                <code>helm install straitkubegateway straitkubegateway/straitkubegateway \
  --namespace kube-system \
  --create-namespace</code>
            </div>

            <div class="meta-grid">
                <div class="meta-item">
                    <div class="meta-label">Chart Name</div>
                    <div class="meta-value">${CHART_NAME}</div>
                </div>
                <div class="meta-item">
                    <div class="meta-label">Latest Chart Version</div>
                    <div class="meta-value">v${CHART_VERSION}</div>
                </div>
                <div class="meta-item">
                    <div class="meta-label">App Version</div>
                    <div class="meta-value">v${APP_VERSION}</div>
                </div>
                <div class="meta-item">
                    <div class="meta-label">Direct Download</div>
                    <div class="meta-value"><a href="${CHART_NAME}-${CHART_VERSION}.tgz" style="color: var(--accent-cyan); text-decoration: none;">${CHART_NAME}-${CHART_VERSION}.tgz</a></div>
                </div>
            </div>
        </div>

        <div class="links-row">
            <a href="https://github.com/msaeedb40/straitKubegateway" target="_blank" rel="noopener noreferrer">
                <span>GitHub Repository</span> →
            </a>
            <a href="index.yaml" target="_blank">
                <span>Raw index.yaml</span> →
            </a>
        </div>

        <footer>
            &copy; $(date +%Y) StraitKubeGateway Project. All rights reserved.
        </footer>
    </div>

    <script>
        function copyCode(btn, text) {
            navigator.clipboard.writeText(text).then(() => {
                const original = btn.textContent;
                btn.textContent = 'Copied!';
                setTimeout(() => {
                    btn.textContent = original;
                }, 2000);
            });
        }
    </script>
</body>
</html>
EOF

chmod +x "${SCRIPT_DIR}/publish-helm-repo.sh" 2>/dev/null || true

echo "==> Chart repository generated successfully at: ${OUTPUT_DIR}"
