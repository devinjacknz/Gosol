'use client';

import { useState, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/navigation';
import {
  Container,
  Paper,
  Box,
  Typography,
  TextField,
  Button,
  Alert,
  CircularProgress,
  Grid,
  Switch,
  FormControlLabel,
  Divider,
  Avatar,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
} from '@mui/material';
import { Security, Person, Notifications } from '@mui/icons-material';

interface ProfileData {
  name: string;
  email: string;
  phone?: string;
  twoFactorEnabled: boolean;
  emailNotifications: boolean;
}

export default function Profile() {
  const router = useRouter();
  const { data: session } = useSession();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [showDisableDialog, setShowDisableDialog] = useState(false);
  const [profileData, setProfileData] = useState<ProfileData>({
    name: session?.user?.name || '',
    email: session?.user?.email || '',
    phone: '',
    twoFactorEnabled: false,
    emailNotifications: true,
  });

  useEffect(() => {
    fetchProfile();
  }, []);

  const fetchProfile = async () => {
    try {
      const response = await fetch('/api/user/profile');
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message);
      }

      setProfileData(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch profile');
    }
  };

  const handleProfileUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setSuccess('');

    try {
      const response = await fetch('/api/user/profile', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(profileData),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Failed to update profile');
      }

      setSuccess('Profile updated successfully');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update profile');
    } finally {
      setLoading(false);
    }
  };

  const handle2FAToggle = async () => {
    if (!profileData.twoFactorEnabled) {
      router.push('/profile/2fa/setup');
    } else {
      setShowDisableDialog(true);
    }
  };

  const handleDisable2FA = async () => {
    setShowDisableDialog(false);
    setLoading(true);
    setError('');

    try {
      const response = await fetch('/api/auth/2fa/disable', {
        method: 'POST',
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message);
      }

      setProfileData(prev => ({
        ...prev,
        twoFactorEnabled: false,
      }));
      setSuccess('Two-factor authentication disabled successfully');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to disable 2FA');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (field: keyof ProfileData) => (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const value = e.target.type === 'checkbox' ? e.target.checked : e.target.value;
    setProfileData(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  return (
    <Container maxWidth="md">
      <Box sx={{ mt: 4, mb: 4 }}>
        <Grid container spacing={3}>
          {/* Profile Information */}
          <Grid item xs={12}>
            <Paper elevation={3} sx={{ p: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
                <Person sx={{ mr: 1 }} />
                <Typography variant="h6">Profile Information</Typography>
              </Box>

              {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
              {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}

              <Box component="form" onSubmit={handleProfileUpdate}>
                <Grid container spacing={2}>
                  <Grid item xs={12} sm={4} sx={{ textAlign: 'center' }}>
                    <Avatar
                      sx={{ width: 100, height: 100, mx: 'auto', mb: 2 }}
                      src={session?.user?.image || undefined}
                    />
                    <Button variant="outlined" size="small">
                      Change Avatar
                    </Button>
                  </Grid>
                  <Grid item xs={12} sm={8}>
                    <TextField
                      margin="normal"
                      fullWidth
                      label="Name"
                      name="name"
                      value={profileData.name}
                      onChange={handleChange('name')}
                      disabled={loading}
                    />
                    <TextField
                      margin="normal"
                      fullWidth
                      label="Email"
                      name="email"
                      value={profileData.email}
                      onChange={handleChange('email')}
                      disabled={loading}
                    />
                    <TextField
                      margin="normal"
                      fullWidth
                      label="Phone"
                      name="phone"
                      value={profileData.phone}
                      onChange={handleChange('phone')}
                      disabled={loading}
                    />
                  </Grid>
                </Grid>

                <Button
                  type="submit"
                  variant="contained"
                  sx={{ mt: 3 }}
                  disabled={loading}
                >
                  {loading ? <CircularProgress size={24} /> : 'Save Changes'}
                </Button>
              </Box>
            </Paper>
          </Grid>

          {/* Security Settings */}
          <Grid item xs={12}>
            <Paper elevation={3} sx={{ p: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
                <Security sx={{ mr: 1 }} />
                <Typography variant="h6">Security Settings</Typography>
              </Box>

              <FormControlLabel
                control={
                  <Switch
                    checked={profileData.twoFactorEnabled}
                    onChange={handle2FAToggle}
                    disabled={loading}
                  />
                }
                label={
                  <Box>
                    <Typography>Two-Factor Authentication</Typography>
                    <Typography variant="body2" color="text.secondary">
                      {profileData.twoFactorEnabled
                        ? 'Your account is protected with two-factor authentication'
                        : 'Add an extra layer of security to your account'}
                    </Typography>
                  </Box>
                }
              />

              <Divider sx={{ my: 2 }} />

              <Button
                variant="outlined"
                color="primary"
                onClick={() => router.push('/auth/reset-password')}
              >
                Change Password
              </Button>
            </Paper>
          </Grid>

          {/* Notification Settings */}
          <Grid item xs={12}>
            <Paper elevation={3} sx={{ p: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
                <Notifications sx={{ mr: 1 }} />
                <Typography variant="h6">Notification Settings</Typography>
              </Box>

              <FormControlLabel
                control={
                  <Switch
                    checked={profileData.emailNotifications}
                    onChange={handleChange('emailNotifications')}
                    disabled={loading}
                  />
                }
                label="Email Notifications"
              />

              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Receive email notifications about account activity and trading updates.
              </Typography>
            </Paper>
          </Grid>
        </Grid>
      </Box>

      {/* Disable 2FA Confirmation Dialog */}
      <Dialog
        open={showDisableDialog}
        onClose={() => setShowDisableDialog(false)}
      >
        <DialogTitle>Disable Two-Factor Authentication?</DialogTitle>
        <DialogContent>
          <DialogContentText>
            This will remove the extra layer of security from your account. Are you sure you want to continue?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowDisableDialog(false)}>Cancel</Button>
          <Button onClick={handleDisable2FA} color="error">
            Disable 2FA
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
} 