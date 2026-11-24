# @marku/shared

由 OpenAPI 生成的共享 TypeScript 类型。

## 生成

```bash
pnpm generate:types
```

或：

```bash
pnpm --filter @marku/shared generate
```

生成文件：`src/api.d.ts`。

## 使用

```ts
import type { components, paths } from '@marku/shared'

type Site = components['schemas']['Site']
```
