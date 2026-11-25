---
pageType: home
hero:
  name: Marku
  text: 给静态博客，接上真实发生的互动
  tagline: 一个 API 管理多个站点的评论、浏览量与点赞。博客继续静态发布，动态能力按需加载。
  image:
    src: /marku.svg
    alt: Marku
  actions:
    - theme: brand
      text: 快速开始
      link: /guide/getting-started
    - theme: alt
      text: 纯 HTML 示例
      link: /guide/html-example
features:
  - icon: 🧩
    title: 多站点隔离
    details: 每个站点拥有独立 Site Key、评论队列、浏览量、点赞和趋势数据。
  - icon: 🔌
    title: API First
    details: 公共 REST API 是唯一必需接入层，不绑定 Next.js、Astro 或任何前端框架。
  - icon: 🛡️
    title: 评论治理
    details: 先审后发、楼中楼、管理员回复、批量审核、反垃圾与频率限制。
  - icon: 📈
    title: 轻量统计
    details: 页面浏览量、独立访客、点赞趋势与热门页面排行，无需第三方脚本。
  - icon: 🪶
    title: 静态优先
    details: 文章本身仍是静态 HTML。互动模块失败时不影响阅读与 SEO。
  - icon: 🚀
    title: 自部署
    details: 单个二进制内置 API、Admin 与官方组件；官网独立部署，默认 SQLite，也支持 PostgreSQL、MySQL。
---

## 一条很薄的动态层

Marku 不接管博客构建。你仍然生成 HTML，只是把评论区、浏览量和点赞交给一个独立 API：

```text
静态博客页面
    │  X-Site-Key
    ▼
Marku Public API
    │
    ├── 评论 / 审核 / 回复
    ├── 浏览量 / 独立访客
    └── 点赞 / 取消点赞
```

## 从这里开始

- [5 分钟启动 Marku](/guide/getting-started)
- [把一个现有博客接入公共 API](/guide/integration)
- [查看纯 HTML 完整示例](/guide/html-example)
- [阅读 API 参考](/api/overview)
