describe('Trading Flow', () => {
  beforeEach(() => {
    // 访问交易页面
    cy.visit('/trading')
    
    // 等待 WebSocket 连接
    cy.wait(1000)
  })

  it('completes a full trading flow', () => {
    // 选择交易对
    cy.get('[data-testid=symbol-select]')
      .click()
      .get('[data-testid=symbol-option-btc]')
      .click()

    // 验证K线图加载
    cy.get('[data-testid=chart-container]')
      .should('be.visible')
      .and('have.css', 'height', '500px')

    // 填写订单表单
    cy.get('[data-testid=price-input]')
      .type('50000')
    
    cy.get('[data-testid=size-input]')
      .type('1')

    cy.get('[data-testid=side-select]')
      .click()
      .get('[data-testid=side-option-buy]')
      .click()

    // 提交订单
    cy.get('[data-testid=submit-order]')
      .click()

    // 验证订单出现在列表中
    cy.get('[data-testid=orders-table]')
      .contains('BTC/USDT')
      .should('be.visible')

    // 等待订单更新
    cy.wait(2000)

    // 取消订单
    cy.get('[data-testid=cancel-order]')
      .first()
      .click()

    // 验证订单被取消
    cy.get('[data-testid=orders-table]')
      .contains('BTC/USDT')
      .should('not.exist')
  })

  it('handles market data updates', () => {
    // 验证价格更新
    cy.get('[data-testid=last-price]')
      .invoke('text')
      .then((text1) => {
        cy.wait(2000)
        cy.get('[data-testid=last-price]')
          .invoke('text')
          .should('not.eq', text1)
      })

    // 验证深度图更新
    cy.get('[data-testid=depth-chart]')
      .should('be.visible')
      .and('have.css', 'height', '200px')
  })

  it('displays error messages', () => {
    // 测试无效订单
    cy.get('[data-testid=price-input]')
      .type('-1')
    
    cy.get('[data-testid=submit-order]')
      .click()

    // 验证错误提示
    cy.get('[role=alert]')
      .should('be.visible')
      .and('contain', '价格必须大于0')
  })

  it('persists user preferences', () => {
    // 更改交易对
    cy.get('[data-testid=symbol-select]')
      .click()
      .get('[data-testid=symbol-option-eth]')
      .click()

    // 刷新页面
    cy.reload()

    // 验证选择被保持
    cy.get('[data-testid=symbol-select]')
      .should('have.value', 'ETH/USDT')
  })

  it('handles WebSocket reconnection', () => {
    // 模拟断开连接
    cy.window().then((win) => {
      win.dispatchEvent(new Event('offline'))
    })

    // 验证断开连接提示
    cy.get('[role=alert]')
      .should('be.visible')
      .and('contain', '网络连接已断开')

    // 模拟重新连接
    cy.window().then((win) => {
      win.dispatchEvent(new Event('online'))
    })

    // 验证重连成功
    cy.get('[role=alert]')
      .should('be.visible')
      .and('contain', '网络已重新连接')
  })

  it('handles high frequency updates', () => {
    // 切换到高频交易对
    cy.get('[data-testid=symbol-select]')
      .click()
      .get('[data-testid=symbol-option-btc]')
      .click()

    // 验证快速更新
    cy.get('[data-testid=trade-list]')
      .children()
      .should('have.length.gt', 10)
      .then(($trades) => {
        const initialCount = $trades.length
        cy.wait(1000)
        cy.get('[data-testid=trade-list]')
          .children()
          .should('have.length.gt', initialCount)
      })
  })
}) 