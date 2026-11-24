FROM node:24-alpine AS build

RUN corepack enable && corepack prepare pnpm@12.23.0 --activate

WORKDIR /repo
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/admin/package.json apps/admin/package.json
COPY packages/shared/package.json packages/shared/package.json
RUN pnpm install --frozen-lockfile --filter @marku/admin

COPY packages/shared packages/shared
COPY apps/admin apps/admin
ARG VITE_API_BASE_URL=http://localhost:36580/api/v1
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
RUN pnpm --filter @marku/admin build

FROM nginxinc/nginx-unprivileged:1.27-alpine

COPY apps/admin/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /repo/apps/admin/dist /usr/share/nginx/html

EXPOSE 36581
