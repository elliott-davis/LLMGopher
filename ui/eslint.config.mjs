import path from "node:path";
import { fileURLToPath } from "node:url";
import { FlatCompat } from "@eslint/eslintrc";
import { defineConfig, globalIgnores } from "eslint/config";
import playwright from "eslint-plugin-playwright";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const compat = new FlatCompat({ baseDirectory: __dirname });

const eslintConfig = defineConfig([
  ...compat.extends("next/core-web-vitals", "next/typescript"),
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    "node_modules/**",
    ".next/**",
    "dist/**",
    "out/**",
    "build/**",
    "coverage/**",
    "*.min.js",
    "next-env.d.ts",
  ]),
]);

// Playwright: enforce role/label-first selectors in E2E specs (Decision 3, research.md)
const playwrightConfig = {
  files: ["tests/e2e/**/*.spec.ts"],
  ...playwright.configs["flat/recommended"],
  rules: {
    ...playwright.configs["flat/recommended"].rules,
    // Prefer getByRole/getByLabel over raw CSS/XPath
    "playwright/no-raw-locators": "error",
  },
};

export default [...eslintConfig, playwrightConfig];
