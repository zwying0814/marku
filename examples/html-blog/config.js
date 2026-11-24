/*
 * 这是浏览器端公开配置。Site Key 会出现在页面源码中，它不是服务端密钥。
 * 创建站点或执行 `pnpm seed` 后，把 Site Key 填到 siteKey。
 */
const markuRuntimeConfig = window.MARKU_EXAMPLE ?? {}

window.MARKU_EXAMPLE = {
  apiBaseUrl:
    window.MARKU_API_BASE_URL ??
    markuRuntimeConfig.apiBaseUrl ??
    'http://localhost:36580/api/v1',
  siteKey:
    window.MARKU_SITE_KEY ??
    markuRuntimeConfig.siteKey ??
    'mk_replace_with_your_site_key',
}
