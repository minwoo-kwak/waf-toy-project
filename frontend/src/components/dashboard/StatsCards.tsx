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
    <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap' }}>
      {statsCards.map((card, index) => (
        <Box key={index} sx={{ flex: '1 1 250px', minWidth: '250px' }}>
          <Card
            sx={{
              height: '140px',
              display: 'flex',
              flexDirection: 'column',
              position: 'relative',
              overflow: 'hidden',
            }}
          >
            <CardContent sx={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
              <Box
                sx={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'flex-start',
                  mb: 1,
                }}
              >
                <Box>
                  <Typography
                    variant="body2"
                    color="textSecondary"
                    gutterBottom
                    sx={{ fontWeight: 500 }}
                  >
                    {card.title}
                  </Typography>
                  {loading ? (
                    <Skeleton width={80} height={32} />
                  ) : (
                    <Typography variant="h4" sx={{ fontWeight: 'bold', color: card.color }}>
                      {card.value}
                    </Typography>
                  )}
                </Box>
                <Box
                  sx={{
                    p: 1,
                    borderRadius: '8px',
                    backgroundColor: card.bgColor,
                  }}
                >
                  <card.icon sx={{ color: card.color, fontSize: 28 }} />
                </Box>
              </Box>

              {/* Progress bar for block rate */}
              {card.title === 'Block Rate' && stats && (
                <Box sx={{ mt: 'auto' }}>
                  <LinearProgress
                    variant="determinate"
                    value={Math.min(parseFloat(blockRate), 100)}
                    sx={{
                      height: 6,
                      borderRadius: 3,
                      backgroundColor: 'rgba(0,0,0,0.1)',
                      '& .MuiLinearProgress-bar': {
                        backgroundColor: card.color,
                        borderRadius: 3,
                      },
                    }}
                  />
                </Box>
              )}
            </CardContent>
          </Card>
        </Box>
      ))}
    </Box>
  );
};

export default StatsCards;