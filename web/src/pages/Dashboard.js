import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import {
  Shield,
  Activity,
  Users,
  Globe,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  CheckCircle,
  Clock,
  Zap
} from 'lucide-react';

// Components
import MetricCard from '../components/MetricCard';
import ThreatChart from '../components/ThreatChart';
import GeographicMap from '../components/GeographicMap';
import RecentActivity from '../components/RecentActivity';
import SystemStatus from '../components/SystemStatus';

const Dashboard = ({ realTimeData }) => {
  const [metrics, setMetrics] = useState({
    totalRequests: 125847,
    blockedRequests: 2341,
    blockRate: 1.86,
    uniqueVisitors: 8924,
    avgResponseTime: 245.6,
    uptime: 99.98
  });

  const [threatStats, setThreatStats] = useState([
    { type: 'SQL Injection', count: 456, trend: 'up', severity: 'high' },
    { type: 'XSS', count: 234, trend: 'down', severity: 'medium' },
    { type: 'Path Traversal', count: 189, trend: 'up', severity: 'medium' },
    { type: 'Command Injection', count: 123, trend: 'down', severity: 'high' }
  ]);

  // Animation variants
  const containerVariants = {
    initial: { opacity: 0 },
    animate: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1,
        delayChildren: 0.2
      }
    }
  };

  const itemVariants = {
    initial: { y: 20, opacity: 0 },
    animate: {
      y: 0,
      opacity: 1,
      transition: {
        type: "spring",
        stiffness: 300,
        damping: 30
      }
    }
  };

  // Update metrics with real-time data
  useEffect(() => {
    if (realTimeData) {
      setMetrics(prev => ({
        ...prev,
        totalRequests: prev.totalRequests + (realTimeData.newRequests || 0),
        blockedRequests: prev.blockedRequests + (realTimeData.newBlocked || 0)
      }));
    }
  }, [realTimeData]);

  const getMetricIcon = (type) => {
    const icons = {
      requests: Globe,
      blocked: Shield,
      visitors: Users,
      uptime: Activity
    };
    return icons[type] || Activity;
  };

  const getTrendIcon = (trend) => {
    return trend === 'up' ? TrendingUp : TrendingDown;
  };

  const getSeverityColor = (severity) => {
    const colors = {
      high: 'text-red-500',
      medium: 'text-yellow-500',
      low: 'text-blue-500'
    };
    return colors[severity] || 'text-gray-500';
  };

  return (
    <motion.div
      variants={containerVariants}
      initial="initial"
      animate="animate"
      className="space-y-8"
    >
      {/* Page Header */}
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold text-apple-900 dark:text-apple-50 mb-2">
            Security Dashboard
          </h1>
          <p className="text-lg text-apple-600 dark:text-apple-400">
            Real-time WAF monitoring and threat analytics
          </p>
        </div>
        
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2 glass rounded-2xl px-4 py-2">
            <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse" />
            <span className="text-sm font-medium text-apple-700 dark:text-apple-300">
              Live Monitoring
            </span>
          </div>
          
          <div className="glass rounded-2xl px-4 py-2">
            <div className="flex items-center space-x-2">
              <Clock className="w-4 h-4 text-apple-500" />
              <span className="text-sm text-apple-600 dark:text-apple-400">
                Last updated: {new Date().toLocaleTimeString()}
              </span>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Key Metrics Grid */}
      <motion.div 
        variants={itemVariants}
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
      >
        <MetricCard
          title="Total Requests"
          value={metrics.totalRequests.toLocaleString()}
          change="+12.5%"
          trend="up"
          icon={Globe}
          color="blue"
          description="Last 24 hours"
        />
        
        <MetricCard
          title="Blocked Attacks"
          value={metrics.blockedRequests.toLocaleString()}
          change="-8.2%"
          trend="down"
          icon={Shield}
          color="red"
          description="Threats neutralized"
        />
        
        <MetricCard
          title="Unique Visitors"
          value={metrics.uniqueVisitors.toLocaleString()}
          change="+15.7%"
          trend="up"
          icon={Users}
          color="green"
          description="Legitimate users"
        />
        
        <MetricCard
          title="System Uptime"
          value={`${metrics.uptime}%`}
          change="99.98%"
          trend="up"
          icon={Activity}
          color="green"
          description="Service availability"
        />
      </motion.div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        {/* Left Column - Charts */}
        <div className="lg:col-span-2 space-y-8">
          
          {/* Threat Analysis Chart */}
          <motion.div variants={itemVariants} className="card">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h3 className="text-xl font-bold text-apple-900 dark:text-apple-50">
                  Threat Analysis
                </h3>
                <p className="text-apple-600 dark:text-apple-400 mt-1">
                  Attack patterns over the last 24 hours
                </p>
              </div>
              
              <div className="flex items-center space-x-2">
                <button className="px-4 py-2 text-sm font-medium bg-primary-100 text-primary-700 rounded-xl hover:bg-primary-200 transition-colors">
                  24h
                </button>
                <button className="px-4 py-2 text-sm font-medium text-apple-600 dark:text-apple-400 hover:bg-apple-100 dark:hover:bg-apple-700 rounded-xl transition-colors">
                  7d
                </button>
                <button className="px-4 py-2 text-sm font-medium text-apple-600 dark:text-apple-400 hover:bg-apple-100 dark:hover:bg-apple-700 rounded-xl transition-colors">
                  30d
                </button>
              </div>
            </div>
            
            <div className="h-80">
              <ThreatChart data={threatStats} />
            </div>
          </motion.div>

          {/* Geographic Attack Map */}
          <motion.div variants={itemVariants} className="card">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h3 className="text-xl font-bold text-apple-900 dark:text-apple-50">
                  Geographic Threats
                </h3>
                <p className="text-apple-600 dark:text-apple-400 mt-1">
                  Attack origins worldwide
                </p>
              </div>
              
              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2">
                  <div className="w-3 h-3 bg-red-500 rounded-full" />
                  <span className="text-xs text-apple-600 dark:text-apple-400">High Risk</span>
                </div>
                <div className="flex items-center space-x-2">
                  <div className="w-3 h-3 bg-yellow-500 rounded-full" />
                  <span className="text-xs text-apple-600 dark:text-apple-400">Medium Risk</span>
                </div>
                <div className="flex items-center space-x-2">
                  <div className="w-3 h-3 bg-blue-500 rounded-full" />
                  <span className="text-xs text-apple-600 dark:text-apple-400">Low Risk</span>
                </div>
              </div>
            </div>
            
            <div className="h-96">
              <GeographicMap />
            </div>
          </motion.div>
        </div>

        {/* Right Column - Sidebar */}
        <div className="space-y-8">
          
          {/* System Status */}
          <motion.div variants={itemVariants}>
            <SystemStatus />
          </motion.div>

          {/* Recent Activity */}
          <motion.div variants={itemVariants}>
            <RecentActivity />
          </motion.div>

          {/* Top Threats */}
          <motion.div variants={itemVariants} className="card">
            <div className="flex items-center space-x-3 mb-6">
              <div className="w-10 h-10 bg-gradient-to-br from-red-500 to-orange-500 rounded-2xl flex items-center justify-center">
                <AlertTriangle className="w-6 h-6 text-white" />
              </div>
              <div>
                <h3 className="text-lg font-bold text-apple-900 dark:text-apple-50">
                  Active Threats
                </h3>
                <p className="text-sm text-apple-600 dark:text-apple-400">
                  Current threat landscape
                </p>
              </div>
            </div>

            <div className="space-y-4">
              {threatStats.map((threat, index) => {
                const TrendIcon = getTrendIcon(threat.trend);
                
                return (
                  <motion.div
                    key={threat.type}
                    initial={{ opacity: 0, x: 20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: index * 0.1 }}
                    className="flex items-center justify-between p-4 glass rounded-xl hover:bg-apple-50/50 dark:hover:bg-apple-800/30 transition-colors"
                  >
                    <div className="flex items-center space-x-3">
                      <div className={`w-3 h-3 rounded-full ${getSeverityColor(threat.severity)} bg-current`} />
                      <div>
                        <div className="font-medium text-apple-900 dark:text-apple-100">
                          {threat.type}
                        </div>
                        <div className="text-sm text-apple-500 dark:text-apple-400">
                          {threat.count} attempts
                        </div>
                      </div>
                    </div>
                    
                    <div className={`flex items-center space-x-1 ${
                      threat.trend === 'up' ? 'text-red-500' : 'text-green-500'
                    }`}>
                      <TrendIcon className="w-4 h-4" />
                    </div>
                  </motion.div>
                );
              })}
            </div>
          </motion.div>

          {/* Quick Actions */}
          <motion.div variants={itemVariants} className="card">
            <h3 className="text-lg font-bold text-apple-900 dark:text-apple-50 mb-4">
              Quick Actions
            </h3>
            
            <div className="space-y-3">
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className="w-full flex items-center justify-between p-4 bg-primary-50 dark:bg-primary-900/20 rounded-xl hover:bg-primary-100 dark:hover:bg-primary-900/30 transition-colors"
              >
                <div className="flex items-center space-x-3">
                  <Shield className="w-5 h-5 text-primary-500" />
                  <span className="font-medium text-primary-700 dark:text-primary-400">
                    Enable Emergency Mode
                  </span>
                </div>
                <Zap className="w-4 h-4 text-primary-500" />
              </motion.button>
              
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className="w-full flex items-center justify-between p-4 bg-apple-100 dark:bg-apple-800/50 rounded-xl hover:bg-apple-200 dark:hover:bg-apple-700/50 transition-colors"
              >
                <div className="flex items-center space-x-3">
                  <Activity className="w-5 h-5 text-apple-600 dark:text-apple-400" />
                  <span className="font-medium text-apple-700 dark:text-apple-300">
                    Generate Report
                  </span>
                </div>
              </motion.button>
            </div>
          </motion.div>
        </div>
      </div>
    </motion.div>
  );
};

export default Dashboard;