import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { createRequire } from 'node:module';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { after, before, test } from 'node:test';
import { fileURLToPath } from 'node:url';

const require = createRequire(import.meta.url);
const docsDir = fileURLToPath(new URL('.', import.meta.url));
let tempDir;
let html;

before(() => {
  tempDir = mkdtempSync(join(tmpdir(), 'osmosis-docs-test-'));
  const output = join(tempDir, 'index.html');
  execFileSync(process.execPath, [join(docsDir, 'build-docs.mjs'), output]);
  html = readFileSync(output, 'utf8');
});
after(() => {
  if (tempDir) rmSync(tempDir, { recursive: true, force: true });
});

test('browser renderer is self-contained rather than a remote prebuilt bundle', () => {
  assert.doesNotMatch(html, /<script\b[^>]*\bsrc\s*=/i);
  assert.match(html, /Redoc\.hydrate\(__redoc_state, container\)/);
});

test('browser sanitizer matches the patched dependency instead of a vendored copy', () => {
  const version = require('dompurify').version;
  const versions = [...html.matchAll(/DOMPurify (\d+\.\d+\.\d+)/g)].map(match => match[1]);
  assert.deepEqual([...new Set(versions)], [version]);
});

test('documentation remains pre-rendered with a stable local specification download', () => {
  assert.match(html, /<title>Osmosis - gRPC Gateway docs<\/title>/);
  assert.match(html, /href="openapi\.yaml"/);
  assert.match(html, /<div id="redoc"><div/);
});
