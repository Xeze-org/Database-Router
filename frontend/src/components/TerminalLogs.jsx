import React, { useEffect, useRef } from 'react';

const TerminalLogs = ({ logs }) => {
  const scrollRef = useRef(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs]);

  return (
    <>
      <div className="terminal-header">
        <span className="terminal-title">Live Telemetry</span>
        <div className="terminal-dots">
          <span style={{ background: 'var(--neon-red)' }} />
          <span style={{ background: 'var(--neon-orange)' }} />
          <span style={{ background: 'var(--neon-green)' }} />
        </div>
      </div>
      <div ref={scrollRef} className="terminal-body">
        {logs.length === 0 ? (
          <div style={{ color: '#333', fontStyle: 'italic' }}>Awaiting traffic…</div>
        ) : (
          logs.map((log) => (
            <div key={log.id} className={`log-line ${log.type}`}>
              <span className="timestamp">{log.time}</span>
              {log.msg}
            </div>
          ))
        )}
      </div>
    </>
  );
};

export default TerminalLogs;
