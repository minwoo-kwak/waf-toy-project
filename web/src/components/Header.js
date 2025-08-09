import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Menu,
  Sun,
  Moon,
  Bell,
  Search,
  User,
  ChevronDown,
  Globe,
  Activity,
  AlertTriangle
} from 'lucide-react';

const Header = ({ onMenuClick, theme, onThemeToggle, isConnected, realTimeData }) => {
  const [showNotifications, setShowNotifications] = useState(false);
  const [showProfile, setShowProfile] = useState(false);

  const notifications = [
    {
      id: 1,
      type: 'threat',
      title: 'SQL Injection Blocked',
      message: 'Multiple attempts from 192.168.1.100',
      time: '2 minutes ago',
      severity: 'high'
    },
    {
      id: 2,
      type: 'system',
      title: 'Rate Limit Updated',
      message: 'New configuration applied successfully',
      time: '15 minutes ago',
      severity: 'info'
    },
    {
      id: 3,
      type: 'threat',
      title: 'XSS Attack Detected',
      message: 'Malicious payload neutralized',
      time: '1 hour ago',
      severity: 'medium'
    }
  ];

  const getSeverityColor = (severity) => {
    switch (severity) {
      case 'high': return 'text-red-500 bg-red-50 dark:bg-red-900/20';
      case 'medium': return 'text-yellow-500 bg-yellow-50 dark:bg-yellow-900/20';
      case 'info': return 'text-blue-500 bg-blue-50 dark:bg-blue-900/20';
      default: return 'text-gray-500 bg-gray-50 dark:bg-gray-900/20';
    }
  };

  const getSeverityIcon = (severity) => {
    switch (severity) {
      case 'high': 
      case 'medium': 
        return <AlertTriangle className="w-4 h-4" />;
      default: 
        return <Activity className="w-4 h-4" />;
    }
  };

  return (
    <motion.header 
      initial={{ y: -20, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      className="glass border-b border-apple-200/20 dark:border-apple-700/30 px-6 lg:px-8 py-4"
    >
      <div className="flex items-center justify-between">
        {/* Left Section */}
        <div className="flex items-center space-x-4">
          {/* Mobile Menu Button */}
          <button
            onClick={onMenuClick}
            className="lg:hidden p-2 rounded-xl hover:bg-apple-100 dark:hover:bg-apple-700/50 transition-colors"
          >
            <Menu className="w-6 h-6 text-apple-600 dark:text-apple-300" />
          </button>

          {/* Search */}
          <div className="relative hidden md:block">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Search className="w-5 h-5 text-apple-400" />
            </div>
            <input
              type="text"
              placeholder="Search logs, IPs, threats..."
              className="block w-80 pl-10 pr-4 py-3 border border-apple-200 dark:border-apple-700 rounded-2xl
                       bg-white/50 dark:bg-apple-800/50 backdrop-blur-sm
                       text-apple-900 dark:text-apple-100 placeholder-apple-500
                       focus:ring-2 focus:ring-primary-500 focus:border-transparent
                       transition-all duration-200"
            />
          </div>
        </div>

        {/* Center Section - Real-time Stats */}
        <div className="hidden lg:flex items-center space-x-6">
          {/* Connection Status */}
          <div className="flex items-center space-x-2">
            <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`} />
            <span className="text-sm text-apple-600 dark:text-apple-300">
              {isConnected ? 'Live' : 'Offline'}
            </span>
          </div>

          {/* Real-time Metrics */}
          {realTimeData && (
            <motion.div 
              initial={{ scale: 0.8, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              className="flex items-center space-x-4 px-4 py-2 glass rounded-2xl"
            >
              <div className="flex items-center space-x-2">
                <Globe className="w-4 h-4 text-blue-500" />
                <span className="text-sm font-medium text-apple-700 dark:text-apple-300">
                  {realTimeData.activeConnections || 0}
                </span>
                <span className="text-xs text-apple-500">requests/min</span>
              </div>
              <div className="w-px h-4 bg-apple-300 dark:bg-apple-600" />
              <div className="flex items-center space-x-2">
                <AlertTriangle className="w-4 h-4 text-red-500" />
                <span className="text-sm font-medium text-apple-700 dark:text-apple-300">
                  {realTimeData.threatsBlocked || 0}
                </span>
                <span className="text-xs text-apple-500">blocked</span>
              </div>
            </motion.div>
          )}
        </div>

        {/* Right Section */}
        <div className="flex items-center space-x-4">
          {/* Theme Toggle */}
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={onThemeToggle}
            className="p-3 rounded-2xl glass hover:bg-apple-100 dark:hover:bg-apple-700/50 transition-colors"
          >
            <AnimatePresence mode="wait">
              {theme === 'dark' ? (
                <motion.div
                  key="sun"
                  initial={{ rotate: -180, opacity: 0 }}
                  animate={{ rotate: 0, opacity: 1 }}
                  exit={{ rotate: 180, opacity: 0 }}
                  transition={{ duration: 0.3 }}
                >
                  <Sun className="w-5 h-5 text-yellow-500" />
                </motion.div>
              ) : (
                <motion.div
                  key="moon"
                  initial={{ rotate: -180, opacity: 0 }}
                  animate={{ rotate: 0, opacity: 1 }}
                  exit={{ rotate: 180, opacity: 0 }}
                  transition={{ duration: 0.3 }}
                >
                  <Moon className="w-5 h-5 text-apple-600" />
                </motion.div>
              )}
            </AnimatePresence>
          </motion.button>

          {/* Notifications */}
          <div className="relative">
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => setShowNotifications(!showNotifications)}
              className="relative p-3 rounded-2xl glass hover:bg-apple-100 dark:hover:bg-apple-700/50 transition-colors"
            >
              <Bell className="w-5 h-5 text-apple-600 dark:text-apple-300" />
              {notifications.length > 0 && (
                <span className="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">
                  {notifications.length}
                </span>
              )}
            </motion.button>

            {/* Notifications Dropdown */}
            <AnimatePresence>
              {showNotifications && (
                <motion.div
                  initial={{ opacity: 0, y: 10, scale: 0.95 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: 10, scale: 0.95 }}
                  transition={{ duration: 0.2 }}
                  className="absolute right-0 mt-2 w-80 glass rounded-2xl shadow-2xl border border-apple-200/20 dark:border-apple-700/30 z-50"
                >
                  <div className="p-4 border-b border-apple-200/20 dark:border-apple-700/30">
                    <h3 className="text-lg font-semibold text-apple-900 dark:text-apple-100">
                      Notifications
                    </h3>
                  </div>
                  <div className="max-h-96 overflow-y-auto scrollbar-thin">
                    {notifications.map((notification) => (
                      <motion.div
                        key={notification.id}
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        className="p-4 border-b border-apple-200/10 dark:border-apple-700/20 hover:bg-apple-50/50 dark:hover:bg-apple-800/30 transition-colors"
                      >
                        <div className="flex items-start space-x-3">
                          <div className={`p-2 rounded-xl ${getSeverityColor(notification.severity)}`}>
                            {getSeverityIcon(notification.severity)}
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-apple-900 dark:text-apple-100 truncate">
                              {notification.title}
                            </p>
                            <p className="text-sm text-apple-600 dark:text-apple-400 mt-1">
                              {notification.message}
                            </p>
                            <p className="text-xs text-apple-500 dark:text-apple-500 mt-2">
                              {notification.time}
                            </p>
                          </div>
                        </div>
                      </motion.div>
                    ))}
                  </div>
                  <div className="p-4">
                    <button className="w-full text-center text-primary-500 hover:text-primary-600 text-sm font-medium">
                      View All Notifications
                    </button>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          {/* Profile */}
          <div className="relative">
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => setShowProfile(!showProfile)}
              className="flex items-center space-x-3 p-2 rounded-2xl glass hover:bg-apple-100 dark:hover:bg-apple-700/50 transition-colors"
            >
              <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center">
                <User className="w-5 h-5 text-white" />
              </div>
              <div className="hidden md:block text-left">
                <div className="text-sm font-medium text-apple-900 dark:text-apple-100">
                  Admin User
                </div>
                <div className="text-xs text-apple-500 dark:text-apple-400">
                  Security Manager
                </div>
              </div>
              <ChevronDown className="w-4 h-4 text-apple-500" />
            </motion.button>
          </div>
        </div>
      </div>
    </motion.header>
  );
};

export default Header;