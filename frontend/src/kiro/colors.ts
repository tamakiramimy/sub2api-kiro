/**
 * Kiro 平台专属颜色定义。
 *
 * 仅在此维护 Kiro 的品牌色值，`utils/platformColors.ts` 中的各 Record 只引用
 * 这里的字段做展开合并，不直接内联 Kiro 的颜色字符串。
 */
export const kiroColors = {
  badge: 'bg-orange-500/10 text-orange-600 border-orange-500/30 dark:text-orange-400',
  badgeLight: 'bg-orange-500/10 text-orange-600 dark:bg-orange-500/10 dark:text-orange-300',
  border: 'border-orange-500/20 dark:border-orange-500/20',
  accentBar: 'bg-gradient-to-r from-orange-400 to-orange-500',
  text: 'text-orange-600 dark:text-orange-400',
  icon: 'text-orange-500 dark:text-orange-400',
  button:
    'bg-orange-500 text-white hover:bg-orange-600 active:bg-orange-700 dark:bg-orange-500/80 dark:hover:bg-orange-500',
  discount: 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300',
  gradient: 'from-orange-500 to-orange-600',
  gradientText: 'text-orange-100',
  gradientSubtext: 'text-orange-200',
} as const
