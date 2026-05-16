/**
 * 全局 message 实例
 * 通过 Ant Design App.useApp() 获取，支持动态主题和 ConfigProvider 上下文
 *
 * 使用方式：
 *   import { msg } from '../hooks/useMessage'
 *   msg.success('操作成功')
 *   msg.error('操作失败')
 */
import type { MessageInstance } from 'antd/es/message/interface'

// 全局单例，由 GlobalMessageProvider 在 App 挂载时注入
let _msg: MessageInstance | null = null

export function setGlobalMessage(instance: MessageInstance) {
  _msg = instance
}

// 代理对象：调用时若实例未初始化则静默忽略（不会崩溃）
export const msg: MessageInstance = new Proxy({} as MessageInstance, {
  get(_target, prop) {
    return (...args: unknown[]) => {
      if (_msg) {
        return (_msg as unknown as Record<string, (...a: unknown[]) => unknown>)[prop as string]?.(...args)
      }
    }
  },
})
