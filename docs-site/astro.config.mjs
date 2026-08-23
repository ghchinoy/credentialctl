import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import catppuccin from '@catppuccin/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://ghchinoy.github.io',
	base: '/credentialctl',
	integrations: [
		starlight({
			title: 'CredentialCTL',
			description: 'Agent-Aware CLI and Bubble Tea TUI for C2PA Content Credential Validation',
			plugins: [
				catppuccin({
					dark: { flavor: 'mocha', accent: 'sky' },
					light: { flavor: 'latte', accent: 'sky' },
				}),
			],
			social: [
				{
					icon: 'github',
					label: 'CredentialCTL on GitHub',
					href: 'https://github.com/ghchinoy/credentialctl',
				},
				{
					icon: 'external',
					label: 'Credentio Contributions',
					href: 'https://github.com/ghchinoy/credentio-contributions',
				},
			],
			sidebar: [
				{
					label: 'Overview',
					items: [
						{ label: 'Introduction', slug: 'index' },
						{ label: 'Quick Start', slug: 'quickstart' },
						{ label: 'System Architecture', slug: 'system-architecture' },
					],
				},
				{
					label: 'Command-Line Interface',
					items: [
						{ label: 'validate', slug: 'cli-validate' },
						{ label: 'folder', slug: 'cli-folder' },
						{ label: 'inspect', slug: 'cli-inspect' },
					],
				},
				{
					label: 'Terminal User Interface',
					items: [
						{ label: 'TUI Overview', slug: 'tui-overview' },
						{ label: 'Folder Review View', slug: 'tui-folderview' },
						{ label: 'File Inspector View', slug: 'tui-inspector' },
					],
				},
				{
					label: 'Integration & Guides',
					items: [
						{ label: 'Agent Automation (A2A)', slug: 'guide-automation' },
						{ label: 'Troubleshooting & CGO', slug: 'guide-troubleshooting' },
					],
				},
			],
		}),
	],
});
