#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";

const publicDir = process.argv[2] ?? "website/public";
const htmlFiles = [];

function walk(directory) {
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    const file = path.join(directory, entry.name);
    if (entry.isDirectory()) walk(file);
    else if (entry.isFile() && entry.name.endsWith(".html")) htmlFiles.push(file);
  }
}

function matches(document, expression) {
  return [...document.matchAll(expression)];
}

function fail(file, message) {
  const relative = path.relative(publicDir, file);
  throw new Error(`${relative}: ${message}`);
}

walk(publicDir);
htmlFiles.sort();

const titles = new Map();
const canonicals = new Map();
let canonicalPages = 0;
let aliases = 0;

for (const file of htmlFiles) {
  const document = fs.readFileSync(file, "utf8");
  const isAlias = /<meta\s+http-equiv=(?:["']?refresh["']?)[^>]*>/i.test(document);

  if (!/^<!doctype html>/i.test(document)) fail(file, "missing HTML5 doctype");
  if (!/<html\b[^>]*\blang=(?:["']?en["']?)/i.test(document)) fail(file, "missing English html lang");
  if (!/<meta\s+charset=(?:["']?utf-8["']?)/i.test(document)) fail(file, "missing UTF-8 charset");
  if (!/<\/head>/i.test(document) || !/<\/html>/i.test(document)) fail(file, "incomplete document structure");
  if (/>\s*(?:Column|Lorem ipsum)\s*</i.test(document)) fail(file, "placeholder content is rendered");

  const titleMatches = matches(document, /<title>(.*?)<\/title>/gis);
  if (titleMatches.length !== 1 || titleMatches[0][1].trim() === "") fail(file, "expected one non-empty title");

  const canonicalMatches = matches(document, /<link\s+rel=(?:["']?canonical["']?)[^>]*\bhref=(?:["']?)([^\s"'>]+)["']?[^>]*>/gi);
  if (canonicalMatches.length !== 1) fail(file, "expected exactly one canonical URL");
  const canonical = canonicalMatches[0][1];
  if (!canonical.startsWith("https://mergewayhq.github.io/")) fail(file, `non-production canonical URL ${canonical}`);

  for (const image of matches(document, /<img\b[^>]*>/gi)) {
    if (!/\balt=(?:"[^"]*"|'[^']*'|[^\s>]+)/i.test(image[0])) fail(file, `image lacks alt text: ${image[0]}`);
  }

  const ids = matches(document, /\bid=(?:"([^"]+)"|'([^']+)'|([^\s>]+))/gi).map((match) => match[1] ?? match[2] ?? match[3]);
  const duplicateIds = ids.filter((id, index) => ids.indexOf(id) !== index);
  if (duplicateIds.length > 0) fail(file, `duplicate id ${duplicateIds[0]}`);

  if (isAlias) {
    aliases += 1;
    continue;
  }

  canonicalPages += 1;
  if (!/<main\b/i.test(document)) fail(file, "missing main landmark");
  const breadcrumb = document.match(/<nav\b[^>]*aria-label=(?:["']?Breadcrumb["']?)[^>]*>([\s\S]*?)<\/nav>/i);
  if (path.relative(publicDir, file) === "index.html") {
    if (breadcrumb) fail(file, "homepage must not render breadcrumbs");
    if (!/<img\b[^>]*\bclass=(?:"[^"]*\blanding-logo\b[^"]*"|'[^']*\blanding-logo\b[^']*'|landing-logo(?=\s|>))[^>]*\bsrc=(?:"\/mergeway-logo\.svg"|'\/mergeway-logo\.svg'|\/mergeway-logo\.svg(?=\s|>))/i.test(document)) {
      fail(file, "homepage is missing the Mergeway logo");
    }
  } else if (breadcrumb && /<li><a\s+href=(?:"\/(?:docs\/)?"|'\/(?:docs\/)?'|\/(?:docs\/)?(?=\s|>))/i.test(breadcrumb[1])) {
    fail(file, "breadcrumb must not link to the homepage or documentation root");
  }
  const headings = matches(document, /<h([1-6])\b/gi).map((match) => Number(match[1]));
  if (headings.length === 0 || headings[0] !== 1) fail(file, "heading outline does not begin with h1");
  for (let index = 1; index < headings.length; index += 1) {
    if (headings[index] > headings[index - 1] + 1) fail(file, `heading level skips from h${headings[index - 1]} to h${headings[index]}`);
  }

  const title = titleMatches[0][1].replace(/\s+/g, " ").trim();
  if (titles.has(title)) fail(file, `duplicate title also used by ${titles.get(title)}`);
  if (canonicals.has(canonical)) fail(file, `duplicate canonical also used by ${canonicals.get(canonical)}`);
  titles.set(title, path.relative(publicDir, file));
  canonicals.set(canonical, path.relative(publicDir, file));
}

console.log(`Basic HTML validity: ${htmlFiles.length} documents, ${canonicalPages} canonical pages, ${aliases} aliases; unique titles, canonicals, IDs, heading outlines, language, landmarks, and image alternatives verified.`);
