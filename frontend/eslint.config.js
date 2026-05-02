import js from "@eslint/js";
import globals from "globals";

/** @type {import("eslint").Linter.Config[]} */
export default [
	js.configs.recommended,
	{
		name: "frontend/browser",
		files: ["**/*.js"],
		ignores: ["tests/**", "jest.config.js"],
		languageOptions: {
			ecmaVersion: "latest",
			sourceType: "module",
			globals: {
				...globals.browser,
				Telegram: "readonly",
			},
		},
		rules: {
			"no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
		},
	},
	{
		name: "frontend/tests",
		files: ["tests/**/*.js"],
		languageOptions: {
			ecmaVersion: "latest",
			sourceType: "module",
			globals: {
				...globals.browser,
				...globals.jest,
				...globals.node,
			},
		},
		rules: {
			"no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
		},
	},
	{
		name: "frontend/config",
		files: ["jest.config.js"],
		languageOptions: {
			ecmaVersion: "latest",
			sourceType: "module",
			globals: {
				...globals.node,
			},
		},
	},
	{
		name: "frontend/mocks",
		files: ["tests/__mocks__/**/*.js"],
		languageOptions: {
			ecmaVersion: "latest",
			sourceType: "commonjs",
			globals: {
				...globals.node,
			},
		},
	},
];
