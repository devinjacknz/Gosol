import { ThemeConfig } from 'antd'

export const lightTheme: ThemeConfig = {
  token: {
    colorPrimary: '#1677ff',
    colorSuccess: '#52c41a',
    colorWarning: '#faad14',
    colorError: '#ff4d4f',
    colorInfo: '#1677ff',
    borderRadius: 6,
  },
  components: {
    Table: {
      headerBg: '#fafafa',
      headerColor: '#262626',
      rowHoverBg: '#f5f5f5',
    },
    Card: {
      headerBg: '#ffffff',
    },
  },
}

export const darkTheme: ThemeConfig = {
  token: {
    colorPrimary: '#1677ff',
    colorBgBase: '#141414',
    colorTextBase: '#ffffff',
    colorBorder: '#303030',
    colorSuccess: '#52c41a',
    colorWarning: '#faad14',
    colorError: '#ff4d4f',
    colorInfo: '#1677ff',
    borderRadius: 6,
  },
  components: {
    Table: {
      headerBg: '#1f1f1f',
      headerColor: '#ffffff',
      rowHoverBg: '#262626',
    },
    Card: {
      headerBg: '#1f1f1f',
    },
  },
} 