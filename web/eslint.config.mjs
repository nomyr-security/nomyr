import jsxA11y from "eslint-plugin-jsx-a11y";
import babelParser from "@babel/eslint-parser";

export default [
  {ignores: [".next/**", "out/**", "node_modules/**"]},
  {
    files: ["**/*.{js,jsx,ts,tsx}"],
    languageOptions: {
      parser: babelParser,
      parserOptions: {
        requireConfigFile: false,
        babelOptions: {plugins: ["@babel/plugin-syntax-typescript", "@babel/plugin-syntax-jsx"]},
      },
    },
    plugins: {"jsx-a11y": jsxA11y},
    rules: jsxA11y.flatConfigs.recommended.rules,
  },
];
