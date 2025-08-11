import React from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Typography,
  Box,
  Skeleton,
  Tooltip,
} from '@mui/material';
import {
  Block as BlockIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Warning as WarningIcon,
  Info as InfoIcon,
} from '@mui/icons-material';
import { WAFLog } from '../../types/waf';
import { formatDistanceToNow } from 'date-fns';

interface LogsTableProps {
  logs: WAFLog[];
  loading: boolean;
}

const LogsTable: React.FC<LogsTableProps> = ({ logs, loading }) => {
  const getSeverityIcon = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical':
        return <ErrorIcon sx={{ fontSize: 16, color: 'error.main' }} />;
      case 'high':
        return <ErrorIcon sx={{ fontSize: 16, color: 'error.main' }} />;
      case 'medium':
        return <WarningIcon sx={{ fontSize: 16, color: 'warning.main' }} />;
      case 'low':
        return <InfoIcon sx={{ fontSize: 16, color: 'info.main' }} />;
      default:
        return <InfoIcon sx={{ fontSize: 16, color: 'grey.500' }} />;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical':
        return 'error';
      case 'high':
        return 'error';
      case 'medium':
        return 'warning';
      case 'low':
        return 'info';
      default:
        return 'default';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    try {
      return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
    } catch {
      return 'Unknown';
    }
  };

  if (loading) {
    return (
      <TableContainer component={Paper} variant="outlined">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Time</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>IP</TableCell>
              <TableCell>Attack Type</TableCell>
              <TableCell>Severity</TableCell>
              <TableCell>Message</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {[...Array(5)].map((_, index) => (
              <TableRow key={index}>
                <TableCell><Skeleton width={80} /></TableCell>
                <TableCell><Skeleton width={60} /></TableCell>
                <TableCell><Skeleton width={100} /></TableCell>
                <TableCell><Skeleton width={120} /></TableCell>
                <TableCell><Skeleton width={80} /></TableCell>
                <TableCell><Skeleton width={200} /></TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    );
  }

  if (!logs.length) {
    return (
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          py: 4,
        }}
      >
        <Typography variant="h6" color="textSecondary">
          No security events found
        </Typography>
        <Typography variant="body2" color="textSecondary">
          Your WAF is protecting your applications
        </Typography>
      </Box>
    );
  }

  return (
    <TableContainer component={Paper} variant="outlined" sx={{ maxHeight: 400 }}>
      <Table size="small" stickyHeader>
        <TableHead>
          <TableRow>
            <TableCell sx={{ fontWeight: 'bold' }}>Time</TableCell>
            <TableCell sx={{ fontWeight: 'bold' }}>Status</TableCell>
            <TableCell sx={{ fontWeight: 'bold' }}>IP Address</TableCell>
            <TableCell sx={{ fontWeight: 'bold' }}>Attack Type</TableCell>
            <TableCell sx={{ fontWeight: 'bold' }}>Severity</TableCell>
            <TableCell sx={{ fontWeight: 'bold' }}>Message</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {logs.map((log) => (
            <TableRow key={log.id} hover>
              <TableCell>
                <Typography variant="body2" sx={{ fontSize: '0.75rem' }}>
                  {formatTimestamp(log.timestamp)}
                </Typography>
              </TableCell>
              <TableCell>
                <Chip
                  icon={
                    log.blocked ? (
                      <BlockIcon sx={{ fontSize: 16 }} />
                    ) : (
                      <CheckCircleIcon sx={{ fontSize: 16 }} />
                    )
                  }
                  label={log.blocked ? 'Blocked' : 'Allowed'}
                  color={log.blocked ? 'error' : 'success'}
                  size="small"
                  variant="outlined"
                />
              </TableCell>
              <TableCell>
                <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                  {log.client_ip}
                </Typography>
              </TableCell>
              <TableCell>
                <Chip
                  label={log.attack_type || 'Unknown'}
                  size="small"
                  variant="outlined"
                  color={log.attack_type ? 'warning' : 'default'}
                />
              </TableCell>
              <TableCell>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                  {getSeverityIcon(log.severity)}
                  <Chip
                    label={log.severity || 'Unknown'}
                    size="small"
                    color={getSeverityColor(log.severity) as any}
                    variant="outlined"
                  />
                </Box>
              </TableCell>
              <TableCell>
                <Tooltip title={log.message} arrow>
                  <Typography
                    variant="body2"
                    sx={{
                      maxWidth: 200,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {log.message || 'No message'}
                  </Typography>
                </Tooltip>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default LogsTable;