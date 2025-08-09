import React from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  LayoutDashboard,
  Shield,
  BarChart3,
  Settings,
  Zap,
  X,
  Circle
} from 'lucide-react';

const Sidebar = ({ onClose, isConnected }) => {
  const location = useLocation();

  const navigationItems = [
    {
      name: 'Dashboard',
      href: '/',
      icon: LayoutDashboard,
      description: 'Overview & metrics'
    },
    {
      name: 'Security Center',
      href: '/security',
      icon: Shield,
      description: 'Threat monitoring'
    },
    {
      name: 'Analytics',
      href: '/analytics',
      icon: BarChart3,
      description: 'Deep insights'
    },
    {
      name: 'Settings',
      href: '/settings',
      icon: Settings,
      description: 'Configuration'
    }
  ];

  const sidebarVariants = {
    initial: { x: -280, opacity: 0 },
    animate: { 
      x: 0, 
      opacity: 1,
      transition: {
        type: "spring",
        stiffness: 300,
        damping: 30
      }
    }
  };

  const itemVariants = {
    initial: { x: -20, opacity: 0 },
    animate: (i) => ({
      x: 0,
      opacity: 1,
      transition: {
        delay: i * 0.1,
        duration: 0.3
      }
    })
  };

  return (
    <motion.aside
      variants={sidebarVariants}
      initial="initial"
      animate="animate"
      className="flex flex-col h-full w-70 glass border-r border-apple-200/20 dark:border-apple-700/30"
    >
      {/* Header */}
      <div className="flex items-center justify-between p-6 border-b border-apple-200/20 dark:border-apple-700/30">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg">
            <Shield className="w-6 h-6 text-white" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-apple-900 dark:text-apple-50">
              WAF Guard
            </h1>
            <div className="flex items-center space-x-2 mt-1">
              <Circle className={`w-2 h-2 ${isConnected ? 'text-green-500 fill-current' : 'text-red-500 fill-current'}`} />
              <span className="text-xs text-apple-500 dark:text-apple-400">
                {isConnected ? 'Connected' : 'Disconnected'}
              </span>
            </div>
          </div>
        </div>
        
        <button
          onClick={onClose}
          className="lg:hidden p-2 rounded-xl hover:bg-apple-100 dark:hover:bg-apple-700/50 transition-colors"
        >
          <X className="w-5 h-5 text-apple-500" />
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-4 py-6 space-y-2">
        {navigationItems.map((item, index) => {
          const isActive = location.pathname === item.href;
          
          return (
            <motion.div
              key={item.name}
              custom={index}
              variants={itemVariants}
              initial="initial"
              animate="animate"
            >
              <NavLink
                to={item.href}
                onClick={onClose}
                className={`
                  group flex items-center px-4 py-3 rounded-2xl transition-all duration-300 relative overflow-hidden
                  ${isActive 
                    ? 'bg-primary-500 text-white shadow-lg shadow-primary-500/25' 
                    : 'text-apple-700 dark:text-apple-300 hover:bg-apple-100 dark:hover:bg-apple-700/50'
                  }
                `}
              >
                {/* Active indicator */}
                {isActive && (
                  <motion.div
                    layoutId="activeTab"
                    className="absolute inset-0 bg-gradient-to-r from-primary-500 to-primary-600 rounded-2xl"
                    transition={{ type: "spring", stiffness: 400, damping: 30 }}
                  />
                )}
                
                <div className="relative flex items-center space-x-4 z-10">
                  <item.icon className={`w-6 h-6 transition-transform group-hover:scale-110 ${
                    isActive ? 'text-white' : 'text-apple-500 dark:text-apple-400'
                  }`} />
                  <div>
                    <div className="font-semibold text-sm">
                      {item.name}
                    </div>
                    <div className={`text-xs mt-0.5 ${
                      isActive 
                        ? 'text-white/80' 
                        : 'text-apple-500 dark:text-apple-500'
                    }`}>
                      {item.description}
                    </div>
                  </div>
                </div>
              </NavLink>
            </motion.div>
          );
        })}
      </nav>

      {/* System Status */}
      <div className="p-6 border-t border-apple-200/20 dark:border-apple-700/30">
        <div className="glass rounded-2xl p-4 bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-900/20 dark:to-emerald-900/20">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 bg-green-500 rounded-xl flex items-center justify-center shadow-sm">
              <Zap className="w-4 h-4 text-white" />
            </div>
            <div>
              <div className="text-sm font-semibold text-green-700 dark:text-green-400">
                System Online
              </div>
              <div className="text-xs text-green-600 dark:text-green-500">
                All services operational
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="p-6 border-t border-apple-200/20 dark:border-apple-700/30">
        <div className="text-center">
          <p className="text-xs text-apple-500 dark:text-apple-400">
            WAF Security Dashboard v1.0
          </p>
          <p className="text-xs text-apple-400 dark:text-apple-500 mt-1">
            Designed with ❤️ by Jonathan Ive
          </p>
        </div>
      </div>
    </motion.aside>
  );
};

export default Sidebar;