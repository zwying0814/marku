import { createReadStream } from 'node:fs'
import { stat } from 'node:fs/promises'
import { createServer } from 'node:http'
import { dirname, extname, resolve, sep } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = dirname(fileURLToPath(import.meta.url))
const port = Number(process.env.PORT ?? 36583)

const contentTypes = {
  '.css': 'text/css; charset=utf-8',
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.txt': 'text/plain; charset=utf-8',
}

function sendFile(response, filePath, method) {
  response.writeHead(200, {
    'Content-Type':
      contentTypes[extname(filePath).toLowerCase()] ??
      'application/octet-stream',
    'Cache-Control': 'no-cache',
  })
  if (method === 'HEAD') {
    response.end()
    return
  }
  createReadStream(filePath).pipe(response)
}

const server = createServer(async (request, response) => {
  if (request.method !== 'GET' && request.method !== 'HEAD') {
    response.writeHead(405, { Allow: 'GET, HEAD' })
    response.end('Method Not Allowed')
    return
  }

  const pathname = decodeURIComponent(
    new URL(request.url ?? '/', 'http://localhost').pathname,
  )
  const requestedPath = resolve(root, `.${pathname}`)

  if (requestedPath !== root && !requestedPath.startsWith(`${root}${sep}`)) {
    response.writeHead(403)
    response.end('Forbidden')
    return
  }

  try {
    const fileStat = await stat(requestedPath)
    if (fileStat.isFile()) {
      sendFile(response, requestedPath, request.method)
      return
    }
  } catch {
    // Extensionless article URLs fall through to index.html below.
  }

  if (extname(pathname) === '') {
    sendFile(response, resolve(root, 'index.html'), request.method)
    return
  }

  response.writeHead(404, { 'Content-Type': 'text/plain; charset=utf-8' })
  response.end('Not Found')
})

server.listen(port, '127.0.0.1', () => {
  console.log(
    `Marku HTML example → http://127.0.0.1:${port}/posts/static-blogs-dynamic-content`,
  )
})
