import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Alert,
  CircularProgress,
  Tooltip,
  Divider,
  Container,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  CloudQueue as ProxyIcon,
  Launch as LaunchIcon,
  Security as SecurityIcon,
  Refresh as RefreshIcon,
} from '@mui/icons-material';
import { proxyAPI } from '../../services/api';
import { ProxyTarget, ProxyTargetRequest } from '../../types/waf';

const ProxyManagement: React.FC = () => {
  const [proxyTargets, setProxyTargets] = useState<ProxyTarget[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [open, setOpen] = useState(false);
  const [editingProxy, setEditingProxy] = useState<ProxyTarget | null>(null);
  const [formData, setFormData] = useState<ProxyTargetRequest>({
    name: '',
    origin_url: '',
    proxy_domain: '',
    status: 'active',
  });
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    loadProxyTargets();
  }, []);

  const loadProxyTargets = async () => {
    try {
      setLoading(true);
      const response = await proxyAPI.getProxyTargets();
      setProxyTargets(response.proxy_targets);
      setError(null);
    } catch (err: any) {
      console.error('Failed to load proxy targets:', err);
      setError('Failed to load proxy targets. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (proxy?: ProxyTarget) => {
    if (proxy) {
      setEditingProxy(proxy);
      setFormData({
        name: proxy.name,
        origin_url: proxy.origin_url,
        proxy_domain: proxy.proxy_domain,
        status: proxy.status,
      });
    } else {
      setEditingProxy(null);
      setFormData({
        name: '',
        origin_url: '',
        proxy_domain: '',
        status: 'active',
      });
    }
    setFormErrors({});
    setOpen(true);
  };

  const handleCloseDialog = () => {
    setOpen(false);
    setEditingProxy(null);
    setFormData({
      name: '',
      origin_url: '',
      proxy_domain: '',
      status: 'active',
    });
    setFormErrors({});
  };

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};

    if (!formData.name.trim()) {
      errors.name = 'Name is required';
    }

    if (!formData.origin_url.trim()) {
      errors.origin_url = 'Origin URL is required';
    } else {
      try {
        new URL(formData.origin_url);
      } catch {
        errors.origin_url = 'Invalid URL format';
      }
    }

    if (!formData.proxy_domain.trim()) {
      errors.proxy_domain = 'Proxy domain is required';
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async () => {
    if (!validateForm()) return;

    try {
      setSubmitting(true);
      if (editingProxy) {
        await proxyAPI.updateProxyTarget(editingProxy.id, formData);
      } else {
        await proxyAPI.createProxyTarget(formData);
      }
      await loadProxyTargets();
      handleCloseDialog();
    } catch (err: any) {
      console.error('Failed to save proxy target:', err);
      setError(`Failed to ${editingProxy ? 'update' : 'create'} proxy target. Please try again.`);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this proxy target?')) {
      return;
    }

    try {
      await proxyAPI.deleteProxyTarget(id);
      await loadProxyTargets();
    } catch (err: any) {
      console.error('Failed to delete proxy target:', err);
      setError('Failed to delete proxy target. Please try again.');
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'success';
      case 'inactive':
        return 'default';
      default:
        return 'default';
    }
  };

  const generateSuggestedDomain = () => {
    if (formData.origin_url) {
      try {
        const url = new URL(formData.origin_url);
        const hostname = url.hostname.replace(/[^a-zA-Z0-9]/g, '-');
        const suggested = `${hostname}-protected.waftest.p-e.kr:31268`;
        setFormData({ ...formData, proxy_domain: suggested });
      } catch {
        // Invalid URL이거나 비어있는 경우 기본 템플릿 제공
        const suggested = `myapp-protected.waftest.p-e.kr:31268`;
        setFormData({ ...formData, proxy_domain: suggested });
      }
    } else {
      // Origin URL이 없는 경우에도 기본 템플릿 제공
      const suggested = `myapp-protected.waftest.p-e.kr:31268`;
      setFormData({ ...formData, proxy_domain: suggested });
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress size={48} />
      </Box>
    );
  }

  return (
    <Container maxWidth="xl" sx={{ py: 4 }}>
      {/* Header */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
        <Box sx={{ display: 'flex', alignItems: 'center' }}>
          <ProxyIcon sx={{ mr: 2, fontSize: 32, color: 'primary.main' }} />
          <Typography variant="h4" sx={{ fontWeight: 700, color: '#1e293b' }}>
            Proxy Management
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', gap: 2 }}>
          <Button
            variant="outlined"
            startIcon={<RefreshIcon />}
            onClick={loadProxyTargets}
            disabled={loading}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => handleOpenDialog()}
            sx={{
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              '&:hover': {
                background: 'linear-gradient(135deg, #5a67d8 0%, #6b46c1 100%)',
              },
            }}
          >
            Add Proxy Target
          </Button>
        </Box>
      </Box>

      {/* Error Alert */}
      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {/* Info Card */}
      <Card sx={{ mb: 3, background: 'linear-gradient(135deg, #e0f2fe 0%, #f3e5f5 100%)' }}>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
            <SecurityIcon sx={{ mr: 2, color: 'primary.main' }} />
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              How WAF Proxy Works
            </Typography>
          </Box>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            WAF Proxy protects your servers by routing traffic through our ModSecurity-enabled gateway.
            When you add a proxy target, we create a protected domain that filters malicious requests before
            forwarding them to your origin server.
          </Typography>
          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
            <Chip
              label="🛡️ ModSecurity Protection"
              size="small"
              sx={{ bgcolor: 'rgba(76, 175, 80, 0.1)', color: 'success.main' }}
            />
            <Chip
              label="🚀 OWASP CRS Rules"
              size="small"
              sx={{ bgcolor: 'rgba(33, 150, 243, 0.1)', color: 'primary.main' }}
            />
            <Chip
              label="📊 Real-time Monitoring"
              size="small"
              sx={{ bgcolor: 'rgba(156, 39, 176, 0.1)', color: 'secondary.main' }}
            />
          </Box>
        </CardContent>
      </Card>

      {/* Proxy Targets Table */}
      <Card elevation={0} sx={{ borderRadius: 3, border: '1px solid #e2e8f0' }}>
        <CardContent>
          <Typography variant="h6" sx={{ mb: 3, fontWeight: 600 }}>
            Configured Proxy Targets ({proxyTargets.length})
          </Typography>

          {proxyTargets.length === 0 ? (
            <Box
              sx={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                py: 8,
                color: 'text.secondary',
              }}
            >
              <ProxyIcon sx={{ fontSize: 64, mb: 2, opacity: 0.3 }} />
              <Typography variant="h6" sx={{ mb: 1, fontWeight: 600 }}>
                No proxy targets configured
              </Typography>
              <Typography variant="body2" sx={{ mb: 3, textAlign: 'center', maxWidth: 400 }}>
                Get started by adding your first proxy target. We'll create a protected domain
                that routes traffic through our WAF before reaching your origin server.
              </Typography>
              <Button
                variant="contained"
                startIcon={<AddIcon />}
                onClick={() => handleOpenDialog()}
                sx={{
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  '&:hover': {
                    background: 'linear-gradient(135deg, #5a67d8 0%, #6b46c1 100%)',
                  },
                }}
              >
                Add Your First Proxy Target
              </Button>
            </Box>
          ) : (
            <TableContainer component={Paper} elevation={0}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell sx={{ fontWeight: 600 }}>Name</TableCell>
                    <TableCell sx={{ fontWeight: 600 }}>Origin URL</TableCell>
                    <TableCell sx={{ fontWeight: 600 }}>Proxy Domain</TableCell>
                    <TableCell sx={{ fontWeight: 600 }}>Status</TableCell>
                    <TableCell sx={{ fontWeight: 600 }}>Created</TableCell>
                    <TableCell sx={{ fontWeight: 600 }}>Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {proxyTargets.map((proxy) => (
                    <TableRow key={proxy.id}>
                      <TableCell>
                        <Typography variant="body2" sx={{ fontWeight: 600 }}>
                          {proxy.name}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                            {proxy.origin_url}
                          </Typography>
                          <Tooltip title="Open origin URL">
                            <IconButton
                              size="small"
                              onClick={() => window.open(proxy.origin_url, '_blank')}
                            >
                              <LaunchIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                            {proxy.proxy_domain}
                          </Typography>
                          <Tooltip title="Open protected URL">
                            <IconButton
                              size="small"
                              onClick={() => window.open(`https://${proxy.proxy_domain}`, '_blank')}
                            >
                              <LaunchIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={proxy.status}
                          color={getStatusColor(proxy.status) as any}
                          size="small"
                          sx={{ fontWeight: 600 }}
                        />
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color="text.secondary">
                          {new Date(proxy.created_at).toLocaleDateString()}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', gap: 1 }}>
                          <Tooltip title="Edit proxy target">
                            <IconButton
                              size="small"
                              onClick={() => handleOpenDialog(proxy)}
                              color="primary"
                            >
                              <EditIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title="Delete proxy target">
                            <IconButton
                              size="small"
                              onClick={() => handleDelete(proxy.id)}
                              color="error"
                            >
                              <DeleteIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                        </Box>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </CardContent>
      </Card>

      {/* Add/Edit Dialog */}
      <Dialog open={open} onClose={handleCloseDialog} maxWidth="md" fullWidth>
        <DialogTitle>
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <ProxyIcon sx={{ mr: 2, color: 'primary.main' }} />
            {editingProxy ? 'Edit Proxy Target' : 'Add New Proxy Target'}
          </Box>
        </DialogTitle>
        <DialogContent>
          <Box sx={{ mt: 2, display: 'flex', flexDirection: 'column', gap: 3 }}>
            <TextField
              fullWidth
              label="Name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              error={!!formErrors.name}
              helperText={formErrors.name || 'A descriptive name for this proxy target'}
              placeholder="My Web Application"
            />

            <Box>
              <TextField
                fullWidth
                label="Origin URL"
                value={formData.origin_url}
                onChange={(e) => setFormData({ ...formData, origin_url: e.target.value })}
                error={!!formErrors.origin_url}
                helperText={formErrors.origin_url || 'The URL of your server to be protected'}
                placeholder="http://192.168.1.100:30081"
              />
              <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                💡 For DVWA testing, use: http://&lt;cluster-ip&gt;:30081
              </Typography>
            </Box>

            <Box>
              <TextField
                fullWidth
                label="Proxy Domain"
                value={formData.proxy_domain}
                onChange={(e) => setFormData({ ...formData, proxy_domain: e.target.value })}
                error={!!formErrors.proxy_domain}
                helperText={formErrors.proxy_domain || 'The protected domain that will route to your origin'}
                placeholder="myapp-protected.waftest.p-e.kr:31268"
                InputProps={{
                  endAdornment: (
                    <Button
                      size="small"
                      onClick={generateSuggestedDomain}
                    >
                      Suggest
                    </Button>
                  ),
                }}
              />
              <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                🔐 This domain will be protected by ModSecurity and OWASP CRS rules
              </Typography>
            </Box>

            <FormControl fullWidth>
              <InputLabel>Status</InputLabel>
              <Select
                value={formData.status}
                label="Status"
                onChange={(e) => setFormData({ ...formData, status: e.target.value as 'active' | 'inactive' })}
              >
                <MenuItem value="active">Active</MenuItem>
                <MenuItem value="inactive">Inactive</MenuItem>
              </Select>
            </FormControl>

            {!editingProxy && (
              <Alert severity="info">
                <Typography variant="body2">
                  <strong>Next Steps:</strong> After creating the proxy target, you'll need to:
                </Typography>
                <Box component="ul" sx={{ mt: 1, mb: 0, pl: 2 }}>
                  <li>Configure DNS to point your proxy domain to our servers</li>
                  <li>Test the protected endpoint with security payloads</li>
                  <li>Monitor security logs in the Dashboard</li>
                </Box>
              </Alert>
            )}
          </Box>
        </DialogContent>
        <DialogActions sx={{ p: 3, pt: 1 }}>
          <Button onClick={handleCloseDialog} disabled={submitting}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            variant="contained"
            disabled={submitting}
            sx={{
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              '&:hover': {
                background: 'linear-gradient(135deg, #5a67d8 0%, #6b46c1 100%)',
              },
            }}
          >
            {submitting ? (
              <CircularProgress size={20} color="inherit" />
            ) : (
              editingProxy ? 'Update Proxy Target' : 'Create Proxy Target'
            )}
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default ProxyManagement;