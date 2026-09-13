/**
 * 开发环境日志。生产构建下静默，避免生产控制台残留调试输出（项目约定：不残留 console 调用）。
 */
export function devWarn(...args: unknown[]): void {
  if (import.meta.env.DEV) {
    console.warn(...args)
  }
}
