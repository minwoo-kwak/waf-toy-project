import React from 'react';
import { motion } from 'framer-motion';
import { TrendingUp, TrendingDown } from 'lucide-react';

const MetricCard = ({ 
  title, 
  value, 
  change, 
  trend, 
  icon: Icon, 
  color = 'blue', 
  description,
  animate = true 
}) => {
  const getColorClasses = (color) => {
    const colorMap = {
      blue: {
        bg: 'metric-card-blue',
        icon: 'from-blue-500 to-blue-600',
        text: 'text-blue-600 dark:text-blue-400',
        trend: trend === 'up' ? 'text-green-500' : 'text-red-500'
      },
      red: {
        bg: 'metric-card-red',
        icon: 'from-red-500 to-red-600',
        text: 'text-red-600 dark:text-red-400',
        trend: trend === 'down' ? 'text-green-500' : 'text-red-500'
      },
      green: {
        bg: 'metric-card-green',
        icon: 'from-green-500 to-green-600',
        text: 'text-green-600 dark:text-green-400',
        trend: trend === 'up' ? 'text-green-500' : 'text-red-500'
      },
      yellow: {
        bg: 'metric-card-yellow',
        icon: 'from-yellow-500 to-yellow-600',
        text: 'text-yellow-600 dark:text-yellow-400',
        trend: trend === 'up' ? 'text-green-500' : 'text-red-500'
      }
    };
    return colorMap[color] || colorMap.blue;
  };

  const colors = getColorClasses(color);
  const TrendIcon = trend === 'up' ? TrendingUp : TrendingDown;

  const cardVariants = {
    initial: { 
      opacity: 0, 
      y: 20,
      scale: 0.9
    },
    animate: { 
      opacity: 1, 
      y: 0,
      scale: 1,
      transition: {
        type: "spring",
        stiffness: 300,
        damping: 30,
        duration: 0.6
      }
    },
    hover: {
      y: -8,
      scale: 1.02,
      transition: {
        type: "spring",
        stiffness: 400,
        damping: 25
      }
    }
  };

  const iconVariants = {
    initial: { rotate: -10, scale: 0.8 },
    animate: { 
      rotate: 0, 
      scale: 1,
      transition: {
        delay: 0.2,
        type: "spring",
        stiffness: 300,
        damping: 20
      }
    },
    hover: {
      rotate: 5,
      scale: 1.1,
      transition: {
        type: "spring",
        stiffness: 400,
        damping: 15
      }
    }
  };

  const valueVariants = {
    initial: { opacity: 0, x: -20 },
    animate: { 
      opacity: 1, 
      x: 0,
      transition: {
        delay: 0.3,
        duration: 0.5
      }
    }
  };

  return (
    <motion.div
      variants={animate ? cardVariants : {}}
      initial={animate ? "initial" : false}
      animate={animate ? "animate" : false}
      whileHover={animate ? "hover" : false}
      className={`${colors.bg} border rounded-3xl p-6 shadow-lg hover:shadow-xl transition-all duration-300 backdrop-blur-sm relative overflow-hidden`}
    >
      {/* Background Gradient Overlay */}
      <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent pointer-events-none" />
      
      <div className="relative z-10">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <motion.div
            variants={animate ? iconVariants : {}}
            className={`w-12 h-12 bg-gradient-to-br ${colors.icon} rounded-2xl flex items-center justify-center shadow-lg`}
          >
            <Icon className="w-6 h-6 text-white" />
          </motion.div>
          
          {change && (
            <div className={`flex items-center space-x-1 ${colors.trend}`}>
              <TrendIcon className="w-4 h-4" />
              <span className="text-sm font-medium">
                {change}
              </span>
            </div>
          )}
        </div>

        {/* Content */}
        <div className="space-y-2">
          <motion.div
            variants={animate ? valueVariants : {}}
            className="space-y-1"
          >
            <h3 className="text-3xl font-bold text-apple-900 dark:text-apple-50 leading-none">
              {value}
            </h3>
            <h4 className="text-lg font-semibold text-apple-700 dark:text-apple-300">
              {title}
            </h4>
          </motion.div>
          
          {description && (
            <motion.p 
              variants={animate ? valueVariants : {}}
              className="text-sm text-apple-600 dark:text-apple-400"
            >
              {description}
            </motion.p>
          )}
        </div>

        {/* Bottom Accent */}
        <div className="mt-4 pt-4 border-t border-white/10">
          <div className="flex items-center justify-between text-xs">
            <span className="text-apple-500 dark:text-apple-400">
              Last updated
            </span>
            <span className="text-apple-600 dark:text-apple-300 font-medium">
              {new Date().toLocaleTimeString()}
            </span>
          </div>
        </div>
      </div>

      {/* Animated Background Elements */}
      <div className="absolute -top-4 -right-4 w-24 h-24 bg-white/5 rounded-full animate-pulse-slow" />
      <div className="absolute -bottom-6 -left-6 w-32 h-32 bg-white/5 rounded-full animate-pulse-slow" style={{ animationDelay: '1s' }} />
    </motion.div>
  );
};

// Enhanced version with progress bar
export const MetricCardWithProgress = ({ 
  title, 
  value, 
  maxValue,
  change, 
  trend, 
  icon: Icon, 
  color = 'blue', 
  description,
  animate = true 
}) => {
  const percentage = (value / maxValue) * 100;
  const colors = {
    blue: 'bg-blue-500',
    red: 'bg-red-500',
    green: 'bg-green-500',
    yellow: 'bg-yellow-500'
  };

  return (
    <MetricCard
      title={title}
      value={value}
      change={change}
      trend={trend}
      icon={Icon}
      color={color}
      description={description}
      animate={animate}
    >
      {/* Progress Bar */}
      <div className="mt-4">
        <div className="flex items-center justify-between mb-2">
          <span className="text-xs text-apple-500 dark:text-apple-400">
            Progress
          </span>
          <span className="text-xs text-apple-600 dark:text-apple-300 font-medium">
            {percentage.toFixed(1)}%
          </span>
        </div>
        <div className="w-full h-2 bg-apple-200/30 dark:bg-apple-700/30 rounded-full overflow-hidden">
          <motion.div
            initial={{ width: 0 }}
            animate={{ width: `${percentage}%` }}
            transition={{ delay: 0.5, duration: 1, ease: "easeOut" }}
            className={`h-full ${colors[color] || colors.blue} rounded-full`}
          />
        </div>
      </div>
    </MetricCard>
  );
};

export default MetricCard;