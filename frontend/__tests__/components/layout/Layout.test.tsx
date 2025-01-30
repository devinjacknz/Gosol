import React from 'react';
import { render, screen } from '../../utils/test-utils';
import Layout from '@/components/layout/Layout';

describe('Layout Component', () => {
  it('renders children correctly', () => {
    render(
      <Layout>
        <div data-testid="test-child">Test Content</div>
      </Layout>
    );

    expect(screen.getByTestId('test-child')).toBeInTheDocument();
  });

  it('renders header and navigation', () => {
    render(
      <Layout>
        <div>Content</div>
      </Layout>
    );

    // 验证页头存在
    expect(screen.getByRole('banner')).toBeInTheDocument();
    
    // 验证导航菜单存在
    expect(screen.getByRole('navigation')).toBeInTheDocument();
  });

  it('toggles sidebar correctly', async () => {
    render(
      <Layout>
        <div>Content</div>
      </Layout>
    );

    // 找到侧边栏切换按钮
    const toggleButton = screen.getByRole('button', { name: /toggle sidebar/i });
    
    // 点击按钮
    await userEvent.click(toggleButton);
    
    // 验证侧边栏状态改变
    expect(screen.getByRole('complementary')).toHaveClass('sidebar-collapsed');
    
    // 再次点击
    await userEvent.click(toggleButton);
    
    // 验证侧边栏恢复原状
    expect(screen.getByRole('complementary')).not.toHaveClass('sidebar-collapsed');
  });

  it('renders footer with correct content', () => {
    render(
      <Layout>
        <div>Content</div>
      </Layout>
    );

    const footer = screen.getByRole('contentinfo');
    expect(footer).toBeInTheDocument();
    expect(footer).toHaveTextContent(/© 2024/);
  });

  it('applies correct theme based on system preference', () => {
    // 模拟系统深色模式
    window.matchMedia = jest.fn().mockImplementation(query => ({
      matches: query === '(prefers-color-scheme: dark)',
      media: query,
      onchange: null,
      addListener: jest.fn(),
      removeListener: jest.fn(),
    }));

    render(
      <Layout>
        <div>Content</div>
      </Layout>
    );

    // 验证深色主题样式是否应用
    expect(document.documentElement).toHaveClass('dark');
  });
}); 