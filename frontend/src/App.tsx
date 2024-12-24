import React, { useState } from 'react';
import axios from 'axios';
import './App.css';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

function App() {
  const [cpuLoading, setCpuLoading] = useState(false);
  const [cpuActive, setCpuActive] = useState(false);

  const handleCpuLoad = async () => {
    try {
      setCpuLoading(true);
      const action = cpuActive ? 'stop' : 'start';
      await axios.post(`${API_URL}/api/v1/load/cpu/${action}`);
      setCpuActive(!cpuActive);
    } catch (error) {
      console.error('CPU load error:', error);
    } finally {
      setCpuLoading(false);
    }
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>Kubernetes Load Generator</h1>
        <div>
          <button 
            onClick={handleCpuLoad}
            disabled={cpuLoading}
            style={{ 
              backgroundColor: cpuActive ? '#ff4444' : '#4CAF50',
              margin: '10px',
              padding: '15px 30px',
              border: 'none',
              borderRadius: '5px',
              color: 'white',
              cursor: 'pointer'
            }}
          >
            {cpuLoading ? 'Processing...' : 
             cpuActive ? 'Stop CPU Load' : 'Start CPU Load'}
          </button>

          <button 
            disabled
            style={{ 
              margin: '10px',
              padding: '15px 30px',
              border: 'none',
              borderRadius: '5px',
              opacity: 0.5
            }}
          >
            Memory Load (Coming Soon)
          </button>

          <button 
            disabled
            style={{ 
              margin: '10px',
              padding: '15px 30px',
              border: 'none',
              borderRadius: '5px',
              opacity: 0.5
            }}
          >
            Network Load (Coming Soon)
          </button>
        </div>
      </header>
    </div>
  );
}

export default App;