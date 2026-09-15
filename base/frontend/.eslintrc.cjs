/* eslint-env node */
/**
 * 底座前端 ESLint 配置（eslintrc 风格，对应 eslint 8.x）。
 * 之前仓库缺少该文件，导致 `npm run lint` 全项目解析报错。
 */
module.exports = {
  root: true,
  env: {
    browser: true,
    es2022: true,
    node: true
  },
  extends: [
    'plugin:vue/vue3-essential',
    'eslint:recommended',
    '@vue/eslint-config-typescript',
    '@vue/eslint-config-prettier/skip-formatting'
  ],
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module'
  },
  rules: {
    // 页面组件文件名多为 index.vue，单文件多词限制不适用
    'vue/multi-word-component-names': 'off',
    // 该项目大量使用 any 承接后端响应，属既有风格
    '@typescript-eslint/no-explicit-any': 'off',
    '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
    'no-empty': ['error', { allowEmptyCatch: true }]
  },
  ignorePatterns: ['dist/**', 'node_modules/**', '*.d.ts']
}
