import React, { useMemo, useState } from 'react';

/* ─── ViewBox ─── */
const VB_W = 1100;
const VB_H = 650;

/* ─── Node definitions (centered layout) ─── */
const NODES = {
  client:  { id: 'client',  x: 60,  y: 280, w: 110, h: 60, label: 'Client App', sub: 'Web / Mobile', icon: '📱', color: 'cyan', explainer: 'The entry point for end-user applications. It uses Xeze SDKs to securely connect to the backend over HTTPS/WS.' },
  backend: { id: 'backend', x: 230, y: 255, w: 145, h: 110, label: 'Backend Server', sub: 'mTLS • Core + SDK', icon: '🖥️', color: 'cyan', isBackend: true, explainer: 'The application backend that processes requests, resolves business logic, and uses Xeze Core libraries to interact with the database router via gRPC.' },
  proxy:   { id: 'proxy',   x: 440, y: 280, w: 120, h: 60, label: 'Caddy Proxy', sub: 'mTLS + TLS', icon: '🛡️', color: 'cyan', explainer: 'A high-performance reverse proxy that handles incoming gRPC traffic, terminates external TLS, and enforces strict mTLS authentication before forwarding to the router.' },
  router:  { id: 'router',  x: 630, y: 248, w: 175, h: 125, label: 'Database Router', sub: 'gRPC :50051', icon: '🚦', color: 'purple', isRouter: true, explainer: 'The core Go service that abstracts database complexity. It routes queries to PostgreSQL, MongoDB, or Redis, manages connection pools, and enforces access control.' },
  vault:   { id: 'vault',   x: 655, y: 80,  w: 130, h: 60, label: 'HashiCorp Vault', sub: 'Dynamic Secrets', icon: '🔐', color: 'orange', explainer: 'Provides dynamic, short-lived database credentials and manages PKI for mTLS certificates, ensuring high security and zero standing privileges.' },
  pg:      { id: 'pg',      x: 920, y: 145, w: 130, h: 60, label: 'PostgreSQL', sub: 'Port 5432', icon: '🐘', color: 'green', explainer: 'The primary relational database used for structured, ACID-compliant data storage like user profiles, transactions, and core relational models.' },
  mongo:   { id: 'mongo',   x: 920, y: 280, w: 130, h: 60, label: 'MongoDB', sub: 'Port 27017', icon: '🍃', color: 'green', explainer: 'The NoSQL document database utilized for flexible, schema-less data structures, caching complex JSON documents, and rapid iteration.' },
  redis:   { id: 'redis',   x: 920, y: 415, w: 130, h: 60, label: 'Redis', sub: 'Port 6379', icon: '⚡', color: 'green', explainer: 'An in-memory key-value store used for high-speed caching, session management, and rate-limiting to alleviate load on persistent databases.' },
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

/* ─── Zone definitions ─── */
const ZONES = [
  { x: 40,  y: 215, w: 155, h: 185, label: 'CLIENT',    color: 'rgba(0, 229, 255, 0.04)',  border: 'rgba(0, 229, 255, 0.08)' },
  { x: 215, y: 215, w: 175, h: 185, label: 'APPLICATION', color: 'rgba(0, 229, 255, 0.03)',  border: 'rgba(0, 229, 255, 0.06)' },
  { x: 420, y: 215, w: 160, h: 185, label: 'SECURITY',   color: 'rgba(191, 90, 242, 0.03)', border: 'rgba(191, 90, 242, 0.06)' },
  { x: 610, y: 55,  w: 215, h: 345, label: 'ROUTING',    color: 'rgba(191, 90, 242, 0.03)', border: 'rgba(191, 90, 242, 0.06)' },
  { x: 900, y: 110, w: 170, h: 395, label: 'DATA',       color: 'rgba(48, 209, 88, 0.03)',  border: 'rgba(48, 209, 88, 0.06)' },
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

/* ─── Zone background ─── */
const ZoneRect = ({ zone }) => (
  <g>
    <rect x={zone.x} y={zone.y} width={zone.w} height={zone.h} rx={12}
          fill={zone.color} stroke={zone.border} strokeWidth={0.7} strokeDasharray="6 4" />
    <text x={zone.x + zone.w / 2} y={zone.y + 14}
          style={{ fontSize: '7px', fill: zone.border, fontFamily: 'var(--font-mono)', textAnchor: 'middle', fontWeight: 600, letterSpacing: '2.5px', opacity: 0.8 }}>
      {zone.label}
    </text>
  </g>
);

/* ─── LayeredNode (Router or Backend) ─── */
const LayeredNode = ({ node, layers, scaleExtra = 1, isOverloaded, isActive, poolUsage, layerColor, onClick }) => {
  const n = { ...node };
  if (n.isRouter && scaleExtra > 1) {
    const g = (scaleExtra - 1) * 6;
    n.x -= g; n.y -= g * 0.5; n.w += g * 2; n.h += g;
  }
  const active = isActive ? ' active' : '';
  const overloaded = isOverloaded && n.isRouter ? ' overloaded' : '';
  const borderColor = layerColor || 'rgba(191, 90, 242, 0.2)';

  return (
    <g className={`node-card${active}${overloaded}`} onClick={onClick} style={{ cursor: 'pointer' }}>
      <rect x={n.x} y={n.y} width={n.w} height={n.h} className={`node-bg ${n.color}`} />
      <text x={n.x + n.w / 2} y={n.y + 16} className="node-label" style={{ fontSize: '11px' }}>
        {n.icon} {n.label}
      </text>

      {layers.map((layer, i) => {
        const ly = n.y + 30 + i * 24;
        return (
          <g key={i}>
            <rect x={n.x + 8} y={ly} width={n.w - 16} height={20} rx={4}
                  fill="rgba(255,255,255,0.04)" stroke={borderColor} strokeWidth={0.5} />
            <text x={n.x + n.w / 2} y={ly + 11}
                  style={{ fontSize: '8px', fill: '#999', fontFamily: 'var(--font-mono)', textAnchor: 'middle', dominantBaseline: 'middle' }}>
              {layer.label}
            </text>
          </g>
        );
      })}

      {/* Connection pool bar (router only) */}
      {n.isRouter && poolUsage !== undefined && (
        <g>
          <rect x={n.x + 8} y={n.y + n.h - 10} width={n.w - 16} height={6} rx={3} fill="rgba(255,255,255,0.06)" />
          <rect x={n.x + 8} y={n.y + n.h - 10}
                width={Math.min(n.w - 16, (n.w - 16) * (poolUsage / 25))}
                height={6} rx={3}
                fill={poolUsage > 20 ? 'var(--neon-red)' : poolUsage > 12 ? 'var(--neon-orange)' : 'var(--neon-purple)'} />
          <text x={n.x + n.w - 10} y={n.y + n.h - 5}
                style={{ fontSize: '7px', fill: '#888', fontFamily: 'var(--font-mono)', textAnchor: 'end', dominantBaseline: 'middle' }}>
            {poolUsage}/25
          </text>
        </g>
      )}
    </g>
  );
};

/* ─── Simple NodeCard ─── */
const NodeCard = ({ node, isActive, onClick }) => {
  const active = isActive ? ' active' : '';
  return (
    <g className={`node-card${active}`} onClick={onClick} style={{ cursor: 'pointer' }}>
      <rect x={node.x} y={node.y} width={node.w} height={node.h} className={`node-bg ${node.color}`} />
      <text x={node.x + node.w / 2} y={node.y + 20} className="node-icon">{node.icon}</text>
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
  <g>
    <rect x={x - 30} y={y - 8} width={60} height={16} rx={4}
          fill="rgba(10, 10, 15, 0.7)" stroke={color} strokeWidth={0.3} />
    <text x={x} y={y}
          style={{ fontSize: '7.5px', fill: color, fontFamily: 'var(--font-mono)', textAnchor: 'middle', dominantBaseline: 'middle' }}>
      {text}
    </text>
  </g>
);

/* ─── Main NetworkMap ─── */
const NetworkMap = ({ trafficLoad, computeScale, isOverloaded, activeStep, activeDb, poolUsage }) => {
  const speed = useMemo(() => Math.max(0.4, 3.5 - (trafficLoad / 25)), [trafficLoad]);
  const pDur = useMemo(() => Math.max(0.8, 4 - (trafficLoad / 22)), [trafficLoad]);
  const thick = trafficLoad > 60;
  const numP = Math.min(5, Math.max(1, Math.floor(trafficLoad / 18)));
  const [selectedNode, setSelectedNode] = useState(null);

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

  /* ── Edge label positions (derived from node positions) ── */
  const edgeLabels = useMemo(() => ({
    cbX: (NODES.client.x + NODES.client.w + NODES.backend.x) / 2,
    cbY: NODES.client.y + NODES.client.h / 2 - 14,
    bpX: (NODES.backend.x + NODES.backend.w + NODES.proxy.x) / 2,
    bpY: NODES.proxy.y + NODES.proxy.h / 2 - 14,
    prX: (NODES.proxy.x + NODES.proxy.w + NODES.router.x) / 2,
    prY: NODES.proxy.y + NODES.proxy.h / 2 - 14,
    rvX: (NODES.router.x + NODES.router.w / 2 + NODES.vault.x + NODES.vault.w / 2) / 2 + 15,
    rvY: (NODES.vault.y + NODES.vault.h + NODES.router.y) / 2,
    pgX: (NODES.router.x + NODES.router.w + NODES.pg.x) / 2,
    pgY: (NODES.router.y + NODES.pg.y + NODES.pg.h / 2) / 2,
    mgX: (NODES.router.x + NODES.router.w + NODES.mongo.x) / 2,
    mgY: NODES.mongo.y + NODES.mongo.h / 2,
    rdX: (NODES.router.x + NODES.router.w + NODES.redis.x) / 2,
    rdY: (NODES.router.y + NODES.router.h + NODES.redis.y + NODES.redis.h / 2) / 2 + 10,
  }), []);

  return (
    <>
      <svg viewBox={`0 0 ${VB_W} ${VB_H}`} preserveAspectRatio="xMidYMid meet">
        <defs>
          <radialGradient id="bgGlow" cx="50%" cy="45%" r="55%">
            <stop offset="0%" stopColor="rgba(191, 90, 242, 0.04)" />
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

        {/* ── Zone backgrounds ── */}
        {ZONES.map((z, i) => <ZoneRect key={i} zone={z} />)}

        {/* ── Connections ── */}
        <FlowConnection id="path-cb" path={paths.cb} color="cyan" speed={speed} thick={thick} isActive={isConnActive('client', 'backend')} />
        <FlowConnection id="path-bp" path={paths.bp} color={isOverloaded ? 'red' : 'cyan'} speed={speed} thick={thick} isActive={isConnActive('backend', 'proxy')} />
        <FlowConnection id="path-pr" path={paths.pr} color={isOverloaded ? 'red' : 'cyan'} speed={speed} thick={thick} isActive={isConnActive('proxy', 'router')} />
        <FlowConnection id="path-rv" path={paths.rv} color="orange" speed={speed * 1.5} isActive={isConnActive('router', 'vault')} />
        <FlowConnection id="path-rpg" path={paths.rpg} color="green" speed={speed} thick={thick} isActive={isConnActive('router', 'pg')} />
        <FlowConnection id="path-rmg" path={paths.rmg} color="green" speed={speed} thick={thick} isActive={isConnActive('router', 'mongo')} />
        <FlowConnection id="path-rrd" path={paths.rrd} color="green" speed={speed} thick={thick} isActive={isConnActive('router', 'redis')} />

        {/* ── Edge labels ── */}
        <EdgeLabel x={edgeLabels.cbX} y={edgeLabels.cbY} text="HTTP / WS" color="rgba(0, 229, 255, 0.5)" />
        <EdgeLabel x={edgeLabels.bpX} y={edgeLabels.bpY} text="gRPC + mTLS" color="rgba(0, 229, 255, 0.5)" />
        <EdgeLabel x={edgeLabels.prX} y={edgeLabels.prY} text="h2c :50051" color="rgba(0, 229, 255, 0.5)" />
        <EdgeLabel x={edgeLabels.rvX} y={edgeLabels.rvY} text="PKI" color="rgba(255, 159, 10, 0.5)" />
        <EdgeLabel x={edgeLabels.pgX} y={edgeLabels.pgY} text="sql.DB" color="rgba(48, 209, 88, 0.5)" />
        <EdgeLabel x={edgeLabels.mgX} y={edgeLabels.mgY} text="mongo.Client" color="rgba(48, 209, 88, 0.5)" />
        <EdgeLabel x={edgeLabels.rdX} y={edgeLabels.rdY} text="redis.Client" color="rgba(48, 209, 88, 0.5)" />

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
        <NodeCard node={NODES.client} isActive={stepHighlight.includes('client')} onClick={() => setSelectedNode('client')} />
        <LayeredNode node={NODES.backend} layers={BACKEND_LAYERS} isActive={stepHighlight.includes('backend')} layerColor="rgba(0, 229, 255, 0.2)" onClick={() => setSelectedNode('backend')} />
        <NodeCard node={NODES.proxy} isActive={stepHighlight.includes('proxy')} onClick={() => setSelectedNode('proxy')} />
        <LayeredNode node={NODES.router} layers={ROUTER_LAYERS} scaleExtra={computeScale} isOverloaded={isOverloaded} isActive={stepHighlight.includes('router')} poolUsage={poolUsage} onClick={() => setSelectedNode('router')} />
        <NodeCard node={NODES.vault} isActive={stepHighlight.includes('vault')} onClick={() => setSelectedNode('vault')} />
        <NodeCard node={NODES.pg} isActive={activeDb === 'pg' && activeStep >= 7} onClick={() => setSelectedNode('pg')} />
        <NodeCard node={NODES.mongo} isActive={activeDb === 'mongo' && activeStep >= 7} onClick={() => setSelectedNode('mongo')} />
        <NodeCard node={NODES.redis} isActive={activeDb === 'redis' && activeStep >= 7} onClick={() => setSelectedNode('redis')} />

        {/* ── Overload tint ── */}
        {isOverloaded && (
          <rect x="0" y="0" width={VB_W} height={VB_H} fill="rgba(255, 69, 58, 0.025)" pointerEvents="none" />
        )}

        {/* ── Architecture label ── */}
        <text x={VB_W / 2} y={VB_H - 12}
              style={{ fontSize: '8.5px', fill: 'rgba(255,255,255,0.15)', fontFamily: 'var(--font-mono)', textAnchor: 'middle', letterSpacing: '1.5px' }}>
          Go 1.24 • gRPC • Protobuf • TLS 1.3 • Connection Pooling (25 max)
        </text>

        {/* ── Click hint ── */}
        <text x={VB_W / 2} y={VB_H - 30}
              style={{ fontSize: '7px', fill: 'rgba(255,255,255,0.1)', fontFamily: 'var(--font-mono)', textAnchor: 'middle', letterSpacing: '1px' }}>
          Click any node for details
        </text>
      </svg>

      {/* ── Explainer Pop Up ── */}
      {selectedNode && NODES[selectedNode] && (
        <div className="explainer-modal">
          <div className="explainer-backdrop" onClick={() => setSelectedNode(null)} />
          <div className="explainer-content">
             <button className="close-btn" onClick={() => setSelectedNode(null)}>✕</button>
             <div className="explainer-header">
                <span className="icon">{NODES[selectedNode].icon}</span>
                <h3>{NODES[selectedNode].label}</h3>
             </div>
             <p>{NODES[selectedNode].explainer}</p>
             <div className="explainer-meta">
               <span className="meta-tag">{NODES[selectedNode].sub}</span>
             </div>
          </div>
        </div>
      )}
    </>
  );
};

export default NetworkMap;
