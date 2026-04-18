import React from 'react';

const StepIndicator = ({ step, stepIndex, total }) => {
  if (!step) return null;

  return (
    <div className="step-overlay">
      <div className="step-progress">
        {Array.from({ length: total }).map((_, i) => (
          <div key={i} className={`step-dot ${i === stepIndex ? 'active' : i < stepIndex ? 'done' : ''}`} />
        ))}
      </div>
      <div className="step-info">
        <span className="step-number">Step {stepIndex + 1}/{total}</span>
        <span className="step-label">{step.label}</span>
      </div>
      <p className="step-desc">{step.desc}</p>
    </div>
  );
};

export default StepIndicator;
