import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { AppRouterCacheProvider } from '@mui/material-nextjs/v14-appRouter';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import theme from '@/lib/theme';
import { Providers } from './providers';
import { SessionProvider } from 'next-auth/react';
import { ZkLoginProvider } from '@/contexts/ZkLoginContext';
import { WalletProvider } from '@/contexts/WalletContext';
import { AssetsProvider } from '@/contexts/AssetsContext';
import ThemeRegistry from '@/components/ThemeRegistry/ThemeRegistry';

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5000,
      refetchOnWindowFocus: false,
    },
  },
});

export const metadata: Metadata = {
  title: "Trading Platform",
  description: "Advanced trading platform with real-time market analysis",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <ThemeRegistry>
          <SessionProvider>
            <ZkLoginProvider>
              <WalletProvider>
                <AssetsProvider>
                  <Providers>
                    <AppRouterCacheProvider>
                      <ThemeProvider theme={theme}>
                        <CssBaseline />
                        <QueryClientProvider client={queryClient}>
                          {children}
                        </QueryClientProvider>
                      </ThemeProvider>
                    </AppRouterCacheProvider>
                  </Providers>
                </AssetsProvider>
              </WalletProvider>
            </ZkLoginProvider>
          </SessionProvider>
        </ThemeRegistry>
      </body>
    </html>
  );
}
