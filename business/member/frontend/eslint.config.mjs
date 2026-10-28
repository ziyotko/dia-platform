import js from '@eslint/js'
import prettier from '@vue/eslint-config-prettier'
import { withVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'
import globals from 'globals'
import pluginVue from 'eslint-plugin-vue'

export default withVueTs(
  {
    ignores: ['**/dist/**', '**/node_modules/**'],
  },
  js.configs.recommended,
  pluginVue.configs['flat/recommended'],
  vueTsConfigs.recommended,
  {
    files: ['**/*.{js,mjs,cjs,jsx,ts,mts,cts,tsx,vue}'],
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
      parserOptions: {
        ecmaVersion: 'latest',
      },
    },
    rules: {
      'vue/multi-word-component-names': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-empty-object-type': ['error', { allowObjectTypes: 'always' }],
      '@typescript-eslint/no-unused-vars': [
        'warn',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
      ],
      'no-constant-condition': ['error', { checkLoops: false }],
      'no-empty': ['error', { allowEmptyCatch: true }],
    },
  },
  prettier
)
