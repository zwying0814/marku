# Next.js 接入示例（不纳入 v0.1）

用于演示 Next.js 博客如何通过公共 API 接入 Marku：

- Client Component 直连评论、浏览量、点赞
- canonical path 处理
- 审核状态与 `429` 错误处理
- 可选的 Route Handler 代理模式

该示例不纳入 v0.1。v0.1 以 `examples/html-blog` 作为框架无关的完整接入参考；
Next.js 示例可在后续版本中复用其请求、canonical path 与错误处理逻辑。
