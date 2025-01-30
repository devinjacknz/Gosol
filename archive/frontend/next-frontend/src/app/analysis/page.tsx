'use client';

import { useEffect, useState } from 'react';
import {
  Container,
  Grid,
  Paper,
  Box,
  Tabs,
  Tab,
  Typography,
} from '@mui/material';
import { useRouter } from 'next/navigation';
import { useWallet } from '@/contexts/WalletContext';
import IndicatorSettings from '@/components/Analysis/IndicatorSettings';
import BacktestForm from '@/components/Analysis/BacktestForm';
import BacktestResults from '@/components/Analysis/BacktestResults';
import StrategyBuilder from '@/components/Analysis/StrategyBuilder';
import PerformanceMetrics from '@/components/Analysis/PerformanceMetrics';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`analysis-tabpanel-${index}`}
      aria-labelledby={`analysis-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

export default function AnalysisPage() {
  const { state: walletState } = useWallet();
  const router = useRouter();
  const [activeTab, setActiveTab] = useState(0);

  useEffect(() => {
    if (!walletState.isConnected) {
      router.push('/');
    }
  }, [walletState.isConnected, router]);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setActiveTab(newValue);
  };

  if (!walletState.isConnected) {
    return null;
  }

  return (
    <Container maxWidth={false} sx={{ mt: 4, mb: 4 }}>
      <Grid container spacing={2}>
        {/* Left Column - Strategy Builder */}
        <Grid item xs={12} md={3}>
          <Paper sx={{ p: 2, height: '100%' }}>
            <StrategyBuilder />
          </Paper>
        </Grid>

        {/* Middle Column - Chart and Analysis */}
        <Grid item xs={12} md={6}>
          <Grid container spacing={2}>
            {/* Tabs */}
            <Grid item xs={12}>
              <Paper sx={{ mb: 2 }}>
                <Tabs
                  value={activeTab}
                  onChange={handleTabChange}
                  aria-label="analysis tabs"
                >
                  <Tab label="Technical Analysis" />
                  <Tab label="Backtesting" />
                  <Tab label="Performance" />
                </Tabs>
              </Paper>
            </Grid>

            {/* Tab Panels */}
            <Grid item xs={12}>
              <TabPanel value={activeTab} index={0}>
                <Paper sx={{ p: 2 }}>
                  <IndicatorSettings />
                </Paper>
              </TabPanel>

              <TabPanel value={activeTab} index={1}>
                <Paper sx={{ p: 2 }}>
                  <BacktestForm />
                </Paper>
                <Box sx={{ mt: 2 }}>
                  <BacktestResults />
                </Box>
              </TabPanel>

              <TabPanel value={activeTab} index={2}>
                <Paper sx={{ p: 2 }}>
                  <PerformanceMetrics />
                </Paper>
              </TabPanel>
            </Grid>
          </Grid>
        </Grid>

        {/* Right Column - Strategy Performance */}
        <Grid item xs={12} md={3}>
          <Paper sx={{ p: 2, height: '100%' }}>
            <Typography variant="h6" gutterBottom>
              Strategy Performance
            </Typography>
            {/* Add strategy performance components */}
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
} 