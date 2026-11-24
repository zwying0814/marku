import { fileURLToPath } from 'node:url'

import { defineConfig } from '@rspress/core'

export default defineConfig({
  root: 'docs',
  outDir: 'doc_build',
  title: 'Marku',
  description: '面向静态博客的无头动态内容 API',
  lang: 'zh',
  base: '/docs/',
  sitemap: false,
  logo: '/marku.svg',
  logoText: 'Marku',
  icon: '/marku.svg',
  globalStyles: fileURLToPath(
    new URL('./styles/global.css', import.meta.url),
  ),
  search: {
    mode: 'local',
  },
  themeConfig: {
    search: true,
    darkMode: 'auto',
    lastUpdated: true,
    nav: [
      {
        text: '指南',
        link: '/guide/getting-started',
        activeMatch: '/guide/',
      },
      {
        text: 'API',
        link: '/api/overview',
        activeMatch: '/api/',
      },
      {
        text: '部署',
        link: '/deploy/docker',
        activeMatch: '/deploy/',
      },
    ],
    sidebar: {
      '/guide/': [
        {
          text: '开始使用',
          items: [
            { text: '快速开始', link: '/guide/getting-started' },
            { text: '博客接入', link: '/guide/integration' },
            { text: '纯 HTML 示例', link: '/guide/html-example' },
          ],
        },
        {
          text: '能力指南',
          items: [
            { text: '评论', link: '/guide/comments' },
            { text: '浏览量与点赞', link: '/guide/analytics' },
          ],
        },
      ],
      '/api/': [
        {
          text: 'API 参考',
          items: [
            { text: '总览', link: '/api/overview' },
            { text: '公共接口', link: '/api/public' },
            { text: '管理接口', link: '/api/admin' },
            { text: '错误与限流', link: '/api/errors' },
          ],
        },
      ],
      '/deploy/': [
        {
          text: '部署与运维',
          items: [
            { text: '环境变量', link: '/deploy/configuration' },
            { text: 'Docker Compose', link: '/deploy/docker' },
            { text: '反向代理与 HTTPS', link: '/deploy/reverse-proxy' },
            { text: '备份与恢复', link: '/deploy/backup' },
            { text: 'v0.2 发布检查', link: '/deploy/release' },
          ],
        },
      ],
    },
    footer: {
      message:
        'Marku · Headless dynamic content for static blogs · v0.2.0',
    },
  },
})
