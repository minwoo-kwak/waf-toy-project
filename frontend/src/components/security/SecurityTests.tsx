import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Grid,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  MenuItem,
  LinearProgress,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import {
  PlayArrow as PlayIcon,
  Security as SecurityIcon,
  ExpandMore as ExpandMoreIcon,
  CheckCircle as CheckCircleIcon,
  Cancel as CancelIcon,
} from '@mui/icons-material';
import { securityAPI } from '../../services/api';
import { SecurityTest, SecurityTestRequest } from '../../types/waf';

const SecurityTests: React.FC = () => {
  const [testTypes, setTestTypes] = useState<any[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedTestType, setSelectedTestType] = useState('');
  const [customPayloads, setCustomPayloads] = useState('');
  const [running, setRunning] = useState(false);
  const [testResult, setTestResult] = useState<SecurityTest | null>(null);
  const [testSummary, setTestSummary] = useState<any>(null);
  const [quickTestResults, setQuickTestResults] = useState<any[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadTestTypes();
  }, []);

  const loadTestTypes = async () => {
    try {
      const response = await securityAPI.getTestTypes();
      setTestTypes(response.test_types);
    } catch (error: any) {
      console.error('Failed to load test types:', error);
      setError('Failed to load test types');
    }
  };

  const runQuickTests = async () => {
    setRunning(true);
    setError(null);
    
    try {
      const response = await securityAPI.getQuickTests();
      setQuickTestResults(response.quick_tests);
    } catch (error: any) {
      console.error('Quick tests failed:', error);
      setError('Quick tests failed');
    } finally {
      setRunning(false);
    }
  };

  const runCustomTest = async () => {
    if (!selectedTestType) {
      setError('Please select a test type');
      return;
    }

    setRunning(true);
    setError(null);

    try {
      const payloads = customPayloads
        ? customPayloads.split('\n').filter(p => p.trim())
        : undefined;

      const request: SecurityTestRequest = {
        test_type: selectedTestType as any,
        payloads,
      };

      const response = await securityAPI.runSecurityTest(request);
      setTestResult(response.test);
      setTestSummary(response.summary);
      setDialogOpen(false);
    } catch (error: any) {
      console.error('Custom test failed:', error);
      setError('Custom test failed');
    } finally {
      setRunning(false);
    }
  };

  const getEffectivenessColor = (effectiveness: string) => {
    switch (effectiveness.toLowerCase()) {
      case 'excellent':
        return 'success';
      case 'good':
        return 'primary';
      case 'fair':
        return 'warning';
      case 'poor':
        return 'error';
      case 'critical':
        return 'error';
      default:
        return 'default';
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <SecurityIcon />
        Security Testing Suite
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Quick Tests */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Quick Security Assessment
          </Typography>
          <Typography variant="body2" color="textSecondary" paragraph>
            Run a comprehensive security test with common attack patterns
          </Typography>
          
          <Button
            variant="contained"
            startIcon={<PlayIcon />}
            onClick={runQuickTests}
            disabled={running}
            sx={{ mb: 2 }}
          >
            {running ? 'Running Tests...' : 'Run Quick Tests'}
          </Button>

          {running && <LinearProgress sx={{ mb: 2 }} />}

          {quickTestResults.length > 0 && (
            <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
              {quickTestResults.map((result, index) => (
                <Box key={index} sx={{ flex: '1 1 300px', minWidth: '300px' }}>
                  <Card variant="outlined">
                    <CardContent>
                      <Typography variant="h6" gutterBottom>
                        {result.test_type.replace('_', ' ').toUpperCase()}
                      </Typography>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                        <Typography variant="body2">Total Tests:</Typography>
                        <Typography variant="body2" sx={{ fontWeight: 'bold' }}>
                          {result.total_tests}
                        </Typography>
                      </Box>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                        <Typography variant="body2">Blocked:</Typography>
                        <Typography variant="body2" sx={{ fontWeight: 'bold', color: 'error.main' }}>
                          {result.blocked_tests}
                        </Typography>
                      </Box>
                      <Chip
                        label={result.effectiveness}
                        color={getEffectivenessColor(result.effectiveness) as any}
                        size="small"
                        sx={{ mt: 1 }}
                      />
                    </CardContent>
                  </Card>
                </Box>
              ))}
            </Box>
          )}
        </CardContent>
      </Card>

      {/* Custom Tests */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Custom Security Tests
          </Typography>
          <Typography variant="body2" color="textSecondary" paragraph>
            Run specific security tests with custom payloads
          </Typography>

          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
            {testTypes.map((testType) => (
              <Box key={testType.id} sx={{ flex: '1 1 300px', minWidth: '300px' }}>
                <Card
                  variant="outlined"
                  sx={{
                    cursor: 'pointer',
                    '&:hover': { boxShadow: 2 },
                  }}
                  onClick={() => {
                    setSelectedTestType(testType.id);
                    setDialogOpen(true);
                  }}
                >
                  <CardContent>
                    <Typography variant="h6" gutterBottom>
                      {testType.name}
                    </Typography>
                    <Typography variant="body2" color="textSecondary" paragraph>
                      {testType.description}
                    </Typography>
                    <Chip
                      label={testType.severity}
                      color={
                        testType.severity === 'CRITICAL'
                          ? 'error'
                          : testType.severity === 'HIGH'
                          ? 'error'
                          : testType.severity === 'MEDIUM'
                          ? 'warning'
                          : 'info'
                      }
                      size="small"
                    />
                  </CardContent>
                </Card>
              </Box>
            ))}
          </Box>
        </CardContent>
      </Card>

      {/* Test Results */}
      {testResult && testSummary && (
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Test Results: {testResult.name}
            </Typography>

            {/* Summary */}
            <Box sx={{ display: 'flex', gap: 2, mb: 3, flexWrap: 'wrap' }}>
              <Box sx={{ flex: '1 1 200px', minWidth: '200px' }}>
                <Card variant="outlined">
                  <CardContent>
                    <Typography variant="body2" color="textSecondary">
                      Total Tests
                    </Typography>
                    <Typography variant="h5">
                      {testSummary.total_tests}
                    </Typography>
                  </CardContent>
                </Card>
              </Box>
              <Box sx={{ flex: '1 1 200px', minWidth: '200px' }}>
                <Card variant="outlined">
                  <CardContent>
                    <Typography variant="body2" color="textSecondary">
                      Blocked
                    </Typography>
                    <Typography variant="h5" color="error.main">
                      {testSummary.blocked_tests}
                    </Typography>
                  </CardContent>
                </Card>
              </Box>
              <Box sx={{ flex: '1 1 200px', minWidth: '200px' }}>
                <Card variant="outlined">
                  <CardContent>
                    <Typography variant="body2" color="textSecondary">
                      Block Rate
                    </Typography>
                    <Typography variant="h5">
                      {testSummary.block_rate.toFixed(1)}%
                    </Typography>
                  </CardContent>
                </Card>
              </Box>
              <Box sx={{ flex: '1 1 200px', minWidth: '200px' }}>
                <Card variant="outlined">
                  <CardContent>
                    <Typography variant="body2" color="textSecondary">
                      Effectiveness
                    </Typography>
                    <Chip
                      label={testSummary.effectiveness}
                      color={getEffectivenessColor(testSummary.effectiveness) as any}
                    />
                  </CardContent>
                </Card>
              </Box>
            </Box>

            {/* Detailed Results */}
            <Accordion>
              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <Typography variant="h6">Detailed Results</Typography>
              </AccordionSummary>
              <AccordionDetails>
                <TableContainer component={Paper} variant="outlined">
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Payload</TableCell>
                        <TableCell>Status</TableCell>
                        <TableCell>HTTP Code</TableCell>
                        <TableCell>Response</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {testResult.results.map((result, index) => (
                        <TableRow key={index}>
                          <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.8rem' }}>
                            {result.payload}
                          </TableCell>
                          <TableCell>
                            <Chip
                              icon={
                                result.blocked ? (
                                  <CancelIcon sx={{ fontSize: 16 }} />
                                ) : (
                                  <CheckCircleIcon sx={{ fontSize: 16 }} />
                                )
                              }
                              label={result.blocked ? 'Blocked' : 'Allowed'}
                              color={result.blocked ? 'error' : 'success'}
                              size="small"
                              variant="outlined"
                            />
                          </TableCell>
                          <TableCell>{result.status_code}</TableCell>
                          <TableCell>
                            <Typography
                              variant="body2"
                              sx={{
                                maxWidth: 200,
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                                whiteSpace: 'nowrap',
                                fontFamily: 'monospace',
                                fontSize: '0.7rem',
                              }}
                            >
                              {result.response}
                            </Typography>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              </AccordionDetails>
            </Accordion>
          </CardContent>
        </Card>
      )}

      {/* Custom Test Dialog */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>Custom Security Test</DialogTitle>
        <DialogContent>
          <TextField
            select
            fullWidth
            label="Test Type"
            value={selectedTestType}
            onChange={(e) => setSelectedTestType(e.target.value)}
            sx={{ mb: 2, mt: 1 }}
          >
            {testTypes.map((type) => (
              <MenuItem key={type.id} value={type.id}>
                {type.name}
              </MenuItem>
            ))}
          </TextField>

          <TextField
            fullWidth
            multiline
            rows={8}
            label="Custom Payloads (one per line)"
            placeholder="Enter custom attack payloads, one per line..."
            value={customPayloads}
            onChange={(e) => setCustomPayloads(e.target.value)}
            helperText="Leave empty to use default payloads for the selected test type"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={runCustomTest}
            variant="contained"
            disabled={running}
            startIcon={running ? <LinearProgress /> : <PlayIcon />}
          >
            {running ? 'Running...' : 'Run Test'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default SecurityTests;