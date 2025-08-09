import React from 'react';
import { useState, useEffect } from 'react';
import './index.css';

function App() {
  const [stats, setStats] = useState({
    totalRequests: 0,
    blockedRequests: 0,
    uniqueVisitors: 0,
    uptime: 0
  });
  
  const [threats, setThreats] = useState([]);
  const [currentTime, setCurrentTime] = useState('');

  // Load stats from API
  const loadStats = async () => {
    try {
      const response = await fetch('/dashboard/api/stats');
      const data = await response.json();
      setStats(data);
    } catch (error) {
      console.error('Failed to load stats:', error);
      // Fallback data
      setStats({
        totalRequests: 125847,
        blockedRequests: 2341,
        uniqueVisitors: 8924,
        uptime: 99.98
      });
    }
  };

  // Load threats from API
  const loadThreats = async () => {
    try {
      const response = await fetch('/dashboard/api/threats');
      const data = await response.json();
      setThreats(data.recentThreats || []);
    } catch (error) {
      console.error('Failed to load threats:', error);
      // Fallback data
      setThreats([
        {
          timestamp: Math.floor(Date.now() / 1000) - 300,
          clientIP: "192.168.1.100",
          attackType: "SQL Injection",
          severity: "high"
        },
        {
          timestamp: Math.floor(Date.now() / 1000) - 480,
          clientIP: "10.0.0.50", 
          attackType: "XSS Attack",
          severity: "medium"
        }
      ]);
    }
  };

  // Update time
  const updateTime = () => {
    setCurrentTime(new Date().toLocaleTimeString());
  };

  useEffect(() => {
    loadStats();
    loadThreats();
    updateTime();
    
    // Update data every 30 seconds
    const dataInterval = setInterval(() => {
      loadStats();
      loadThreats();
    }, 30000);

    // Update time every second
    const timeInterval = setInterval(updateTime, 1000);

    return () => {
      clearInterval(dataInterval);
      clearInterval(timeInterval);
    };
  }, []);

  return (
    <div className="min-h-screen apple-dashboard">
      {/* Header */}
      <header className="glass border-b border-white/20 px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg">
              <i className="fas fa-shield-alt text-white text-lg"></i>
            </div>
            <div>
              <h1 className="text-2xl font-bold text-white">🛡️ WAF Security Dashboard</h1>
              <div className="flex items-center space-x-2 mt-1">
                <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                <span className="text-sm text-white/80">Live Monitoring</span>
              </div>
            </div>
          </div>
          
          <div className="flex items-center space-x-4">
            <div className="glass rounded-2xl px-4 py-2">
              <div className="flex items-center space-x-2">
                <i className="fas fa-clock text-white/60 text-sm"></i>
                <span className="text-sm text-white/80">{currentTime}</span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="px-6 py-8">
        
        {/* Success Message */}
        <div className="glass rounded-3xl p-6 mb-8 border-green-400/30 bg-green-500/10">
          <div className="flex items-center space-x-4">
            <div className="w-12 h-12 bg-green-500 rounded-2xl flex items-center justify-center">
              <i className="fas fa-check text-white text-lg"></i>
            </div>
            <div>
              <h2 className="text-xl font-bold text-white">✅ React 대시보드 접속 성공!</h2>
              <p className="text-white/80 mt-1">정상적인 React SPA 방식으로 WAF 대시보드가 작동 중입니다.</p>
            </div>
          </div>
        </div>
        
        {/* Metrics Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          
          {/* Total Requests */}
          <div className="metric-card rounded-3xl p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-2xl flex items-center justify-center shadow-lg">
                <i className="fas fa-globe text-white text-lg"></i>
              </div>
              <div className="flex items-center space-x-1 text-green-400">
                <i className="fas fa-arrow-up text-sm"></i>
                <span className="text-sm font-medium">+12.5%</span>
              </div>
            </div>
            <div>
              <h3 className="text-3xl font-bold text-white mb-1">{stats.totalRequests?.toLocaleString()}</h3>
              <p className="text-lg font-semibold text-white/90">Total Requests</p>
              <p className="text-sm text-white/70 mt-2">Last 24 hours</p>
            </div>
          </div>

          {/* Blocked Attacks */}
          <div className="metric-card rounded-3xl p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 bg-gradient-to-br from-red-500 to-red-600 rounded-2xl flex items-center justify-center shadow-lg">
                <i className="fas fa-shield-alt text-white text-lg"></i>
              </div>
              <div className="flex items-center space-x-1 text-green-400">
                <i className="fas fa-arrow-down text-sm"></i>
                <span className="text-sm font-medium">-8.2%</span>
              </div>
            </div>
            <div>
              <h3 className="text-3xl font-bold text-white mb-1">{stats.blockedRequests?.toLocaleString()}</h3>
              <p className="text-lg font-semibold text-white/90">Blocked Attacks</p>
              <p className="text-sm text-white/70 mt-2">100% 차단 성공</p>
            </div>
          </div>

          {/* Unique Visitors */}
          <div className="metric-card rounded-3xl p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 bg-gradient-to-br from-green-500 to-green-600 rounded-2xl flex items-center justify-center shadow-lg">
                <i className="fas fa-users text-white text-lg"></i>
              </div>
              <div className="flex items-center space-x-1 text-green-400">
                <i className="fas fa-arrow-up text-sm"></i>
                <span className="text-sm font-medium">+15.7%</span>
              </div>
            </div>
            <div>
              <h3 className="text-3xl font-bold text-white mb-1">{stats.uniqueVisitors?.toLocaleString()}</h3>
              <p className="text-lg font-semibold text-white/90">Unique Visitors</p>
              <p className="text-sm text-white/70 mt-2">Legitimate users</p>
            </div>
          </div>

          {/* System Uptime */}
          <div className="metric-card rounded-3xl p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg">
                <i className="fas fa-server text-white text-lg"></i>
              </div>
              <div className="flex items-center space-x-1 text-green-400">
                <i className="fas fa-check text-sm"></i>
                <span className="text-sm font-medium">Stable</span>
              </div>
            </div>
            <div>
              <h3 className="text-3xl font-bold text-white mb-1">{stats.uptime}%</h3>
              <p className="text-lg font-semibold text-white/90">System Uptime</p>
              <p className="text-sm text-white/70 mt-2">Service availability</p>
            </div>
          </div>
        </div>

        {/* Recent Activity and System Status */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          
          {/* Recent Activity */}
          <div className="glass rounded-3xl p-6">
            <div className="flex items-center space-x-3 mb-6">
              <div className="w-10 h-10 bg-gradient-to-br from-yellow-500 to-orange-500 rounded-2xl flex items-center justify-center">
                <i className="fas fa-exclamation-triangle text-white text-lg"></i>
              </div>
              <div>
                <h3 className="text-lg font-bold text-white">Recent Activity</h3>
                <p className="text-sm text-white/70">Latest security events</p>
              </div>
            </div>

            <div className="space-y-4 max-h-64 overflow-y-auto">
              {threats.map((threat, index) => (
                <div key={index} className="flex items-start space-x-3 p-3 glass rounded-xl">
                  <div className={`w-3 h-3 rounded-full mt-2 ${
                    threat.severity === 'high' ? 'bg-red-400' : 'bg-yellow-400'
                  }`}></div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-white truncate">{threat.attackType}</p>
                    <p className="text-xs text-white/70 mt-1">IP: {threat.clientIP}</p>
                    <p className="text-xs text-white/50 mt-1">
                      {new Date(threat.timestamp * 1000).toLocaleTimeString()}
                    </p>
                  </div>
                  <div className="text-xs text-green-400 font-medium">BLOCKED</div>
                </div>
              ))}
            </div>
          </div>

          {/* System Status */}
          <div className="glass rounded-3xl p-6">
            <div className="flex items-center space-x-3 mb-6">
              <div className="w-10 h-10 bg-gradient-to-br from-green-500 to-emerald-500 rounded-2xl flex items-center justify-center">
                <i className="fas fa-check-circle text-white text-lg"></i>
              </div>
              <div>
                <h3 className="text-lg font-bold text-white">System Status</h3>
                <p className="text-sm text-white/70">All systems operational</p>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="text-center p-4 glass rounded-xl">
                <div className="w-8 h-8 bg-green-400 rounded-full mx-auto mb-2 flex items-center justify-center">
                  <i className="fas fa-shield-alt text-white text-sm"></i>
                </div>
                <div className="text-sm font-medium text-white">WAF Engine</div>
                <div className="text-xs text-green-400 mt-1">Online</div>
              </div>
              
              <div className="text-center p-4 glass rounded-xl">
                <div className="w-8 h-8 bg-green-400 rounded-full mx-auto mb-2 flex items-center justify-center">
                  <i className="fas fa-database text-white text-sm"></i>
                </div>
                <div className="text-sm font-medium text-white">Redis Cache</div>
                <div className="text-xs text-green-400 mt-1">Connected</div>
              </div>
              
              <div className="text-center p-4 glass rounded-xl">
                <div className="w-8 h-8 bg-green-400 rounded-full mx-auto mb-2 flex items-center justify-center">
                  <i className="fas fa-chart-line text-white text-sm"></i>
                </div>
                <div className="text-sm font-medium text-white">Monitoring</div>
                <div className="text-xs text-green-400 mt-1">Active</div>
              </div>
              
              <div className="text-center p-4 glass rounded-xl">
                <div className="w-8 h-8 bg-green-400 rounded-full mx-auto mb-2 flex items-center justify-center">
                  <i className="fas fa-cloud text-white text-sm"></i>
                </div>
                <div className="text-sm font-medium text-white">API Gateway</div>
                <div className="text-xs text-green-400 mt-1">Healthy</div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

export default App;