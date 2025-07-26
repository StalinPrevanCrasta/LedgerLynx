import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './App.css';

function App() {
  const [apiStatus, setApiStatus] = useState('Checking...');
  const [services, setServices] = useState({});

  useEffect(() => {
    checkApiStatus();
  }, []);

  const checkApiStatus = async () => {
    try {
      const response = await axios.get('/api/status');
      setApiStatus('Connected');
      setServices(response.data.services);
    } catch (error) {
      setApiStatus('Disconnected');
      console.error('API connection failed:', error);
    }
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>🔗 LedgerLynx</h1>
        <h2>Blockchain-based Chain of Custody</h2>
        
        <div className="status-panel">
          <h3>System Status</h3>
          <p>API Status: <span className={apiStatus === 'Connected' ? 'status-good' : 'status-bad'}>{apiStatus}</span></p>
          
          {Object.keys(services).length > 0 && (
            <div>
              <p>CouchDB: <span className="status-good">{services.couchdb}</span></p>
              <p>IPFS: <span className="status-good">{services.ipfs}</span></p>
            </div>
          )}
        </div>

        <div className="action-panel">
          <button onClick={checkApiStatus}>Refresh Status</button>
        </div>
      </header>
    </div>
  );
}

export default App;
