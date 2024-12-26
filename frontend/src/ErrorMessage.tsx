import React from 'react';
import { AlertCircle } from 'lucide-react';

interface ErrorMessageProps {
  message: string | null;
}

export const ErrorMessage: React.FC<ErrorMessageProps> = ({ message }) => {
  if (!message) return null;
  
  return (
    <div style={{
      backgroundColor: 'rgba(211, 47, 47, 0.05)',
      border: '1px solid rgba(211, 47, 47, 0.2)',
      borderRadius: '8px',
      padding: '16px',
      margin: '20px',
      maxWidth: '500px',
      display: 'flex',
      alignItems: 'center',
      gap: '12px',
      boxShadow: '0 2px 4px rgba(0, 0, 0, 0.1)',
      animation: 'slideIn 0.3s ease-out'
    }}>
      <AlertCircle 
        size={24} 
        color="#d32f2f"
        style={{ flexShrink: 0 }}
      />
      <div style={{
        color: '#d32f2f',
        fontSize: '0.95rem',
        fontWeight: 500,
        lineHeight: '1.5',
        margin: 0
      }}>
        {message}
      </div>
    </div>
  );
};