/**
 * straitKubegateway — Interactive UI/UX Logic
 */

document.addEventListener('DOMContentLoaded', () => {
  initInstaller();
  initPlatformTabs();
  initKubeadmVersionSelector();
  initTopology();
  initPolicyTester();
  initHookExplorer();
  initCopyButtons();
  initScrollEffects();
});

/* ==========================================================================
   1. Interactive Helm Install Configurator
   ========================================================================== */

let activePreset = 'quickstart';

const presets = {
  quickstart: {
    wireguard: false,
    gwapi: true,
    transit: false,
    bgp: false,
    metrics: true,
    ipv6: false,
    dsr: false
  },
  production: {
    wireguard: true,
    gwapi: true,
    transit: true,
    bgp: true,
    metrics: true,
    ipv6: true,
    dsr: true
  },
  transit: {
    wireguard: true,
    gwapi: true,
    transit: true,
    bgp: true,
    metrics: true,
    ipv6: false,
    dsr: false
  },
  edge: {
    wireguard: false,
    gwapi: false,
    transit: false,
    bgp: false,
    metrics: false,
    ipv6: false,
    dsr: false
  },
  kind: {
    wireguard: false,
    gwapi: true,
    transit: false,
    bgp: false,
    metrics: true,
    ipv6: false,
    dsr: false
  },
  k3s: {
    wireguard: false,
    gwapi: true,
    transit: false,
    bgp: false,
    metrics: true,
    ipv6: false,
    dsr: false
  },
  minikube: {
    wireguard: false,
    gwapi: true,
    transit: false,
    bgp: false,
    metrics: true,
    ipv6: false,
    dsr: false
  },
  kubeadm: {
    wireguard: true,
    gwapi: true,
    transit: false,
    bgp: false,
    metrics: true,
    ipv6: false,
    dsr: false
  }
};

function initInstaller() {
  const presetButtons = document.querySelectorAll('[data-preset]');
  presetButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      presetButtons.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      activePreset = btn.getAttribute('data-preset');
      applyPreset(activePreset);
    });
  });

  const toggles = document.querySelectorAll('.installer-toggle');
  toggles.forEach(toggle => {
    toggle.addEventListener('change', () => {
      renderInstallCommand();
    });
  });

  renderInstallCommand();
}

function applyPreset(name) {
  const cfg = presets[name];
  if (!cfg) return;

  const setVal = (id, val) => {
    const el = document.getElementById(id);
    if (el) el.checked = !!val;
  };

  setVal('opt-wireguard', cfg.wireguard);
  setVal('opt-gwapi', cfg.gwapi);
  setVal('opt-transit', cfg.transit);
  setVal('opt-bgp', cfg.bgp);
  setVal('opt-metrics', cfg.metrics);
  setVal('opt-ipv6', cfg.ipv6);
  setVal('opt-dsr', cfg.dsr);

  renderInstallCommand();
}

function renderInstallCommand() {
  const codeElem = document.getElementById('code-install-cmd');
  if (!codeElem) return;

  const wireguard = document.getElementById('opt-wireguard')?.checked;
  const gwapi = document.getElementById('opt-gwapi')?.checked;
  const transit = document.getElementById('opt-transit')?.checked;
  const bgp = document.getElementById('opt-bgp')?.checked;
  const metrics = document.getElementById('opt-metrics')?.checked;
  const ipv6 = document.getElementById('opt-ipv6')?.checked;
  const dsr = document.getElementById('opt-dsr')?.checked;

  let lines = [
    'helm install straitkubegateway straitkubegateway/straitkubegateway \\',
    '  --namespace kube-system \\',
    '  --create-namespace \\',
    '  --set kubeProxyReplacement=true \\',
    '  --set disableDefaultCNI=true'
  ];

  if (wireguard) {
    lines.push('  --set encryption.wireguard.enabled=true');
  }
  if (!gwapi) {
    lines.push('  --set gatewayAPI.enabled=false');
  }
  if (transit) {
    lines.push('  --set transit.enabled=true');
    lines.push('  --set transit.backboneSegment=0');
  }
  if (bgp) {
    lines.push('  --set routing.bgp.enabled=true');
    lines.push('  --set routing.bfd.enabled=true');
  }
  if (metrics) {
    lines.push('  --set observability.prometheus.enabled=true');
    lines.push('  --set observability.openTelemetry.enabled=true');
  } else {
    lines.push('  --set observability.prometheus.enabled=false');
  }
  if (ipv6) {
    lines.push('  --set networking.dualStack.enabled=true');
  }
  if (dsr) {
    lines.push('  --set service.lb.mode="DSR"');
  }

  if (activePreset === 'kind' || activePreset === 'k3s' || activePreset === 'minikube') {
    lines.push('  --set straitd.image.pullPolicy=IfNotPresent');
    lines.push('  --set sgController.image.pullPolicy=IfNotPresent');
  }

  // Format with backslashes
  const formatted = lines.map((l, i) => {
    if (i === lines.length - 1) return l;
    return l.endsWith('\\') ? l : `${l} \\`;
  }).join('\n');

  codeElem.textContent = formatted;
}

/* ==========================================================================
   1b. Local Platform Cluster Tabs (Kind, K3s, Minikube)
   ========================================================================== */

function initPlatformTabs() {
  const tabs = document.querySelectorAll('.platform-tab');
  const panels = document.querySelectorAll('.platform-panel');

  tabs.forEach(tab => {
    tab.addEventListener('click', () => {
      const targetPlatform = tab.getAttribute('data-platform');
      if (!targetPlatform) return;

      tabs.forEach(t => t.classList.remove('active'));
      tab.classList.add('active');

      panels.forEach(p => {
        if (p.getAttribute('data-panel') === targetPlatform) {
          p.classList.add('active');
        } else {
          p.classList.remove('active');
        }
      });
    });
  });
}

/* ==========================================================================
   1c. Kubeadm v1.m.p Version Dynamic Generator
   ========================================================================== */

function initKubeadmVersionSelector() {
  const minorInput = document.getElementById('kubeadm-minor');
  const patchInput = document.getElementById('kubeadm-patch');
  const quickBtns = document.querySelectorAll('.btn-version-quick');

  function updateKubeadmSnippets() {
    const m = minorInput ? (parseInt(minorInput.value, 10) || 32) : 32;
    const p = patchInput ? (parseInt(patchInput.value, 10) || 0) : 0;

    const repoCode = document.getElementById('code-kubeadm-repo');
    const initCode = document.getElementById('code-kubeadm-init');

    if (repoCode) {
      repoCode.textContent = `sudo apt-get update && sudo apt-get install -y apt-transport-https ca-certificates curl gpg
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.${m}/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.${m}/deb/ /" | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubelet=1.${m}.${p}-* kubeadm=1.${m}.${p}-* kubectl=1.${m}.${p}-*
sudo apt-mark hold kubelet kubeadm kubectl`;
    }

    if (initCode) {
      initCode.textContent = `sudo kubeadm init \\
  --kubernetes-version=v1.${m}.${p} \\
  --pod-network-cidr=10.244.0.0/16 \\
  --skip-phases=addon/kube-proxy \\
  --ignore-preflight-errors=NumCPU,Mem

mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# Untaint control-plane so StraitKubeGateway CNI can schedule
kubectl taint nodes --all node-role.kubernetes.io/control-plane- || true`;
    }

    // Update active quick button
    quickBtns.forEach(btn => {
      const btnM = parseInt(btn.getAttribute('data-m'), 10);
      const btnP = parseInt(btn.getAttribute('data-p'), 10);
      if (btnM === m && btnP === p) {
        btn.classList.add('active');
      } else {
        btn.classList.remove('active');
      }
    });
  }

  if (minorInput) {
    minorInput.addEventListener('input', updateKubeadmSnippets);
  }
  if (patchInput) {
    patchInput.addEventListener('input', updateKubeadmSnippets);
  }

  quickBtns.forEach(btn => {
    btn.addEventListener('click', () => {
      const m = btn.getAttribute('data-m');
      const p = btn.getAttribute('data-p');
      if (minorInput) minorInput.value = m;
      if (patchInput) patchInput.value = p;
      updateKubeadmSnippets();
    });
  });

  updateKubeadmSnippets();
}

/* ==========================================================================
   2. Interactive Topology Visualizer
   ========================================================================== */

const topologyModes = {
  hubspoke: {
    title: 'Hub-and-Spoke Topology (Segment 0 Backbone)',
    desc: 'Cluster-B acts as Central Gateway on Segment 0. Clusters A, C, and D attach via dedicated segment attachments with strict traffic isolation.',
    svg: `
      <g id="hub-mesh-group">
        <!-- Central Backbone Hub -->
        <circle cx="450" cy="200" r="50" fill="rgba(34, 211, 238, 0.12)" stroke="#22d3ee" stroke-width="2.5" />
        <text x="450" y="195" text-anchor="middle" fill="#f8fafc" font-weight="800" font-size="13">Transit Gateway</text>
        <text x="450" y="213" text-anchor="middle" fill="#22d3ee" font-family="JetBrains Mono" font-size="10">Segment 0 (Core)</text>

        <!-- Cluster Nodes -->
        <!-- Cluster A -->
        <g transform="translate(160, 90)">
          <circle cx="0" cy="0" r="38" fill="rgba(59, 130, 246, 0.14)" stroke="#60a5fa" stroke-width="2" />
          <text x="0" y="-4" text-anchor="middle" fill="#f8fafc" font-weight="700" font-size="12">Cluster A</text>
          <text x="0" y="12" text-anchor="middle" fill="#60a5fa" font-family="JetBrains Mono" font-size="9">EU-Central</text>
        </g>

        <!-- Cluster B -->
        <g transform="translate(450, 45)">
          <circle cx="0" cy="0" r="38" fill="rgba(16, 185, 129, 0.14)" stroke="#34d399" stroke-width="2" />
          <text x="0" y="-4" text-anchor="middle" fill="#f8fafc" font-weight="700" font-size="12">Cluster B</text>
          <text x="0" y="12" text-anchor="middle" fill="#34d399" font-family="JetBrains Mono" font-size="9">US-East</text>
        </g>

        <!-- Cluster C -->
        <g transform="translate(740, 90)">
          <circle cx="0" cy="0" r="38" fill="rgba(99, 102, 241, 0.14)" stroke="#818cf8" stroke-width="2" />
          <text x="0" y="-4" text-anchor="middle" fill="#f8fafc" font-weight="700" font-size="12">Cluster C</text>
          <text x="0" y="12" text-anchor="middle" fill="#818cf8" font-family="JetBrains Mono" font-size="9">AP-East</text>
        </g>

        <!-- Cluster D -->
        <g transform="translate(450, 345)">
          <circle cx="0" cy="0" r="38" fill="rgba(245, 158, 11, 0.14)" stroke="#fbbf24" stroke-width="2" />
          <text x="0" y="-4" text-anchor="middle" fill="#f8fafc" font-weight="700" font-size="12">Cluster D</text>
          <text x="0" y="12" text-anchor="middle" fill="#fbbf24" font-family="JetBrains Mono" font-size="9">On-Prem Baremetal</text>
        </g>

        <!-- Dynamic Connection Lines -->
        <line x1="195" y1="105" x2="405" y2="185" stroke="#22d3ee" stroke-width="2" stroke-dasharray="6,4" opacity="0.8" />
        <line x1="450" y1="83" x2="450" y2="150" stroke="#34d399" stroke-width="2" stroke-dasharray="6,4" opacity="0.8" />
        <line x1="705" y1="105" x2="495" y2="185" stroke="#818cf8" stroke-width="2" stroke-dasharray="6,4" opacity="0.8" />
        <line x1="450" y1="307" x2="450" y2="250" stroke="#fbbf24" stroke-width="2" stroke-dasharray="6,4" opacity="0.8" />

        <!-- Segment badges on lines -->
        <rect x="250" y="125" width="80" height="20" rx="4" fill="#0b1329" stroke="rgba(34, 211, 238, 0.3)" />
        <text x="290" y="139" text-anchor="middle" fill="#22d3ee" font-family="JetBrains Mono" font-size="9">Seg 100</text>

        <rect x="580" y="125" width="80" height="20" rx="4" fill="#0b1329" stroke="rgba(129, 140, 248, 0.3)" />
        <text x="620" y="139" text-anchor="middle" fill="#818cf8" font-family="JetBrains Mono" font-size="9">Seg 200</text>

        <rect x="460" y="275" width="80" height="20" rx="4" fill="#0b1329" stroke="rgba(251, 191, 36, 0.3)" />
        <text x="500" y="289" text-anchor="middle" fill="#fbbf24" font-family="JetBrains Mono" font-size="9">Seg 300</text>
      </g>
    `
  },
  peertopeer: {
    title: 'Full Mesh Peer-to-Peer Interconnect',
    desc: 'Direct WireGuard encrypted overlays interconnected across multiple Kubernetes clusters with BGP route exchange and sub-millisecond eBPF fast-path.',
    svg: `
      <g id="peer-mesh-group">
        <!-- Triangle / Quad Mesh -->
        <line x1="220" y1="120" x2="680" y2="120" stroke="#38bdf8" stroke-width="2" stroke-dasharray="8,4" opacity="0.6" />
        <line x1="220" y1="120" x2="450" y2="310" stroke="#34d399" stroke-width="2" stroke-dasharray="8,4" opacity="0.6" />
        <line x1="680" y1="120" x2="450" y2="310" stroke="#818cf8" stroke-width="2" stroke-dasharray="8,4" opacity="0.6" />

        <!-- Node Alpha -->
        <g transform="translate(220, 120)">
          <circle cx="0" cy="0" r="45" fill="rgba(56, 189, 248, 0.15)" stroke="#38bdf8" stroke-width="2.5" />
          <text x="0" y="-6" text-anchor="middle" fill="#f8fafc" font-weight="700" font-size="13">Cluster Alpha</text>
          <text x="0" y="12" text-anchor="middle" fill="#38bdf8" font-family="JetBrains Mono" font-size="10">AWS us-west-2</text>
          <text x="0" y="26" text-anchor="middle" fill="#64748b" font-family="JetBrains Mono" font-size="8">ASN: 65001</text>
        </g>

        <!-- Node Beta -->
        <g transform="translate(680, 120)">
          <circle cx="0" cy="0" r="45" fill="rgba(99, 102, 241, 0.15)" stroke="#818cf8" stroke-width="2.5" />
          <text x="0" y="-6" text-anchor="middle" fill="#f8fafc" font-weight="700" font-size="13">Cluster Beta</text>
          <text x="0" y="12" text-anchor="middle" fill="#818cf8" font-family="JetBrains Mono" font-size="10">GCP europe-west3</text>
          <text x="0" y="26" text-anchor="middle" fill="#64748b" font-family="JetBrains Mono" font-size="8">ASN: 65002</text>
        </g>

        <!-- Node Gamma -->
        <g transform="translate(450, 310)">
          <circle cx="0" cy="0" r="45" fill="rgba(16, 185, 129, 0.15)" stroke="#34d399" stroke-width="2.5" />
          <text x="0" y="-6" text-anchor="middle" fill="#f8fafc" font-weight="700" font-size="13">Cluster Gamma</text>
          <text x="0" y="12" text-anchor="middle" fill="#34d399" font-family="JetBrains Mono" font-size="10">Bare-Metal Equinix</text>
          <text x="0" y="26" text-anchor="middle" fill="#64748b" font-family="JetBrains Mono" font-size="8">ASN: 65003</text>
        </g>

        <!-- Center WireGuard Badge -->
        <rect x="375" y="160" width="150" height="34" rx="8" fill="#0a0f1d" stroke="#22d3ee" stroke-width="1.5" />
        <text x="450" y="181" text-anchor="middle" fill="#22d3ee" font-weight="700" font-size="11">🔒 WireGuard Mesh</text>
      </g>
    `
  },
  segmentrouting: {
    title: 'TransitSegmentAttachment & Route Policy',
    desc: 'Fine-grained policy routing across 32-bit segments (0 to 4,294,967,295). Route policies dynamically propagate via eBPF FIB & BPF Maps without kube-proxy.',
    svg: `
      <g id="segment-route-group">
        <!-- Segment 0 Core -->
        <rect x="120" y="50" width="660" height="60" rx="10" fill="rgba(34, 211, 238, 0.08)" stroke="#22d3ee" stroke-width="1.5" />
        <text x="450" y="85" text-anchor="middle" fill="#22d3ee" font-weight="800" font-size="13">BACKBONE TRANSIT SEGMENT (ID = 0)</text>

        <!-- Attachments -->
        <g transform="translate(180, 190)">
          <rect x="-70" y="-40" width="140" height="80" rx="10" fill="#0d172e" stroke="#60a5fa" stroke-width="1.5" />
          <text x="0" y="-15" text-anchor="middle" fill="#f8fafc" font-weight="700" font-size="11">Payment VNet</text>
          <text x="0" y="5" text-anchor="middle" fill="#60a5fa" font-family="JetBrains Mono" font-size="10">Segment 10</text>
          <text x="0" y="22" text-anchor="middle" fill="#64748b" font-size="9">10.100.0.0/16</text>
          <line x1="0" y1="-40" x2="0" y2="-80" stroke="#60a5fa" stroke-width="2" stroke-dasharray="4,4" />
        </g>

        <g transform="translate(450, 190)">
          <rect x="-70" y="-40" width="140" height="80" rx="10" fill="#0d172e" stroke="#818cf8" stroke-width="1.5" />
          <text x="0" y="-15" text-anchor="middle" fill="#f8fafc" font-weight="700" font-size="11">Analytics VNet</text>
          <text x="0" y="5" text-anchor="middle" fill="#818cf8" font-family="JetBrains Mono" font-size="10">Segment 20</text>
          <text x="0" y="22" text-anchor="middle" fill="#64748b" font-size="9">10.200.0.0/16</text>
          <line x1="0" y1="-40" x2="0" y2="-80" stroke="#818cf8" stroke-width="2" stroke-dasharray="4,4" />
        </g>

        <g transform="translate(720, 190)">
          <rect x="-70" y="-40" width="140" height="80" rx="10" fill="#0d172e" stroke="#34d399" stroke-width="1.5" />
          <text x="0" y="-15" text-anchor="middle" fill="#f8fafc" font-weight="700" font-size="11">Frontend VNet</text>
          <text x="0" y="5" text-anchor="middle" fill="#34d399" font-family="JetBrains Mono" font-size="10">Segment 30</text>
          <text x="0" y="22" text-anchor="middle" fill="#64748b" font-size="9">10.300.0.0/16</text>
          <line x1="0" y1="-40" x2="0" y2="-80" stroke="#34d399" stroke-width="2" stroke-dasharray="4,4" />
        </g>

        <!-- Route Card -->
        <rect x="180" y="290" width="540" height="70" rx="10" fill="rgba(6, 10, 20, 0.9)" stroke="rgba(255,255,255,0.1)" />
        <text x="200" y="315" fill="#22d3ee" font-family="JetBrains Mono" font-size="11" font-weight="700">TransitSegmentRoute:</text>
        <text x="200" y="333" fill="#94a3b8" font-family="JetBrains Mono" font-size="10">destination: 10.200.0.0/16  →  nexthop: Attachment-Seg20  (Allow)</text>
        <text x="200" y="348" fill="#f43f5e" font-family="JetBrains Mono" font-size="10">destination: 10.100.0.0/16  →  nexthop: Attachment-Seg10  (Isolated by Default)</text>
      </g>
    `
  }
};

function initTopology() {
  const topoButtons = document.querySelectorAll('[data-topo]');
  const topoTitle = document.getElementById('topo-title');
  const topoDesc = document.getElementById('topo-desc');
  const topoCanvas = document.getElementById('topo-canvas');

  const update = (mode) => {
    const data = topologyModes[mode];
    if (!data) return;
    if (topoTitle) topoTitle.textContent = data.title;
    if (topoDesc) topoDesc.textContent = data.desc;
    if (topoCanvas) topoCanvas.innerHTML = data.svg;
  };

  topoButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      topoButtons.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      const mode = btn.getAttribute('data-topo');
      update(mode);
    });
  });

  update('hubspoke');
}

/* ==========================================================================
   3. Interactive NetworkPolicy Evaluator Simulator
   ========================================================================== */

function initPolicyTester() {
  const btnEvaluate = document.getElementById('btn-eval-policy');
  if (!btnEvaluate) return;

  btnEvaluate.addEventListener('click', () => {
    const srcNs = document.getElementById('eval-src-ns')?.value || 'default';
    const dstNs = document.getElementById('eval-dst-ns')?.value || 'production';
    const proto = document.getElementById('eval-proto')?.value || 'TCP';
    const port = document.getElementById('eval-port')?.value || '80';
    const segment = document.getElementById('eval-segment')?.value || '0';

    const resultBox = document.getElementById('eval-result');
    const resultVerdict = document.getElementById('eval-verdict');
    const resultReason = document.getElementById('eval-reason');
    const resultLatency = document.getElementById('eval-latency');

    if (!resultBox) return;

    // Simulation logic
    let verdict = 'ALLOW';
    let color = '#10b981';
    let reason = 'Matched Allow Rule (Priority 10: podSelector=api, port=' + port + ')';
    let latency = (Math.random() * 0.4 + 0.1).toFixed(2) + ' µs (eBPF in-kernel)';

    if (dstNs === 'secure' && srcNs !== 'secure') {
      verdict = 'DENY';
      color = '#f43f5e';
      reason = 'Implicit Segment Isolation & Zero-Trust Namespace Boundary';
      latency = '0.08 µs (Early XDP drop)';
    } else if (segment === 'isolated') {
      verdict = 'REJECT';
      color = '#f59e0b';
      reason = 'Segment Policy: Segment ' + segment + ' is not attached to Core Backbone Gateway';
      latency = '0.12 µs (NetKit reject)';
    }

    resultVerdict.textContent = verdict;
    resultVerdict.style.color = color;
    resultReason.textContent = reason;
    resultLatency.textContent = latency;
    resultBox.style.display = 'block';
  });
}

/* ==========================================================================
   4. eBPF Hook Explorer
   ========================================================================== */

const hookDetails = {
  xdp: {
    name: 'XDP (eXpress Data Path)',
    layer: 'NIC Driver / Kernel Ingress',
    purpose: 'Earliest packet processing before sk_buff allocation. Performs L3/L4 DDoS mitigation, packet filtration, and wire-rate forwarding.',
    perf: 'Sub-microsecond (< 0.15 µs)'
  },
  netkit: {
    name: 'NetKit',
    layer: 'Container & Pod Network Namespace',
    purpose: 'Modern Linux kernel networking replacement for legacy veth pairs. Native BPF execution directly on container interface attachments.',
    perf: 'Zero-copy pod-to-host datapath'
  },
  tcx: {
    name: 'TCX (Traffic Control BPF)',
    layer: 'Host Interface Ingress / Egress',
    purpose: 'Handles host-level packet forwarding, service load balancing (Maglev-128 hash), NAT (SNAT/DNAT), and encapsulation (VXLAN/Geneve).',
    perf: 'Full hardware offload ready'
  },
  sockops: {
    name: 'sockops & Socket Hooks',
    layer: 'L4 Transport / TCP Socket',
    purpose: 'Socket-level acceleration for co-located pod communications and connection tracking bypass.',
    perf: 'Eliminates TCP/IP stack overhead'
  },
  lsm: {
    name: 'BPF LSM (Linux Security Module)',
    layer: 'Kernel Security Subsystem',
    purpose: 'Zero-trust security enforcement at the Linux kernel boundary. Intercepts unauthorized raw socket allocation or socket manipulation.',
    perf: 'Deterministic security checks'
  }
};

function initHookExplorer() {
  const hookBtns = document.querySelectorAll('[data-hook]');
  const detailName = document.getElementById('hook-detail-name');
  const detailLayer = document.getElementById('hook-detail-layer');
  const detailPurpose = document.getElementById('hook-detail-purpose');
  const detailPerf = document.getElementById('hook-detail-perf');

  hookBtns.forEach(btn => {
    btn.addEventListener('click', () => {
      hookBtns.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      const hookKey = btn.getAttribute('data-hook');
      const info = hookDetails[hookKey];
      if (!info) return;

      if (detailName) detailName.textContent = info.name;
      if (detailLayer) detailLayer.textContent = info.layer;
      if (detailPurpose) detailPurpose.textContent = info.purpose;
      if (detailPerf) detailPerf.textContent = info.perf;
    });
  });
}

/* ==========================================================================
   5. Copy to Clipboard Utility
   ========================================================================== */

function initCopyButtons() {
  const copyButtons = document.querySelectorAll('.btn-copy, .code-copy-btn');
  copyButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      const targetId = btn.getAttribute('data-copy-target');
      let text = '';
      if (targetId) {
        const targetEl = document.getElementById(targetId);
        if (targetEl) text = targetEl.textContent || targetEl.innerText;
      } else {
        const customText = btn.getAttribute('data-copy-text');
        if (customText) text = customText;
      }

      if (!text) return;

      navigator.clipboard.writeText(text.trim()).then(() => {
        const originalHtml = btn.innerHTML;
        btn.innerHTML = `
          <svg width="14" height="14" fill="none" stroke="#10b981" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"/>
          </svg>
          <span style="color:#10b981;">Copied!</span>
        `;
        btn.classList.add('copied');
        setTimeout(() => {
          btn.innerHTML = originalHtml;
          btn.classList.remove('copied');
        }, 2200);
      });
    });
  });
}

/* ==========================================================================
   6. Scroll & Navbar Interaction
   ========================================================================== */

function initScrollEffects() {
  const navbar = document.getElementById('navbar');
  window.addEventListener('scroll', () => {
    if (window.scrollY > 30) {
      navbar?.classList.add('scrolled');
      navbar.style.borderBottomColor = 'rgba(34, 211, 238, 0.25)';
      navbar.style.boxShadow = '0 10px 30px rgba(0, 0, 0, 0.5)';
    } else {
      navbar?.classList.remove('scrolled');
      navbar.style.borderBottomColor = 'rgba(255, 255, 255, 0.07)';
      navbar.style.boxShadow = 'none';
    }
  });
}
