import React from 'react';
import { Box, Typography } from '@mui/material';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import { WAFStats } from '../../types/waf';

interface AttackChartProps {
  stats: WAFStats | null;
}

const COLORS = [
  '#ff6b6b', // Red
  '#4ecdc4', // Teal
  '#45b7d1', // Blue
  '#f9ca24', // Yellow
  '#6c5ce7', // Purple
  '#a0e7e5', // Light Teal
  '#ffeaa7', // Light Yellow
  '#fab1a0', // Light Orange
  '#fd79a8', // Pink
  '#00b894', // Green
];

const AttackChart: React.FC<AttackChartProps> = ({ stats }) => {
  if (!stats || !stats.attacks_by_type) {
    return (
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: 300,
        }}
      >
        <Typography variant="body2" color="textSecondary">
          No attack data available
        </Typography>
      </Box>
    );
  }

  const attackTypes = Object.entries(stats.attacks_by_type);
  
  if (attackTypes.length === 0) {
    return (
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: 300,
        }}
      >
        <Typography variant="body2" color="textSecondary">
          No attacks detected
        </Typography>
      </Box>
    );
  }

  const chartData = attackTypes.map(([type, count]) => ({
    name: type,
    value: count,
  }));

  const renderCustomizedLabel = ({ cx, cy, midAngle, innerRadius, outerRadius, percent }: any) => {
    if (percent < 0.05) return null; // Hide labels for slices smaller than 5%
    
    const RADIAN = Math.PI / 180;
    const radius = innerRadius + (outerRadius - innerRadius) * 0.5;
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);

    return (
      <text
        x={x}
        y={y}
        fill="white"
        textAnchor={x > cx ? 'start' : 'end'}
        dominantBaseline="central"
        fontSize="12"
        fontWeight="bold"
      >
        {`${(percent * 100).toFixed(0)}%`}
      </text>
    );
  };

  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0];
      return (
        <Box
          sx={{
            bgcolor: 'background.paper',
            p: 1.5,
            border: 1,
            borderColor: 'divider',
            borderRadius: 1,
            boxShadow: 2,
          }}
        >
          <Typography variant="body2" sx={{ fontWeight: 'bold' }}>
            {data.payload.name}
          </Typography>
          <Typography variant="body2" color="textSecondary">
            Count: {data.value}
          </Typography>
        </Box>
      );
    }
    return null;
  };

  return (
    <Box sx={{ width: '100%', height: 300 }}>
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={chartData}
            cx="50%"
            cy="50%"
            labelLine={false}
            label={renderCustomizedLabel}
            outerRadius={80}
            fill="#8884d8"
            dataKey="value"
          >
            {chartData.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip content={<CustomTooltip />} />
          <Legend 
            verticalAlign="bottom" 
            height={36}
            formatter={(value: string) => (
              <span style={{ fontSize: '12px' }}>{value}</span>
            )}
          />
        </PieChart>
      </ResponsiveContainer>
    </Box>
  );
};

export default AttackChart;