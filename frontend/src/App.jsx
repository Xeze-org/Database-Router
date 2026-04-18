import React, { useState, useEffect, useCallback, useRef } from 'react';
import './index.css';
import NetworkMap from './components/NetworkMap';
import Controls from './components/Controls';
import TerminalLogs from './components/TerminalLogs';
import StepIndicator from './components/StepIndicator';

/* ─── Sequence steps (real flow: Client → Backend → Caddy → Router → DB) ─── */
const FLOW_STEPS = [
  { id: 0, label: 'User Request',   desc: 'Client App sends HTTP/WS request to Backend Server', from: 'client', to: 'backend' },
  { id: 1, label: 'SDK Resolves',   desc: 'Xeze SDK + Core builds gRPC InsertDataRequest via protobuf', from: 'backend', to: 'backend' },
  { id: 2, label: 'gRPC + mTLS',    desc: 'Core client dials Caddy with mTLS client certificate', from: 'backend', to: 'proxy' },
  { id: 3, label: 'mTLS Verify',    desc: 'Caddy verifies client cert (require_and_verify) against CA', from: 'proxy', to: 'proxy' },
  { id: 4, label: 'h2c Forward',    desc: 'Caddy reverse_proxy h2c://db-router:50051', from: 'proxy', to: 'router' },
  { id: 5, label: 'Server → Service', desc: 'postgres_server.go → protoFieldsToRow → service.InsertData()', from: 'router', to: 'router' },
  { id: 6, label: 'Vault Lookup',   desc: 'Core fetches dynamic credentials from Vault PKI engine', from: 'router', to: 'vault' },
  { id: 7, label: 'DB Query',       desc: 'db.Manager → GetPostgresConnection() → conn.QueryRowContext', from: 'router', to: 'pg' },
  { id: 8, label: 'Response',       desc: 'InsertDataResponse { inserted_id } → SDK → Client App', from: 'pg', to: 'client' },
];

function App() {
  const [trafficLoad, setTrafficLoad] = useState(35);
  const [computeScale, setComputeScale] = useState(2);
  const [activeStep, setActiveStep] = useState(0);
  const [activeDb, setActiveDb] = useState('pg');
  const [logs, setLogs] = useState([]);
  const logIdRef = useRef(0);

  const capacity = computeScale * 100;
  const isOverloaded = trafficLoad > capacity * 0.8;
  const utilization = Math.min(100, Math.round((trafficLoad / capacity) * 100));
  const poolUsage = Math.min(25, Math.max(1, Math.round(trafficLoad / 4)));

  const addLog = useCallback((msg, type = 'info') => {
    logIdRef.current += 1;
    setLogs(prev => {
      const next = [...prev, {
        id: logIdRef.current,
        time: new Date().toISOString().split('T')[1].slice(0, 12),
        msg,
        type,
      }];
      return next.length > 100 ? next.slice(-100) : next;
    });
  }, []);

  // ── Auto-play step sequencer ──
  useEffect(() => {
    const stepMs = Math.max(400, 3000 - (trafficLoad * 25));
    const interval = setInterval(() => {
      setActiveStep(prev => {
        const next = (prev + 1) % FLOW_STEPS.length;
        const step = FLOW_STEPS[next];

        // Pick random DB for each cycle
        if (next === 0) {
          const dbs = ['pg', 'mongo', 'redis'];
          setActiveDb(dbs[Math.floor(Math.random() * dbs.length)]);
        }
        return next;
      });
    }, stepMs);
    return () => clearInterval(interval);
  }, [trafficLoad]);

  // ── Log generation (realistic) ──
  useEffect(() => {
    const intervalMs = Math.max(150, 2500 - (trafficLoad * 22));
    const interval = setInterval(() => {
      const targets = [
        { name: 'PostgreSQL', icon: '🐘', ops: ['ListDatabases', 'ExecuteQuery', 'InsertData', 'SelectData', 'UpdateData', 'DeleteData'] },
        { name: 'MongoDB', icon: '🍃', ops: ['FindDocuments', 'InsertDocument', 'UpdateDocument', 'DeleteDocument', 'ListCollections'] },
        { name: 'Redis', icon: '⚡', ops: ['SetValue', 'GetValue', 'DeleteKey', 'ListKeys', 'Info'] },
      ];
      const t = targets[Math.floor(Math.random() * targets.length)];
      const op = t.ops[Math.floor(Math.random() * t.ops.length)];
      const r = Math.random();

      if (isOverloaded && r > 0.35) {
        const errs = [
          `gRPC UNAVAILABLE — ${t.name} pool exhausted (${poolUsage}/25 conns)`,
          `DEADLINE_EXCEEDED — ${t.icon} ${op}() 30s timeout`,
          `RESOURCE_EXHAUSTED — router capacity at ${utilization}%`,
          `INTERNAL — ${t.name} connection refused`,
        ];
        addLog(errs[Math.floor(Math.random() * errs.length)], 'error');
      } else if (r > 0.88) {
        addLog(`PKI lease renewed for ${t.name} TLS certificate`, 'vault');
      } else if (r > 0.82) {
        addLog(`HealthService.Check${t.name === 'PostgreSQL' ? 'Postgres' : t.name === 'MongoDB' ? 'Mongo' : 'Redis'}() → connected`, 'info');
      } else {
        const latency = Math.round(1 + Math.random() * (isOverloaded ? 450 : 12));
        addLog(`${t.icon} ${op}() → ${latency}ms`, 'info');
      }
    }, intervalMs);
    return () => clearInterval(interval);
  }, [trafficLoad, isOverloaded, addLog, utilization, poolUsage]);

  const currentStep = FLOW_STEPS[activeStep];

  return (
    <>
      <div className="grid-bg" />
      <div className="app-container">
        {/* ── Header ── */}
        <header className="header">
          <div className="header-title">
            <h1>Xeze Database Router</h1>
            <span className="subtitle">High-Level Architecture</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
            <a href="https://code.xeze.org/xeze/Database-Router" target="_blank" rel="noopener noreferrer" className="repo-link">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ width: 14, height: 14 }}>
                <path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22" />
              </svg>
              Repository
            </a>
            <div className={`status-pill ${isOverloaded ? 'degraded' : 'healthy'}`}>
              <div className={`status-dot ${isOverloaded ? 'degraded' : 'healthy'}`} />
              {isOverloaded ? 'DEGRADED' : 'HEALTHY'}
            </div>
          </div>
        </header>

        {/* ── Main Content ── */}
        <div className="main-area">
          {/* ── Network Map Canvas ── */}
          <div className="network-canvas">
            <NetworkMap
              trafficLoad={trafficLoad}
              computeScale={computeScale}
              isOverloaded={isOverloaded}
              activeStep={activeStep}
              activeDb={activeDb}
              poolUsage={poolUsage}
            />
            {/* ── Step Overlay ── */}
            <StepIndicator step={currentStep} stepIndex={activeStep} total={FLOW_STEPS.length} />
          </div>

          {/* ── Sidebar ── */}
          <aside className="sidebar">
            {/* Metrics */}
            <div className="panel-section">
              <div className="panel-label">Live Metrics</div>
              <div className="metric-grid">
                <div className="metric-card">
                  <span className="label">Throughput</span>
                  <span className={`value ${isOverloaded ? 'red' : 'cyan'}`}>{(trafficLoad * 100).toLocaleString()}</span>
                  <span className="label">req/s</span>
                </div>
                <div className="metric-card">
                  <span className="label">Compute</span>
                  <span className="value purple">{computeScale}x</span>
                  <span className="label">vCPU / RAM</span>
                </div>
                <div className="metric-card">
                  <span className="label">Utilization</span>
                  <span className={`value ${utilization > 80 ? 'red' : utilization > 50 ? 'orange' : 'green'}`}>{utilization}%</span>
                  <span className="label">capacity</span>
                </div>
                <div className="metric-card">
                  <span className="label">Pool</span>
                  <span className={`value ${poolUsage > 20 ? 'red' : poolUsage > 10 ? 'orange' : 'cyan'}`}>{poolUsage}/25</span>
                  <span className="label">conn</span>
                </div>
              </div>
            </div>

            {/* Controls */}
            <div className="panel-section">
              <div className="panel-label">System Controls</div>
              <Controls
                trafficLoad={trafficLoad} setTrafficLoad={setTrafficLoad}
                computeScale={computeScale} setComputeScale={setComputeScale}
                isOverloaded={isOverloaded}
              />
            </div>

            {/* Terminal */}
            <div className="panel-section">
              <TerminalLogs logs={logs} />
            </div>
          </aside>
        </div>
      </div>
    </>
  );
}

export default App;
