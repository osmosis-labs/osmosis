import { execFileSync } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { build } from 'esbuild';

const require = createRequire(import.meta.url);
const docsDir = fileURLToPath(new URL('.', import.meta.url));
// An optional output path lets validation avoid rewriting committed assets.
const output = resolve(docsDir, process.argv[2] || 'static/openapi/index.html');
const tempDir = mkdtempSync(join(tmpdir(), 'osmosis-docs-'));

try {
  const htmlPath = join(tempDir, 'index.html');
  execFileSync(process.execPath, [
    require.resolve('@redocly/cli/bin/cli.js'),
    'build-docs', 'static/openapi/openapi.yaml', '--output', htmlPath,
    '--theme.openapi.downloadDefinitionUrl=openapi.yaml',
  ], {
    cwd: docsDir,
    env: { ...process.env, NODE_ENV: 'production', REDOCLY_TELEMETRY: 'off' },
    stdio: 'inherit',
  });

  // Do not use Redoc's prebuilt standalone bundle: it embeds its own sanitizer
  // and therefore bypasses Yarn security resolutions. Build the browser library
  // against our locked dependencies instead, preserving self-contained HTML.
  const bundle = await build({
    entryPoints: [join(docsDir, 'redoc-browser.mjs')],
    bundle: true,
    platform: 'browser',
    format: 'iife',
    globalName: 'Redoc',
    define: { 'process.env.NODE_ENV': '"production"', global: 'globalThis' },
    minify: true,
    write: false,
  });
  const html = readFileSync(htmlPath, 'utf8');
  const cdnScript = /<script src="https:\/\/cdn\.redocly\.com\/redoc\/v[^"<>]+\/bundles\/redoc\.standalone\.js"><\/script>/g;
  if ([...html.matchAll(cdnScript)].length !== 1) {
    throw new Error('Expected exactly one Redoc CDN script; review the upstream template');
  }
  const script = bundle.outputFiles[0].text.replace(/<\/script/gi, '<\\/script');
  writeFileSync(output, html.replace(cdnScript, () => `<script>${script}</script>`));
  console.log(`Saved self-contained documentation to ${output}`);
} finally {
  rmSync(tempDir, { recursive: true, force: true });
}
