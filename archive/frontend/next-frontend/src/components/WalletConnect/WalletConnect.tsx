'use client';

import { useState } from 'react';
import {
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Typography,
  Box,
  IconButton,
  Menu,
  MenuItem,
  Tooltip,
} from '@mui/material';
import {
  AccountBalanceWallet,
  ExpandMore,
  ContentCopy,
  Logout,
  SwapHoriz,
} from '@mui/icons-material';
import Image from 'next/image';
import { useWallet } from '@/contexts/WalletContext';
import { SUPPORTED_WALLETS, type SupportedWallets } from '@/config/wallet';

export default function WalletConnect() {
  const { state, connect, disconnect } = useWallet();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

  const handleConnect = async (walletType: SupportedWallets) => {
    await connect(walletType);
    setDialogOpen(false);
  };

  const handleCopyAddress = () => {
    if (state.account?.address) {
      navigator.clipboard.writeText(state.account.address);
    }
  };

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleDisconnect = () => {
    disconnect();
    handleMenuClose();
  };

  const formatAddress = (address: string) => {
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  };

  if (!state.isConnected) {
    return (
      <>
        <Button
          variant="contained"
          startIcon={<AccountBalanceWallet />}
          onClick={() => setDialogOpen(true)}
          disabled={state.isConnecting}
        >
          {state.isConnecting ? 'Connecting...' : 'Connect Wallet'}
        </Button>

        <Dialog
          open={dialogOpen}
          onClose={() => setDialogOpen(false)}
          maxWidth="xs"
          fullWidth
        >
          <DialogTitle>Connect Wallet</DialogTitle>
          <DialogContent>
            <List>
              {Object.entries(SUPPORTED_WALLETS).map(([key, wallet]) => (
                <ListItem key={key} disablePadding>
                  <ListItemButton
                    onClick={() => handleConnect(key as SupportedWallets)}
                  >
                    <ListItemIcon>
                      <Image
                        src={wallet.icon}
                        alt={wallet.name}
                        width={32}
                        height={32}
                      />
                    </ListItemIcon>
                    <ListItemText primary={wallet.name} />
                  </ListItemButton>
                </ListItem>
              ))}
            </List>
          </DialogContent>
        </Dialog>
      </>
    );
  }

  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
      <Tooltip title={`Balance: ${state.balance} SUI`}>
        <Typography variant="body2" color="text.secondary">
          {state.balance} SUI
        </Typography>
      </Tooltip>

      <Button
        variant="outlined"
        onClick={handleMenuOpen}
        endIcon={<ExpandMore />}
      >
        {state.account?.address && formatAddress(state.account.address)}
      </Button>

      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
      >
        <MenuItem onClick={handleCopyAddress}>
          <ListItemIcon>
            <ContentCopy fontSize="small" />
          </ListItemIcon>
          <ListItemText>Copy Address</ListItemText>
        </MenuItem>
        <MenuItem onClick={() => setDialogOpen(true)}>
          <ListItemIcon>
            <SwapHoriz fontSize="small" />
          </ListItemIcon>
          <ListItemText>Switch Wallet</ListItemText>
        </MenuItem>
        <MenuItem onClick={handleDisconnect}>
          <ListItemIcon>
            <Logout fontSize="small" />
          </ListItemIcon>
          <ListItemText>Disconnect</ListItemText>
        </MenuItem>
      </Menu>
    </Box>
  );
} 