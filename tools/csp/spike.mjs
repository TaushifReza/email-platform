import { createHash } from "node:crypto";
import { readdir, readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";

const outDir = process.argv[2] ?? "frontend/out";

async function htmlFiles(dir) {
  const entries = await readdir(dir, { withFileTypes: true });
  const files = [];

  for (const entry of entries) {
    const path = join(dir, entry.name);

    if (entry.isDirectory()) {
      files.push(...(await htmlFiles(path)));
    } else if (entry.name.endsWith(".html")) {
      files.push(path);
    }
  }

  return files;
}

function inlineScripts(html) {
  const scripts = [];
  const pattern = /<script\b([^>]*)>([\s\S]*?)<\/script>/gi;
  let match;

  while ((match = pattern.exec(html)) !== null) {
    const attributes = match[1];
    const body = match[2];

    if (!/\bsrc\s*=/i.test(attributes) && body.trim() !== "") {
      scripts.push(body);
    }
  }

  return scripts;
}

const files = await htmlFiles(outDir);
const hashes = new Set();

for (const file of files) {
  const html = await readFile(file, "utf8");

  for (const script of inlineScripts(html)) {
    const hash = createHash("sha256").update(script).digest("base64");
    hashes.add(`'sha256-${hash}'`);
  }
}

const scriptSrc = ["'self'", ...hashes].join(" ");
const result = { scriptSrc, hashes: [...hashes] };

await writeFile(join(outDir, "csp.json"), `${JSON.stringify(result, null, 2)}\n`);

console.log(`Scanned ${files.length} HTML files, hashed ${hashes.size} inline scripts`);
console.log(result.scriptSrc);
