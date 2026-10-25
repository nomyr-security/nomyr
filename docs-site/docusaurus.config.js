import { themes as prismThemes } from 'prism-react-renderer';

const config = {
  title: 'Nomyr Documentation',
  tagline: 'Documentation for the Nomyr Phase 0 foundation scaffold',
  url: 'https://docs.nomyr.dev',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  markdown: { hooks: { onBrokenMarkdownLinks: 'throw' } },
  favicon: 'img/favicon.svg',
  presets: [
    [
      'classic',
      {
        docs: { sidebarPath: './sidebars.js', editUrl: undefined },
        blog: false,
        theme: { customCss: './src/css/custom.css' },
      },
    ],
  ],
  plugins: [
    [
      require.resolve('@easyops-cn/docusaurus-search-local'),
      { hashed: true, language: ['en'], docsRouteBasePath: 'docs', indexBlog: false },
    ],
  ],
  themeConfig: {
    navbar: {
      title: 'Nomyr',
      items: [
        { type: 'docSidebar', sidebarId: 'tutorialSidebar', position: 'left', label: 'Documentation' },
        { href: 'https://github.com/nomyr-security/nomyr', label: 'GitHub', position: 'right' },
      ],
    },
    prism: { theme: prismThemes.github, darkTheme: prismThemes.dracula },
  },
};

export default config;
