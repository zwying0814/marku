import { execFileSync } from 'node:child_process'
import { cp, mkdir, readdir, rm } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const pnpm = process.platform === 'win32' ? 'pnpm.cmd' : 'pnpm'

function run(command, args, cwd = root, env = {}) {
  console.log(`$ ${command} ${args.join(' ')}`)
  execFileSync(command, args, {
    cwd,
    env: { ...process.env, ...env },
    stdio: 'inherit',
  })
}

async function cleanDirectory(directory) {
  await mkdir(directory, { recursive: true })
  for (const entry of await readdir(directory)) {
    if (entry === '.gitkeep') continue
    await rm(join(directory, entry), { recursive: true, force: true })
  }
}

async function copyDirectory(source, target) {
  await cleanDirectory(target)
  await cp(source, target, { recursive: true })
}

if (process.env.MARKU_SKIP_WEB !== '1') {
  run(pnpm, ['--filter', '@marku/client', 'build'])
  run(pnpm, ['--filter', '@marku/embed', 'build'])
  run(pnpm, ['--filter', '@marku/admin', 'build'])

  await copyDirectory(
    join(root, 'apps/admin/dist'),
    join(root, 'apps/api/internal/webui/admin-dist'),
  )
  await copyDirectory(
    join(root, 'packages/embed/dist'),
    join(root, 'apps/api/internal/webui/embed-dist'),
  )
}

const output = resolve(root, process.env.MARKU_OUTPUT ?? 'dist/marku-api')
const targetEnv = { CGO_ENABLED: '0' }
if (process.env.MARKU_GOOS) targetEnv.GOOS = process.env.MARKU_GOOS
if (process.env.MARKU_GOARCH) targetEnv.GOARCH = process.env.MARKU_GOARCH

run(
  'go',
  ['build', '-trimpath', '-ldflags=-s -w', '-o', output, './cmd/server'],
  join(root, 'apps/api'),
  targetEnv,
)
console.log(`built ${output}`)
