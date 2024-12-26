import React from 'react';

interface LoadButtonProps {
  loading: boolean;
  active: boolean;
  onClick: () => void;
  disabled?: boolean;
  label: string;
}

const buttonBaseStyle = {
  margin: '10px',
  padding: '15px 30px',
  border: 'none',
  borderRadius: '5px',
  color: 'white',
  cursor: 'pointer'
};

export const LoadButton: React.FC<LoadButtonProps> = ({
  loading,
  active,
  onClick,
  disabled = false,
  label
}) => (
  <button
    onClick={onClick}
    disabled={disabled || loading}
    style={{
      ...buttonBaseStyle,
      backgroundColor: active ? '#ff4444' : '#4CAF50',
      opacity: disabled ? 0.5 : 1
    }}
  >
    {loading ? 'Processing...' : 
     active ? `Stop ${label} Load` : `Start ${label} Load`}
  </button>
);