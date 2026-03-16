import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'
import i18next from 'eslint-plugin-i18next'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
  },
  {
    // i18next enforcement — separate block so files glob takes effect correctly
    ...i18next.configs['flat/recommended'],
    files: ['**/*.{ts,tsx}'],
    rules: {
      ...i18next.configs['flat/recommended'].rules,
      'i18next/no-literal-string': ['error', {
        mode: 'jsx-only',
        'jsx-attributes': {
          include: ['title', 'placeholder', 'aria-label', 'alt'],
        },
        // ignore: pure numbers/decimals, alphanumeric tokens with spaces (e.g. "MacBook Pro"), ALL_CAPS constants
        ignore: [/^\d+(\.\d+)?$/, /^[a-zA-Z0-9_\-\.\/ ]+$/, /^[A-Z_]+$/],
      }],
    },
  },
])
