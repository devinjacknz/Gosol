'use client';

import { useState, useEffect } from 'react';
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
  Stepper,
  Step,
  StepLabel,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
} from '@mui/material';
import {
  QrCode2,
  Smartphone,
  Security,
  Check,
} from '@mui/icons-material';
import Image from 'next/image';

export default function Setup2FA() {
  const router = useRouter();
  const [activeStep, setActiveStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [qrCode, setQrCode] = useState<string>('');
  const [secret, setSecret] = useState<string>('');
  const [verificationCode, setVerificationCode] = useState('');
  const [backupCodes, setBackupCodes] = useState<string[]>([]);

  useEffect(() => {
    generateSecret();
  }, []);

  const generateSecret = async () => {
    try {
      const response = await fetch('/api/auth/2fa/generate');
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message);
      }

      setQrCode(data.qrCode);
      setSecret(data.secret);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate 2FA secret');
    }
  };

  const verifyAndEnable = async () => {
    setLoading(true);
    setError('');

    try {
      const response = await fetch('/api/auth/2fa/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code: verificationCode, secret }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message);
      }

      setBackupCodes(data.backupCodes);
      setActiveStep(2);
      setSuccess('Two-factor authentication enabled successfully');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to verify code');
    } finally {
      setLoading(false);
    }
  };

  const handleFinish = () => {
    router.push('/profile');
  };

  const steps = [
    'Scan QR Code',
    'Verify Code',
    'Save Backup Codes',
  ];

  return (
    <Container maxWidth="sm">
      <Box sx={{ mt: 4, mb: 4 }}>
        <Paper elevation={3} sx={{ p: 4 }}>
          <Typography component="h1" variant="h5" align="center" gutterBottom>
            Set Up Two-Factor Authentication
          </Typography>

          <Stepper activeStep={activeStep} sx={{ mb: 4 }}>
            {steps.map((label) => (
              <Step key={label}>
                <StepLabel>{label}</StepLabel>
              </Step>
            ))}
          </Stepper>

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          {success && (
            <Alert severity="success" sx={{ mb: 2 }}>
              {success}
            </Alert>
          )}

          {activeStep === 0 && (
            <Box sx={{ textAlign: 'center' }}>
              <Typography variant="h6" gutterBottom>
                Scan this QR code with your authenticator app
              </Typography>
              
              {qrCode && (
                <Box sx={{ my: 3 }}>
                  <Image
                    src={qrCode}
                    alt="2FA QR Code"
                    width={200}
                    height={200}
                  />
                </Box>
              )}

              <Typography variant="body2" color="text.secondary" gutterBottom>
                Or enter this code manually:
              </Typography>
              <Typography
                variant="body1"
                sx={{ fontFamily: 'monospace', mb: 3 }}
              >
                {secret}
              </Typography>

              <List>
                <ListItem>
                  <ListItemIcon>
                    <Smartphone />
                  </ListItemIcon>
                  <ListItemText primary="1. Install an authenticator app" />
                </ListItem>
                <ListItem>
                  <ListItemIcon>
                    <QrCode2 />
                  </ListItemIcon>
                  <ListItemText primary="2. Scan the QR code or enter the secret manually" />
                </ListItem>
                <ListItem>
                  <ListItemIcon>
                    <Security />
                  </ListItemIcon>
                  <ListItemText primary="3. Enter the verification code from the app" />
                </ListItem>
              </List>

              <Button
                variant="contained"
                onClick={() => setActiveStep(1)}
                sx={{ mt: 2 }}
              >
                Next
              </Button>
            </Box>
          )}

          {activeStep === 1 && (
            <Box>
              <Typography variant="h6" gutterBottom>
                Enter Verification Code
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                Enter the 6-digit code from your authenticator app
              </Typography>

              <TextField
                fullWidth
                label="Verification Code"
                value={verificationCode}
                onChange={(e) => setVerificationCode(e.target.value)}
                sx={{ mb: 3 }}
                inputProps={{
                  maxLength: 6,
                  pattern: '[0-9]*',
                }}
              />

              <Box sx={{ display: 'flex', gap: 2 }}>
                <Button
                  variant="outlined"
                  onClick={() => setActiveStep(0)}
                >
                  Back
                </Button>
                <Button
                  variant="contained"
                  onClick={verifyAndEnable}
                  disabled={verificationCode.length !== 6 || loading}
                >
                  {loading ? <CircularProgress size={24} /> : 'Verify and Enable'}
                </Button>
              </Box>
            </Box>
          )}

          {activeStep === 2 && (
            <Box>
              <Typography variant="h6" gutterBottom>
                Save Your Backup Codes
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                Store these backup codes in a safe place. You can use them to access your account if you lose your authenticator device.
              </Typography>

              <Paper
                variant="outlined"
                sx={{
                  p: 2,
                  mb: 3,
                  backgroundColor: 'grey.100',
                  fontFamily: 'monospace',
                }}
              >
                {backupCodes.map((code, index) => (
                  <Typography key={index} sx={{ mb: 1 }}>
                    {code}
                  </Typography>
                ))}
              </Paper>

              <Button
                variant="contained"
                onClick={handleFinish}
                startIcon={<Check />}
              >
                Finish Setup
              </Button>
            </Box>
          )}
        </Paper>
      </Box>
    </Container>
  );
} 