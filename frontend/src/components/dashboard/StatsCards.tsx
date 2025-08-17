import React from 'react';
import {
  Grid,
  Card,
  CardContent,
  Typography,
  Box,
  Skeleton,
  LinearProgress,
} from '@mui/material';
import {
  Security as SecurityIcon,
  Block as BlockIcon,
  Traffic as TrafficIcon,
  Warning as WarningIcon,
} from '@mui/icons-material';
import { WAFStats } from '../../types/waf';

interface StatsCardsProps {
  stats: WAFStats | null;
  loading: boolean;
}

const StatsCards: React.FC<StatsCardsProps> = ({ stats, loading }) => {
  const blockRate = stats
    ? ((stats.blocked_requests / Math.max(stats.total_requests, 1)) * 100).toFixed(1)
    : '0';

  const statsCards = [
    {
      title: 'Total Requests',
      value: stats?.total_requests?.toLocaleString() || '0',
      icon: TrafficIcon,
      color: 'primary.main',
      bgColor: 'rgba(25, 118, 210, 0.1)',
    },
    {
      title: 'Blocked Requests',
      value: stats?.blocked_requests?.toLocaleString() || '0',
      icon: BlockIcon,
      color: 'error.main',
      bgColor: 'rgba(244, 67, 54, 0.1)',
    },
    {
      title: 'Block Rate',
      value: `${blockRate}%`,
      icon: SecurityIcon,
      color: 'success.main',
      bgColor: 'rgba(76, 175, 80, 0.1)',
    },
    {
      title: 'Attack Types',
      value: stats ? Object.keys(stats.attacks_by_type).length.toString() : '0',
      icon: WarningIcon,
      color: 'warning.main',
      bgColor: 'rgba(255, 152, 0, 0.1)',
    },
  ];

  return (
    <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 3 }}>
      {statsCards.map((card, index) => (
        <Card
          key={index}
          elevation={0}
          sx={{
            height: '160px',
            display: 'flex',
            flexDirection: 'column',
            position: 'relative',
            overflow: 'hidden',
            background: 'linear-gradient(145deg, #ffffff 0%, #f8fafc 100%)',
            border: '1px solid #e2e8f0',
            borderRadius: 3,
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            cursor: 'pointer',
            '&:hover': {
              transform: 'translateY(-4px) scale(1.02)',
              boxShadow: '0 20px 40px rgba(0,0,0,0.1)',
              borderColor: card.color,
            }
          }}
        >
          <CardContent sx={{ flex: 1, display: 'flex', flexDirection: 'column', p: 3 }}>
              <Box
                sx={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'flex-start',
                  mb: 1,
                }}
              >
                <Box sx={{ flex: 1 }}>
                  <Typography
                    variant="body2"
                    sx={{ 
                      fontWeight: 600, 
                      color: '#64748b',
                      fontSize: '0.875rem',
                      mb: 1.5,
                      textTransform: 'uppercase',
                      letterSpacing: '0.5px'
                    }}
                  >
                    {card.title}
                  </Typography>
                  {loading ? (
                    <Skeleton width={120} height={40} sx={{ borderRadius: 2 }} />
                  ) : (
                    <Typography 
                      variant="h3" 
                      sx={{ 
                        fontWeight: 800, 
                        color: '#1e293b',
                        fontSize: '2.25rem',
                        lineHeight: 1.2,
                        background: `linear-gradient(135deg, ${card.color} 0%, ${card.color}cc 100%)`,
                        WebkitBackgroundClip: 'text',
                        WebkitTextFillColor: 'transparent',
                        backgroundClip: 'text'
                      }}
                    >
                      {card.value}
                    </Typography>
                  )}
                </Box>
                <Box
                  sx={{
                    p: 2,
                    borderRadius: 3,
                    background: `linear-gradient(135deg, ${card.color}15 0%, ${card.color}25 100%)`,
                    border: `1px solid ${card.color}30`,
                    transition: 'all 0.3s ease'
                  }}
                >
                  <card.icon sx={{ 
                    color: card.color, 
                    fontSize: 32,
                    filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.1))'
                  }} />
                </Box>
              </Box>

            {/* Enhanced Progress bar for block rate */}
            {card.title === 'Block Rate' && stats && (
              <Box sx={{ mt: 'auto', pt: 2 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="caption" sx={{ color: '#64748b', fontWeight: 600 }}>
                    Threat Level
                  </Typography>
                  <Typography variant="caption" sx={{ color: card.color, fontWeight: 700 }}>
                    {parseFloat(blockRate) > 10 ? 'High' : parseFloat(blockRate) > 5 ? 'Medium' : 'Low'}
                  </Typography>
                </Box>
                <LinearProgress
                  variant="determinate"
                  value={Math.min(parseFloat(blockRate), 100)}
                  sx={{
                    height: 8,
                    borderRadius: 4,
                    backgroundColor: '#f1f5f9',
                    border: '1px solid #e2e8f0',
                    '& .MuiLinearProgress-bar': {
                      background: `linear-gradient(90deg, ${card.color} 0%, ${card.color}cc 100%)`,
                      borderRadius: 4,
                      boxShadow: `0 2px 8px ${card.color}40`
                    },
                  }}
                />
              </Box>
            )}
          </CardContent>
        </Card>
      ))}
    </Box>
  );
};

export default StatsCards;