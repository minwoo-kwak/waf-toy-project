import React from 'react';
import { Box, Typography } from '@mui/material';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import { WAFStats } from '../../types/waf';

interface AttackChartProps {
  stats: WAFStats | null;
}

const ATTACK_TYPE_COLORS = {
  'SQL Injection': '#e74c3c',
  'Cross-Site Scripting (XSS)': '#f39c12', 
  'Local File Inclusion (LFI)': '#9b59b6',
  'Remote File Inclusion (RFI)': '#8e44ad',
  'Command Injection': '#c0392b',
  'Path Traversal': '#d35400',
  'HTTP Protocol Violation': '#2980b9',
  'HTTP Protocol Anomaly': '#3498db',
  'Security Policy Violation': '#27ae60',
  'PHP Injection': '#e67e22',
  'Java Injection': '#34495e',
  'Session Fixation': '#16a085',
  'Remote Code Execution': '#8b0000',
  'Unknown': '#95a5a6'
};

const FALLBACK_COLORS = [
  '#ff6b6b', '#4ecdc4', '#45b7d1', '#f9ca24', '#6c5ce7',
  '#a0e7e5', '#ffeaa7', '#fab1a0', '#fd79a8', '#00b894'
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
    color: ATTACK_TYPE_COLORS[type as keyof typeof ATTACK_TYPE_COLORS] || FALLBACK_COLORS[Math.floor(Math.random() * FALLBACK_COLORS.length)]
  }));
  
  // Sort by value for better visualization
  chartData.sort((a, b) => b.value - a.value);

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
    <Box sx={{ width: '100%', height: 350 }}>
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={chartData}
            cx="50%"
            cy="45%"
            labelLine={false}
            label={renderCustomizedLabel}
            outerRadius={90}
            innerRadius={25}
            fill="#8884d8"
            dataKey="value"
            paddingAngle={2}
          >
            {chartData.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={entry.color} />
            ))}
          </Pie>
          <Tooltip content={<CustomTooltip />} />
          <Legend 
            verticalAlign="bottom" 
            height={60}
            formatter={(value: string, entry: any) => (
              <span style={{ 
                fontSize: '11px', 
                fontWeight: 600,
                color: '#374151'
              }}>
                {value} ({entry.payload?.value || 0})
              </span>
            )}
            wrapperStyle={{
              paddingTop: '20px',
              fontSize: '12px'
            }}
          />
        </PieChart>
      </ResponsiveContainer>
    </Box>
  );
};

export default AttackChart;