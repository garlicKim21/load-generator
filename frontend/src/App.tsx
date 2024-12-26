import React, { useState, useCallback } from 'react';
import axios from 'axios';
import { LoadButton } from './LoadButton';
import { ErrorMessage } from './ErrorMessage';
import { LoadType, LoadStates } from './types';
import './App.css';

const initialLoadStates: LoadStates = {
  cpu: { loading: false, active: false },
  memory: { loading: false, active: false },
  network: { loading: false, active: false }
};

function App() {
  const [loadStates, setLoadStates] = useState<LoadStates>(initialLoadStates);
  const [error, setError] = useState<string | null>(null);

  const handleError = useCallback((error: any, operation: string) => {
    console.error(`${operation} error:`, error);
    setError(`Failed to ${operation.toLowerCase()}: ${error.message}`);
    setTimeout(() => setError(null), 5000);
  }, []);

  const handleLoad = useCallback(async (type: LoadType) => {
    try {
      setLoadStates(prev => ({
        ...prev,
        [type]: { ...prev[type], loading: true }
      }));

      const action = loadStates[type].active ? 'stop' : 'start';
      await axios.post(`/api/v1/load/${type}/${action}`);

      setLoadStates(prev => ({
        ...prev,
        [type]: { 
          loading: false, 
          active: !prev[type].active 
        }
      }));
    } catch (error) {
      handleError(error, `${type} load operation`);
      setLoadStates(prev => ({
        ...prev,
        [type]: { ...prev[type], loading: false }
      }));
    }
  }, [loadStates, handleError]);

  return (
    <div className="App">
      <header className="App-header">
        <h1>Kubernetes Load Generator</h1>
        <ErrorMessage message={error} />
        <div>
          <LoadButton
            loading={loadStates.cpu.loading}
            active={loadStates.cpu.active}
            onClick={() => handleLoad('cpu')}
            label="CPU"
          />
          
          <LoadButton
            loading={loadStates.memory.loading}
            active={loadStates.memory.active}
            onClick={() => handleLoad('memory')}
            label="Memory"
          />
          
          <LoadButton
            loading={loadStates.network.loading}
            active={loadStates.network.active}
            onClick={() => handleLoad('network')}
            disabled={true}
            label="Network"
          />
        </div>
      </header>
    </div>
  );
}

export default App;