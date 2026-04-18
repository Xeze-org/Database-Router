import React from 'react';

const Controls = ({ trafficLoad, setTrafficLoad, computeScale, setComputeScale, isOverloaded }) => {
  return (
    <>
      <div className="control-row">
        <div className="control-header">
          <span className="name">Traffic Load</span>
          <span className="val" style={{ color: isOverloaded ? 'var(--neon-red)' : 'var(--neon-cyan)' }}>
            {(trafficLoad * 100).toLocaleString()} req/s
          </span>
        </div>
        <input
          type="range"
          min="1"
          max="100"
          value={trafficLoad}
          onChange={(e) => setTrafficLoad(Number(e.target.value))}
        />
        <div className="control-desc">Simulates incoming gRPC connections through Caddy mTLS proxy.</div>
      </div>

      <div className="control-row">
        <div className="control-header">
          <span className="name">Vertical Scaling</span>
          <span className="val" style={{ color: 'var(--neon-purple)' }}>
            {computeScale}x
          </span>
        </div>
        <input
          type="range"
          min="1"
          max="10"
          step="1"
          value={computeScale}
          onChange={(e) => setComputeScale(Number(e.target.value))}
        />
        <div className="control-desc">Allocates more CPU/Memory capacity to the Router.</div>
      </div>

      {isOverloaded && (
        <div className="warning-box">
          ⚠️ <strong>CAPACITY EXCEEDED</strong> — Traffic surpasses router capacity. 
          Increase vertical scaling to stabilize.
        </div>
      )}
    </>
  );
};

export default Controls;
