import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell
} from 'recharts';

const ThreatChart = ({ data = [] }) => {
  const [activeChart, setActiveChart] = useState('line');
  const [chartData, setChartData] = useState([]);

  // Generate mock time-series data
  useEffect(() => {
    const generateTimeData = () => {
      const now = new Date();
      const timeData = [];
      
      for (let i = 23; i >= 0; i--) {
        const hour = new Date(now.getTime() - i * 60 * 60 * 1000);
        timeData.push({
          time: hour.toLocaleTimeString('en-US', { 
            hour: '2-digit', 
            minute: '2-digit', 
            hour12: false 
          }),
          fullTime: hour.toISOString(),
          'SQL Injection': Math.floor(Math.random() * 50) + 10,
          'XSS': Math.floor(Math.random() * 30) + 5,
          'Path Traversal': Math.floor(Math.random() * 25) + 8,
          'Command Injection': Math.floor(Math.random() * 20) + 3,
          'LDAP Injection': Math.floor(Math.random() * 15) + 2,
          total: 0
        });
      }
      
      // Calculate totals
      timeData.forEach(item => {
        item.total = item['SQL Injection'] + item['XSS'] + item['Path Traversal'] + 
                     item['Command Injection'] + item['LDAP Injection'];
      });
      
      return timeData;
    };

    setChartData(generateTimeData());

    // Update data every 30 seconds
    const interval = setInterval(() => {
      setChartData(generateTimeData());
    }, 30000);

    return () => clearInterval(interval);
  }, []);

  const threatColors = {
    'SQL Injection': '#ef4444',
    'XSS': '#f59e0b',
    'Path Traversal': '#8b5cf6',
    'Command Injection': '#ec4899',
    'LDAP Injection': '#06b6d4',
    total: '#3b82f6'
  };

  const pieData = [
    { name: 'SQL Injection', value: 456, color: '#ef4444' },
    { name: 'XSS', value: 234, color: '#f59e0b' },
    { name: 'Path Traversal', value: 189, color: '#8b5cf6' },
    { name: 'Command Injection', value: 123, color: '#ec4899' },
    { name: 'LDAP Injection', value: 67, color: '#06b6d4' }
  ];

  const CustomTooltip = ({ active, payload, label }) => {
    if (active && payload && payload.length) {
      return (
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          className="glass rounded-2xl p-4 border border-apple-200/20 dark:border-apple-700/30 shadow-xl"
        >
          <p className="text-sm font-medium text-apple-900 dark:text-apple-100 mb-2">
            {label}
          </p>
          {payload.map((entry, index) => (
            <div key={index} className="flex items-center space-x-2 mb-1">
              <div 
                className="w-3 h-3 rounded-full" 
                style={{ backgroundColor: entry.color }}
              />
              <span className="text-sm text-apple-700 dark:text-apple-300">
                {entry.name}: {entry.value}
              </span>
            </div>
          ))}
        </motion.div>
      );
    }
    return null;
  };

  const CustomPieTooltip = ({ active, payload }) => {
    if (active && payload && payload[0]) {
      const data = payload[0].payload;
      return (
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          className="glass rounded-2xl p-4 border border-apple-200/20 dark:border-apple-700/30 shadow-xl"
        >
          <div className="flex items-center space-x-2">
            <div 
              className="w-4 h-4 rounded-full" 
              style={{ backgroundColor: data.color }}
            />
            <span className="text-sm font-medium text-apple-900 dark:text-apple-100">
              {data.name}
            </span>
          </div>
          <p className="text-lg font-bold text-apple-900 dark:text-apple-100 mt-1">
            {data.value} attacks
          </p>
          <p className="text-xs text-apple-500 dark:text-apple-400">
            {((data.value / pieData.reduce((sum, item) => sum + item.value, 0)) * 100).toFixed(1)}%
          </p>
        </motion.div>
      );
    }
    return null;
  };

  const chartTypes = [
    { id: 'line', name: 'Line Chart', icon: '📈' },
    { id: 'area', name: 'Area Chart', icon: '📊' },
    { id: 'bar', name: 'Bar Chart', icon: '📊' },
    { id: 'pie', name: 'Pie Chart', icon: '🥧' }
  ];

  const renderChart = () => {
    switch (activeChart) {
      case 'line':
        return (
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis 
                dataKey="time" 
                stroke="#64748b"
                fontSize={12}
              />
              <YAxis 
                stroke="#64748b"
                fontSize={12}
              />
              <Tooltip content={<CustomTooltip />} />
              {Object.keys(threatColors).filter(key => key !== 'total').map(threat => (
                <Line 
                  key={threat}
                  type="monotone" 
                  dataKey={threat} 
                  stroke={threatColors[threat]}
                  strokeWidth={2}
                  dot={{ r: 4 }}
                  activeDot={{ r: 6, fill: threatColors[threat] }}
                />
              ))}
            </LineChart>
          </ResponsiveContainer>
        );

      case 'area':
        return (
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis 
                dataKey="time" 
                stroke="#64748b"
                fontSize={12}
              />
              <YAxis 
                stroke="#64748b"
                fontSize={12}
              />
              <Tooltip content={<CustomTooltip />} />
              {Object.keys(threatColors).filter(key => key !== 'total').map((threat, index) => (
                <Area 
                  key={threat}
                  type="monotone" 
                  dataKey={threat} 
                  stackId="1"
                  stroke={threatColors[threat]}
                  fill={threatColors[threat]}
                  fillOpacity={0.6}
                />
              ))}
            </AreaChart>
          </ResponsiveContainer>
        );

      case 'bar':
        return (
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis 
                dataKey="time" 
                stroke="#64748b"
                fontSize={12}
              />
              <YAxis 
                stroke="#64748b"
                fontSize={12}
              />
              <Tooltip content={<CustomTooltip />} />
              {Object.keys(threatColors).filter(key => key !== 'total').map(threat => (
                <Bar 
                  key={threat}
                  dataKey={threat} 
                  stackId="1"
                  fill={threatColors[threat]}
                />
              ))}
            </BarChart>
          </ResponsiveContainer>
        );

      case 'pie':
        return (
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={pieData}
                cx="50%"
                cy="50%"
                outerRadius={100}
                dataKey="value"
                animationBegin={0}
                animationDuration={800}
              >
                {pieData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip content={<CustomPieTooltip />} />
            </PieChart>
          </ResponsiveContainer>
        );

      default:
        return null;
    }
  };

  return (
    <div className="w-full h-full">
      {/* Chart Type Selector */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-2">
          {chartTypes.map(chart => (
            <motion.button
              key={chart.id}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => setActiveChart(chart.id)}
              className={`px-4 py-2 rounded-xl text-sm font-medium transition-all duration-200 ${
                activeChart === chart.id
                  ? 'bg-primary-500 text-white shadow-lg shadow-primary-500/25'
                  : 'glass text-apple-600 dark:text-apple-400 hover:bg-apple-100 dark:hover:bg-apple-700/50'
              }`}
            >
              <span className="mr-2">{chart.icon}</span>
              {chart.name}
            </motion.button>
          ))}
        </div>

        <div className="flex items-center space-x-2">
          <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse" />
          <span className="text-sm text-apple-600 dark:text-apple-400">
            Live Data
          </span>
        </div>
      </div>

      {/* Legend for Line/Area/Bar charts */}
      {activeChart !== 'pie' && (
        <div className="flex flex-wrap items-center space-x-6 mb-4">
          {Object.entries(threatColors).filter(([key]) => key !== 'total').map(([threat, color]) => (
            <div key={threat} className="flex items-center space-x-2">
              <div className="w-3 h-3 rounded-full" style={{ backgroundColor: color }} />
              <span className="text-sm text-apple-600 dark:text-apple-400">
                {threat}
              </span>
            </div>
          ))}
        </div>
      )}

      {/* Chart Container */}
      <motion.div
        key={activeChart}
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.3 }}
        className="w-full h-full"
      >
        {renderChart()}
      </motion.div>
    </div>
  );
};

export default ThreatChart;