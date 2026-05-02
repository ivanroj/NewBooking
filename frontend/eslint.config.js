import js from "@eslint/js";
import globals from "globals";

/** @type {import("eslint").Linter.Config[]} */
export default [
	js.configs.recommended,
	{
		name: "frontend/browser",
		files: ["**/*.js"],
		languageOptions: {
			ecmaVersion: "latest",
			sourceType: "script",
			globals: {
				...globals.browser,
				Telegram: "readonly",
			},
		},
		rules: {
			// Keep “prod” ergonomics consistent for now; tighten later once modules land.
			"no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
		},
	},
];
