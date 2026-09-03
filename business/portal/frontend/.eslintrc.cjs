/* eslint-env node */
module.exports = {
  root: true,
  env: {
    browser: true,
    es2021: true,
    node: true,
  },
  extends: [
    'plugin:vue/vue3-recommended',
    'eslint:recommended',
    '@vue/eslint-config-typescript',
    '@vue/eslint-config-prettier',
  ],
  parserOptions: {
    ecmaVersion: 'latest',
  },
  rules: {
    // 视图/组件大量使用单词命名（如 tag.vue、404.vue），关闭多单词组件名限制
    'vue/multi-word-component-names': 'off',
    // 业务代码中广泛使用 any 与类型断言，保持宽松以免误报
    '@typescript-eslint/no-explicit-any': 'off',
    // 未使用变量：下划线开头的参数/变量视为有意忽略
    '@typescript-eslint/no-unused-vars': [
      'warn',
      { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
    ],
    // 允许 while(true) 搭配 break 的合法循环写法（如 RTF 解析）
    'no-constant-condition': ['error', { checkLoops: false }],
    // 允许空 catch（业务中用于吞掉可预期的异常）
    'no-empty': ['error', { allowEmptyCatch: true }],
    // 交由 Prettier 统一格式，仅告警（配合 lint --fix 自动修复）
    'prettier/prettier': 'warn',
  },
}
