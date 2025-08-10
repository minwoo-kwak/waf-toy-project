import React from 'react';
import './App.css';

function App() {
  return (
    <div className="App">
      <header className="App-header">
        <h1>WAF SaaS Platform</h1>
        <p>Web Application Firewall Dashboard</p>
        <div className="status-card">
          <h3>System Status</h3>
          <p>🟢 WAF Engine: Active</p>
          <p>🟢 OWASP CRS: Loaded</p>
          <p>🟡 Dashboard: Development</p>
        </div>
      </header>
    </div>
  );
}

export default App;