import React, { useMemo, useState, useEffect } from 'react';

/* ─── ViewBox ─── */
const VB_W = 1100;
const VB_H = 700;

/* ─── SDK language support ─── */
const SDK_LANGS = [
  { id: 'py',   label: 'Python',  icon: '🐍', pkg: 'pip install xeze-dbr',         highlight: 'Async-ready • mTLS + Vault certs • Per-service stubs' },
  { id: 'node', label: 'Node.js', icon: '💚', pkg: 'npm install @xeze/dbr',         highlight: 'TypeScript types • mTLS bytes • ESM + CJS' },
  { id: 'go',   label: 'Go',      icon: '🔷', pkg: 'go get .../sdk/go',             highlight: 'Native gRPC • mTLS support • Context propagation' },
  { id: 'rust', label: 'Rust',    icon: '🦀', pkg: 'cargo add xeze-dbr',            highlight: 'Memory safe • Tonic transport • mTLS cert loading' },
  { id: 'java', label: 'Java',    icon: '☕', pkg: 'Maven / Gradle',                highlight: 'Spring Boot • Protobuf bindings • mTLS + TLS 1.3' },
];

const CORE_LANGS = [
  { id: 'py',   label: 'Python',  icon: '🐍', pkg: 'pip install xeze-dbr-core',     highlight: 'One client for all DBs • Auto cert fetch • app_namespace' },
  { id: 'node', label: 'Node.js', icon: '💚', pkg: 'npm install @xeze/dbr-core',    highlight: 'Unified PG+Mongo+Redis • Auto mTLS • Namespace enforce' },
  { id: 'go',   label: 'Go',      icon: '🔷', pkg: 'go get .../core/go',             highlight: 'Single interface • Auto Vault certs • DB-per-service' },
  { id: 'rust', label: 'Rust',    icon: '🦀', pkg: 'cargo add xeze-dbr-core',       highlight: 'Compile-time safe • Unified client • Namespace isolation' },
];

/* ─── Node definitions ─── */
const NODES = {
  client:  { x: 20,  y: 245, w: 100, h: 55, label: 'Client App', sub: 'Web / Mobile', icon: '📱', color: 'cyan' },
  backend: { x: 160, y: 225, w: 130, h: 95, label: 'Backend Server', sub: 'mTLS • Core + SDK', icon: '🖥️', color: 'cyan', isBackend: true },
  proxy:   { x: 340, y: 245, w: 110, h: 55, label: 'Caddy Proxy', sub: 'mTLS + TLS', icon: '🛡️', color: 'cyan' },
  router:  { x: 510, y: 220, w: 160, h: 105, label: 'Database Router', sub: 'gRPC :50051', icon: '🚦', color: 'purple', isRouter: true },
  vault:   { x: 530, y: 50,  w: 120, h: 55, label: 'HashiCorp Vault', sub: 'Dynamic Secrets', icon: '🔐', color: 'orange' },
  pg:      { x: 810, y: 80,  w: 120, h: 55, label: 'PostgreSQL', sub: 'Port 5432', icon: '🐘', color: 'green' },
  mongo:   { x: 810, y: 245, w: 120, h: 55, label: 'MongoDB', sub: 'Port 27017', icon: '🍃', color: 'green' },
  redis:   { x: 810, y: 405, w: 120, h: 55, label: 'Redis', sub: 'Port 6379', icon: '⚡', color: 'green' },
};

/* ─── Internal layers inside router ─── */
const ROUTER_LAYERS = [
  { label: 'gRPC Server', sub: 'internal/server/' },
  { label: 'Service Layer', sub: 'internal/service/' },
  { label: 'DB Manager', sub: 'internal/db/' },
];

/* ─── SDK layers inside backend ─── */
const BACKEND_LAYERS = [
  { label: 'Application Logic' },
  { label: 'Xeze SDK' },
  { label: 'Xeze Core (gRPC)' },
];

/* ─── Connection helpers ─── */
function hBez(a, b) {
  const x1 = a.x + a.w, y1 = a.y + a.h / 2;
  const x2 = b.x, y2 = b.y + b.h / 2;
  const dx = (x2 - x1) * 0.45;
  return `M${x1},${y1} C${x1 + dx},${y1} ${x2 - dx},${y2} ${x2},${y2}`;
}

function vBez(a, b) {
  const x1 = a.x + a.w / 2, y1 = a.y;
  const x2 = b.x + b.w / 2, y2 = b.y + b.h;
  const dy = (y1 - y2) * 0.4;
  return `M${x1},${y1} C${x1},${y1 - dy} ${x2},${y2 + dy} ${x2},${y2}`;
}

/* ─── LayeredNode (Router or Backend) ─── */
const LayeredNode = ({ node, layers, scaleExtra = 1, isOverloaded, isActive, poolUsage, layerColor }) => {
  const n = { ...node };
  if (n.isRouter && scaleExtra > 1) {
    const g = (scaleExtra - 1) * 6;
    n.x -= g; n.y -= g * 0.5; n.w += g * 2; n.h += g;
  }
  const active = isActive ? ' active' : '';
  const overloaded = isOverloaded && n.isRouter ? ' overloaded' : '';
  const borderColor = layerColor || 'rgba(191, 90, 242, 0.2)';

  return (
    <g className={`node-card${active}${overloaded}`}>
      <rect x={n.x} y={n.y} width={n.w} height={n.h} className={`node-bg ${n.color}`} />
      <text x={n.x + n.w / 2} y={n.y + 14} className="node-label" style={{ fontSize: '10px' }}>
        {n.icon} {n.label}
      </text>

      {layers.map((layer, i) => {
        const ly = n.y + 26 + i * 22;
        return (
          <g key={i}>
            <rect x={n.x + 6} y={ly} width={n.w - 12} height={18} rx={3}
                  fill="rgba(255,255,255,0.04)" stroke={borderColor} strokeWidth={0.5} />
            <text x={n.x + n.w / 2} y={ly + 10}
                  style={{ fontSize: '7px', fill: '#999', fontFamily: 'var(--font-mono)', textAnchor: 'middle', dominantBaseline: 'middle' }}>
              {layer.label}
            </text>
          </g>
        );
      })}

      {/* Connection pool bar (router only) */}
      {n.isRouter && poolUsage !== undefined && (
        <g>
          <rect x={n.x + 6} y={n.y + n.h - 8} width={n.w - 12} height={5} rx={2.5} fill="rgba(255,255,255,0.06)" />
          <rect x={n.x + 6} y={n.y + n.h - 8}
                width={Math.min(n.w - 12, (n.w - 12) * (poolUsage / 25))}
                height={5} rx={2.5}
                fill={poolUsage > 20 ? 'var(--neon-red)' : poolUsage > 12 ? 'var(--neon-orange)' : 'var(--neon-purple)'} />
          <text x={n.x + n.w - 8} y={n.y + n.h - 4}
                style={{ fontSize: '6px', fill: '#888', fontFamily: 'var(--font-mono)', textAnchor: 'end', dominantBaseline: 'middle' }}>
            {poolUsage}/25
          </text>
        </g>
      )}
    </g>
  );
};

/* ─── Simple NodeCard ─── */
const NodeCard = ({ node, isActive }) => {
  const active = isActive ? ' active' : '';
  return (
    <g className={`node-card${active}`}>
      <rect x={node.x} y={node.y} width={node.w} height={node.h} className={`node-bg ${node.color}`} />
      <text x={node.x + node.w / 2} y={node.y + 18} className="node-icon">{node.icon}</text>
      <text x={node.x + node.w / 2} y={node.y + node.h / 2 + 8} className="node-label">{node.label}</text>
      <text x={node.x + node.w / 2} y={node.y + node.h - 6} className="node-sublabel">{node.sub}</text>
    </g>
  );
};

/* ─── Connection ─── */
const FlowConnection = ({ id, path, color, speed, thick, isActive }) => (
  <g>
    <path d={path} className="connection-bg" />
    <path d={path} id={id}
          className={`connection-flow ${color} ${thick ? 'thick' : ''} ${isActive ? 'glow-active' : ''}`}
          style={{ animationDuration: `${speed}s` }} />
  </g>
);

/* ─── Particle ─── */
const Particle = ({ pathId, color, duration, delay }) => (
  <circle r="3" className={`particle ${color}`}
          style={{ offsetPath: `url(#${pathId})`, animationDuration: `${duration}s`, animationDelay: `${delay}s` }} />
);

/* ─── Label on connection ─── */
const EdgeLabel = ({ x, y, text, color = '#666' }) => (
  <text x={x} y={y}
        style={{ fontSize: '7.5px', fill: color, fontFamily: 'var(--font-mono)', textAnchor: 'middle', dominantBaseline: 'middle' }}>
    {text}
  </text>
);

/* ─── Language Tree (SDK / Core) ─── */
const LangTree = ({ title, langs, baseX, baseY, activeIdx, accentColor }) => {
  const rootX = baseX + 12;
  const rootY = baseY;
  const branchStartX = rootX + 8;

  return (
    <g>
      {/* Title */}
      <text x={rootX} y={rootY}
            style={{ fontSize: '10px', fill: accentColor, fontFamily: 'var(--font-mono)', fontWeight: 700, letterSpacing: '1.5px' }}>
        {title}
      </text>
      {/* Vertical trunk */}
      <line x1={branchStartX} y1={rootY + 6} x2={branchStartX} y2={rootY + 6 + langs.length * 24 - 12}
            stroke="rgba(255,255,255,0.08)" strokeWidth={1} />

      {langs.map((lang, i) => {
        const ly = rootY + 18 + i * 24;
        const isActive = i === activeIdx;
        return (
          <g key={lang.id}>
            {/* Branch line */}
            <line x1={branchStartX} y1={ly + 2} x2={branchStartX + 14} y2={ly + 2}
                  stroke={isActive ? accentColor : 'rgba(255,255,255,0.08)'} strokeWidth={1} />

            {/* Language pill */}
            <rect x={branchStartX + 18} y={ly - 8} width={isActive ? 280 : 130} height={18} rx={4}
                  fill={isActive ? 'rgba(255,255,255,0.06)' : 'rgba(255,255,255,0.02)'}
                  stroke={isActive ? accentColor : 'rgba(255,255,255,0.06)'}
                  strokeWidth={isActive ? 1 : 0.5}
                  style={{ transition: 'all 0.4s ease' }} />

            {/* Icon + name */}
            <text x={branchStartX + 24} y={ly + 2}
                  style={{ fontSize: '9px', fill: isActive ? '#fff' : '#888', fontFamily: 'var(--font-mono)', dominantBaseline: 'middle' }}>
              {lang.icon} {lang.label}
            </text>

            {/* Highlight text (only when active) */}
            {isActive && (
              <text x={branchStartX + 82} y={ly + 2}
                    style={{ fontSize: '7.5px', fill: accentColor, fontFamily: 'var(--font-mono)', dominantBaseline: 'middle', opacity: 0.9 }}>
                — {lang.highlight}
              </text>
            )}

            {/* Glow dot */}
            {isActive && (
              <circle cx={branchStartX + 14} cy={ly + 2} r={3} fill={accentColor}
                      style={{ filter: `drop-shadow(0 0 6px ${accentColor})` }}>
                <animate attributeName="r" values="2;4;2" dur="1.5s" repeatCount="indefinite" />
              </circle>
            )}
          </g>
        );
      })}
    </g>
  );
};

/* ─── Main NetworkMap ─── */
const NetworkMap = ({ trafficLoad, computeScale, isOverloaded, activeStep, activeDb, poolUsage }) => {
  const speed = useMemo(() => Math.max(0.4, 3.5 - (trafficLoad / 25)), [trafficLoad]);
  const pDur = useMemo(() => Math.max(0.8, 4 - (trafficLoad / 22)), [trafficLoad]);
  const thick = trafficLoad > 60;
  const numP = Math.min(5, Math.max(1, Math.floor(trafficLoad / 18)));

  // Cycle highlighted language
  const [hlSdk, setHlSdk] = useState(0);
  const [hlCore, setHlCore] = useState(0);
  useEffect(() => {
    const t1 = setInterval(() => setHlSdk(p => (p + 1) % SDK_LANGS.length), 2200);
    const t2 = setInterval(() => setHlCore(p => (p + 1) % CORE_LANGS.length), 2800);
    return () => { clearInterval(t1); clearInterval(t2); };
  }, []);

  const paths = useMemo(() => ({
    cb: hBez(NODES.client, NODES.backend),
    bp: hBez(NODES.backend, NODES.proxy),
    pr: hBez(NODES.proxy, NODES.router),
    rv: vBez(NODES.router, NODES.vault),
    rpg: hBez(NODES.router, NODES.pg),
    rmg: hBez(NODES.router, NODES.mongo),
    rrd: hBez(NODES.router, NODES.redis),
  }), []);

  // Which connections are "active" based on the current step
  const stepHighlight = activeStep !== undefined ? [
    ['client', 'backend'],   // step 0 — User request
    ['backend', 'backend'],  // step 1 — SDK resolves
    ['backend', 'proxy'],    // step 2 — gRPC + mTLS
    ['proxy', 'proxy'],      // step 3 — mTLS verify
    ['proxy', 'router'],     // step 4 — h2c forward
    ['router', 'router'],    // step 5 — Server + Service
    ['router', 'vault'],     // step 6 — Vault lookup
    ['router', activeDb],    // step 7 — DB query
    [activeDb, 'client'],    // step 8 — Response
  ][activeStep] : [];

  const isConnActive = (from, to) => stepHighlight[0] === from && stepHighlight[1] === to;

  return (
    <svg viewBox={`0 0 ${VB_W} ${VB_H}`} preserveAspectRatio="xMidYMid meet">
      <defs>
        <radialGradient id="bgGlow" cx="50%" cy="40%" r="60%">
          <stop offset="0%" stopColor="rgba(191, 90, 242, 0.035)" />
          <stop offset="100%" stopColor="transparent" />
        </radialGradient>
        <filter id="glow">
          <feGaussianBlur stdDeviation="3" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>

      <rect x="0" y="0" width={VB_W} height={VB_H} fill="url(#bgGlow)" />

      {/* ── Connections ── */}
      <FlowConnection id="path-cb" path={paths.cb} color="cyan" speed={speed} thick={thick} isActive={isConnActive('client', 'backend')} />
      <FlowConnection id="path-bp" path={paths.bp} color={isOverloaded ? 'red' : 'cyan'} speed={speed} thick={thick} isActive={isConnActive('backend', 'proxy')} />
      <FlowConnection id="path-pr" path={paths.pr} color={isOverloaded ? 'red' : 'cyan'} speed={speed} thick={thick} isActive={isConnActive('proxy', 'router')} />
      <FlowConnection id="path-rv" path={paths.rv} color="orange" speed={speed * 1.5} isActive={isConnActive('router', 'vault')} />
      <FlowConnection id="path-rpg" path={paths.rpg} color="green" speed={speed} thick={thick} isActive={isConnActive('router', 'pg')} />
      <FlowConnection id="path-rmg" path={paths.rmg} color="green" speed={speed} thick={thick} isActive={isConnActive('router', 'mongo')} />
      <FlowConnection id="path-rrd" path={paths.rrd} color="green" speed={speed} thick={thick} isActive={isConnActive('router', 'redis')} />

      {/* ── Edge labels ── */}
      <EdgeLabel x={135}  y={268} text="HTTP / WS" color="rgba(0, 229, 255, 0.4)" />
      <EdgeLabel x={290}  y={240} text="gRPC + mTLS" color="rgba(0, 229, 255, 0.5)" />
      <EdgeLabel x={445}  y={240} text="h2c :50051" color="rgba(0, 229, 255, 0.5)" />
      <EdgeLabel x={585}  y={140} text="PKI" color="rgba(255, 159, 10, 0.5)" />
      <EdgeLabel x={745}  y={108} text="sql.DB" color="rgba(48, 209, 88, 0.4)" />
      <EdgeLabel x={745}  y={266} text="mongo.Client" color="rgba(48, 209, 88, 0.4)" />
      <EdgeLabel x={745}  y={426} text="redis.Client" color="rgba(48, 209, 88, 0.4)" />

      {/* ── Particles ── */}
      {Array.from({ length: numP }).map((_, i) => (
        <Particle key={`cb${i}`} pathId="path-cb" color="cyan" duration={pDur} delay={i * (pDur / numP)} />
      ))}
      {Array.from({ length: numP }).map((_, i) => (
        <Particle key={`bp${i}`} pathId="path-bp" color="cyan" duration={pDur} delay={i * (pDur / numP) + 0.1} />
      ))}
      {Array.from({ length: numP }).map((_, i) => (
        <Particle key={`pr${i}`} pathId="path-pr" color="cyan" duration={pDur} delay={i * (pDur / numP) + 0.2} />
      ))}
      {Array.from({ length: Math.max(1, Math.floor(numP / 2)) }).map((_, i) => (
        <Particle key={`rv${i}`} pathId="path-rv" color="orange" duration={pDur * 1.5} delay={i * 2} />
      ))}
      {Array.from({ length: numP }).map((_, i) => (
        <Particle key={`pg${i}`} pathId="path-rpg" color="green" duration={pDur} delay={i * (pDur / numP) + 0.3} />
      ))}
      {Array.from({ length: Math.max(1, numP - 1) }).map((_, i) => (
        <Particle key={`mg${i}`} pathId="path-rmg" color="green" duration={pDur} delay={i * (pDur / (numP - 1 || 1)) + 0.5} />
      ))}
      {Array.from({ length: Math.max(1, numP - 1) }).map((_, i) => (
        <Particle key={`rd${i}`} pathId="path-rrd" color="green" duration={pDur} delay={i * 1.1 + 0.7} />
      ))}

      {/* ── Nodes ── */}
      <NodeCard node={NODES.client} isActive={stepHighlight.includes('client')} />
      <LayeredNode node={NODES.backend} layers={BACKEND_LAYERS} isActive={stepHighlight.includes('backend')} layerColor="rgba(0, 229, 255, 0.2)" />
      <NodeCard node={NODES.proxy} isActive={stepHighlight.includes('proxy')} />
      <LayeredNode node={NODES.router} layers={ROUTER_LAYERS} scaleExtra={computeScale} isOverloaded={isOverloaded} isActive={stepHighlight.includes('router')} poolUsage={poolUsage} />
      <NodeCard node={NODES.vault} isActive={stepHighlight.includes('vault')} />
      <NodeCard node={NODES.pg} isActive={activeDb === 'pg' && activeStep >= 7} />
      <NodeCard node={NODES.mongo} isActive={activeDb === 'mongo' && activeStep >= 7} />
      <NodeCard node={NODES.redis} isActive={activeDb === 'redis' && activeStep >= 7} />

      {/* ── Overload tint ── */}
      {isOverloaded && (
        <rect x="0" y="0" width={VB_W} height={VB_H} fill="rgba(255, 69, 58, 0.025)" pointerEvents="none" />
      )}

      {/* ── SDK / Core Language Trees ── */}
      <LangTree title="SDK  ·  Raw gRPC  ·  Vault Supported" langs={SDK_LANGS} baseX={10} baseY={380} activeIdx={hlSdk} accentColor="var(--neon-cyan)" />
      <LangTree title="CORE  ·  Abstraction  ·  Namespace + SDK" langs={CORE_LANGS} baseX={10} baseY={540} activeIdx={hlCore} accentColor="var(--neon-purple)" />

      {/* Connector line from Backend down to the trees */}
      <line x1={225} y1={320} x2={225} y2={370} stroke="rgba(0, 229, 255, 0.15)" strokeWidth={1} strokeDasharray="3 5" />
      <line x1={225} y1={370} x2={20} y2={370} stroke="rgba(0, 229, 255, 0.15)" strokeWidth={1} strokeDasharray="3 5" />


      {/* ── Architecture labels ── */}
      <text x={VB_W / 2} y={VB_H - 10}
            style={{ fontSize: '8.5px', fill: '#333', fontFamily: 'var(--font-mono)', textAnchor: 'middle', letterSpacing: '1px' }}>
        Go 1.24 • gRPC • Protobuf • TLS 1.3 • Connection Pooling (25 max)
      </text>
    </svg>
  );
};

export default NetworkMap;
