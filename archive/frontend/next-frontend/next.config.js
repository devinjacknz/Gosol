/** @type {import('next').NextConfig} */
const nextConfig = {
  // Enable static exports if needed
  // output: 'export',
  
  // Configure redirects for gradual migration
  async redirects() {
    return [
      {
        source: '/legacy/:path*',
        destination: 'http://localhost:3001/:path*', // Old CRA app
        permanent: false,
      },
    ]
  },

  // Configure API proxy
  async rewrites() {
    return {
      fallback: [
        {
          source: '/api/:path*',
          destination: 'http://localhost:8080/api/:path*', // Go backend
        },
      ],
    }
  },

  // Webpack configuration for WebSocket support
  webpack: (config, { isServer }) => {
    if (!isServer) {
      config.resolve.fallback = {
        ...config.resolve.fallback,
        net: false,
        tls: false,
      }
    }
    return config
  },

  // Enable SWC minification
  swcMinify: true,

  // Configure image domains if needed
  images: {
    domains: ['localhost'],
  },
}

module.exports = nextConfig 