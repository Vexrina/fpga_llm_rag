#!/usr/bin/env python3
"""
Скрапер ссылок с поддержкой:
- рендера JS-страниц (Playwright),
- файлового кэша по TTL (по умолчанию 7 дней),
- обхода "вглубь" по ссылкам (BFS).

Примеры:
  python python_scripts/link_scraper_cached.py \
    --url "https://student.itmo.ru/ru/scholarship/" \
    --max-depth 0

  python python_scripts/link_scraper_cached.py \
    --url "https://student.itmo.ru/ru/scholarship/" \
    --max-depth 1 \
    --cache-ttl-days 7 \
    --output "python_scripts/links_itmo.json"
"""

from __future__ import annotations

import argparse
import asyncio
import hashlib
import html as html_lib
import json
import re
import sys
import time
from collections import deque
from html.parser import HTMLParser
from pathlib import Path
from typing import Any
from urllib.parse import urljoin, urlparse, urlunparse

try:
    from playwright.async_api import async_playwright
except ImportError as exc:  # pragma: no cover
    raise SystemExit(
        "Не установлен playwright. Установите: pip install playwright && playwright install chromium"
    ) from exc


ROOT = Path(__file__).resolve().parent
DEFAULT_CACHE_DIR = ROOT / ".web_cache" / "link_scraper"


def normalize_url(url: str) -> str:
    parsed = urlparse(url.strip())
    scheme = parsed.scheme.lower()
    netloc = parsed.netloc.lower()
    # Убираем фрагменты, чтобы не дублировать одну страницу.
    cleaned = parsed._replace(scheme=scheme, netloc=netloc, fragment="")
    return urlunparse(cleaned)


def cache_key(url: str) -> str:
    return hashlib.sha256(normalize_url(url).encode("utf-8")).hexdigest()


def is_cache_fresh(path: Path, ttl_seconds: int) -> bool:
    if not path.exists():
        return False
    age = time.time() - path.stat().st_mtime
    return age <= ttl_seconds


def same_domain(url: str, base_url: str) -> bool:
    return urlparse(url).netloc.lower() == urlparse(base_url).netloc.lower()


def extract_links_from_html(html: str, base_url: str) -> list[str]:
    # Фолбэк-парсинг на случай чтения из кэша без запуска браузера.
    hrefs = re.findall(r'href\s*=\s*["\']([^"\']+)["\']', html, flags=re.IGNORECASE)
    links: set[str] = set()
    for href in hrefs:
        href = href.strip()
        if not href or href.startswith(("javascript:", "mailto:", "tel:", "#")):
            continue
        abs_url = normalize_url(urljoin(base_url, href))
        if abs_url.startswith(("http://", "https://")):
            links.add(abs_url)
    return sorted(links)


class _TextExtractor(HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.parts: list[str] = []
        self.block_tags = {
            "p",
            "div",
            "section",
            "article",
            "li",
            "ul",
            "ol",
            "h1",
            "h2",
            "h3",
            "h4",
            "h5",
            "h6",
            "br",
            "tr",
            "td",
            "th",
        }

    def handle_starttag(self, tag: str, attrs) -> None:
        if tag in self.block_tags:
            self.parts.append("\n")

    def handle_endtag(self, tag: str) -> None:
        if tag in self.block_tags:
            self.parts.append("\n")

    def handle_data(self, data: str) -> None:
        text = data.strip()
        if text:
            self.parts.append(text + " ")

    def get_text(self) -> str:
        raw = "".join(self.parts)
        raw = html_lib.unescape(raw)
        raw = re.sub(r"[ \t\f\v]+", " ", raw)
        raw = re.sub(r"\n\s*\n\s*\n+", "\n\n", raw)
        return raw.strip()


def _remove_tag_with_content(html: str, tag_name: str) -> str:
    return re.sub(
        rf"<{tag_name}\b[^>]*>.*?</{tag_name}>",
        " ",
        html,
        flags=re.IGNORECASE | re.DOTALL,
    )


def _pick_main_fragment(html: str) -> str:
    # Предпочитаем семантический main, если он есть.
    mains = re.findall(
        r"<main\b[^>]*>.*?</main>",
        html,
        flags=re.IGNORECASE | re.DOTALL,
    )
    if mains:
        return max(mains, key=len)

    # Иначе берем body как fallback.
    body_match = re.search(
        r"<body\b[^>]*>(.*)</body>",
        html,
        flags=re.IGNORECASE | re.DOTALL,
    )
    return body_match.group(1) if body_match else html


def extract_text_from_cached_html(html: str) -> str:
    fragment = _pick_main_fragment(html)
    for tag in ("script", "style", "noscript", "svg", "header", "nav", "footer", "aside"):
        fragment = _remove_tag_with_content(fragment, tag)
    fragment = re.sub(r"<!--.*?-->", " ", fragment, flags=re.DOTALL)

    parser = _TextExtractor()
    parser.feed(fragment)
    text = parser.get_text()

    # Удаляем типичный мусор навигации и пустые строки.
    lines = [ln.strip() for ln in text.splitlines() if ln.strip()]
    return "\n".join(lines).strip()


def extract_texts_from_cache(cache_dir: Path) -> list[dict[str, Any]]:
    items: list[dict[str, Any]] = []
    for html_file in sorted(cache_dir.glob("*.html")):
        html = html_file.read_text(encoding="utf-8", errors="ignore")
        text = extract_text_from_cached_html(html)

        meta_url = None
        json_file = html_file.with_suffix(".json")
        if json_file.exists():
            payload = read_cache(json_file)
            if payload:
                meta_url = payload.get("final_url") or payload.get("url")

        items.append(
            {
                "url": meta_url,
                "html_file": str(html_file),
                "text_length": len(text),
                "text": text,
            }
        )
    return items


async def fetch_page_links_rendered(
    browser: Any,
    url: str,
    timeout_ms: int,
    wait_extra_ms: int,
    center_only: bool,
    expand_collapsibles: bool,
) -> tuple[str, str, list[str]]:
    context = await browser.new_context()
    page = await context.new_page()
    try:
        resp = await page.goto(url, wait_until="networkidle", timeout=timeout_ms)
        if resp is None:
            raise RuntimeError(f"Не удалось открыть страницу: {url}")
        if wait_extra_ms > 0:
            await page.wait_for_timeout(wait_extra_ms)

        if expand_collapsibles:
            await page.evaluate(
                """() => {
                    function isInMain(node) {
                      return !!node.closest('main, [role="main"], article, .content, .main, .page-content');
                    }
                    // <details>
                    document.querySelectorAll('details').forEach((el) => {
                      if (!isInMain(el)) return;
                      el.open = true;
                    });
                    // Кнопки/элементы аккордеонов
                    const candidates = Array.from(
                      document.querySelectorAll(
                        '[aria-expanded="false"], .accordion button, .spoiler button, button'
                      )
                    );
                    for (const el of candidates) {
                      if (!isInMain(el)) continue;
                      const txt = (el.innerText || '').toLowerCase();
                      const controls = el.getAttribute('aria-controls') || '';
                      const cls = (el.className || '').toString().toLowerCase();
                      const looksLikeToggle =
                        txt.includes('показать') ||
                        txt.includes('разверн') ||
                        txt.includes('подробн') ||
                        controls.length > 0 ||
                        cls.includes('accordion') ||
                        cls.includes('spoiler') ||
                        el.getAttribute('aria-expanded') === 'false';
                      if (!looksLikeToggle) continue;
                      try { el.click(); } catch (_) {}
                    }
                }"""
            )
            if wait_extra_ms > 0:
                await page.wait_for_timeout(min(wait_extra_ms, 2000))

        final_url = normalize_url(page.url)
        html = await page.content()
        raw_links = await page.evaluate(
            """({ centerOnly }) => {
                const EXCLUDE_SELECTOR = 'header, nav, aside, footer';
                function isExcluded(node) {
                  return !!node.closest(EXCLUDE_SELECTOR);
                }

                function isVisible(el) {
                  const st = window.getComputedStyle(el);
                  if (st.display === 'none' || st.visibility === 'hidden') return false;
                  const r = el.getBoundingClientRect();
                  return r.width > 0 && r.height > 0;
                }

                function pickMainRoot() {
                  const explicit = document.querySelector('main, [role="main"], article');
                  if (explicit && !isExcluded(explicit)) return explicit;

                  const candidates = Array.from(document.querySelectorAll('section, div, article'));
                  const viewportCenterX = window.innerWidth / 2;
                  let best = null;
                  let bestScore = -1;

                  for (const el of candidates) {
                    if (isExcluded(el) || !isVisible(el)) continue;
                    const linksCount = el.querySelectorAll('a[href]').length;
                    if (linksCount < 2) continue;
                    const r = el.getBoundingClientRect();
                    const area = r.width * r.height;
                    if (area < 20000) continue;
                    const centerBias = Math.max(0, 1 - Math.abs((r.left + r.width / 2) - viewportCenterX) / viewportCenterX);
                    const score = area * 0.0005 + linksCount * 3 + centerBias * 40;
                    if (score > bestScore) {
                      bestScore = score;
                      best = el;
                    }
                  }
                  return best || document.body;
                }

                const root = centerOnly ? pickMainRoot() : document.body;
                return Array.from(root.querySelectorAll('a[href]'))
                  .filter(a => !isExcluded(a))
                  .map(a => a.getAttribute('href'))
                  .filter(Boolean);
            }""",
            {"centerOnly": center_only},
        )

        links: set[str] = set()
        for href in raw_links:
            href = str(href).strip()
            if not href or href.startswith(("javascript:", "mailto:", "tel:", "#")):
                continue
            abs_url = normalize_url(urljoin(final_url, href))
            if abs_url.startswith(("http://", "https://")):
                links.add(abs_url)

        return final_url, html, sorted(links)
    finally:
        await context.close()


def read_cache(cache_json: Path) -> dict[str, Any] | None:
    if not cache_json.exists():
        return None
    try:
        return json.loads(cache_json.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return None


def write_cache(
    cache_json: Path,
    cache_html: Path,
    url: str,
    final_url: str,
    html: str,
    links: list[str],
) -> None:
    cache_html.parent.mkdir(parents=True, exist_ok=True)
    cache_json.parent.mkdir(parents=True, exist_ok=True)

    cache_html.write_text(html, encoding="utf-8")
    payload = {
        "url": url,
        "final_url": final_url,
        "fetched_at_unix": int(time.time()),
        "links": links,
        "html_cache_file": str(cache_html),
    }
    cache_json.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")


async def get_links_with_cache(
    browser: Any,
    url: str,
    cache_dir: Path,
    ttl_seconds: int,
    timeout_ms: int,
    wait_extra_ms: int,
    center_only: bool,
    expand_collapsibles: bool,
    verbose: bool,
) -> tuple[str, list[str], bool]:
    mode_salt = f"{url}|center={int(center_only)}|expand={int(expand_collapsibles)}"
    key = cache_key(mode_salt)
    cache_json = cache_dir / f"{key}.json"
    cache_html = cache_dir / f"{key}.html"

    if is_cache_fresh(cache_json, ttl_seconds):
        payload = read_cache(cache_json)
        if payload and isinstance(payload.get("links"), list):
            final_url = normalize_url(payload.get("final_url", url))
            links = sorted({normalize_url(x) for x in payload["links"] if isinstance(x, str)})
            if verbose:
                print(f"[CACHE HIT] {url}", file=sys.stderr)
            return final_url, links, True

        # fallback: попробуем вытащить ссылки из html кэша
        if cache_html.exists():
            html = cache_html.read_text(encoding="utf-8", errors="ignore")
            links = extract_links_from_html(html, url)
            if verbose:
                print(f"[CACHE HIT/HTML PARSE] {url}", file=sys.stderr)
            return normalize_url(url), links, True

    final_url, html, links = await fetch_page_links_rendered(
        browser=browser,
        url=url,
        timeout_ms=timeout_ms,
        wait_extra_ms=wait_extra_ms,
        center_only=center_only,
        expand_collapsibles=expand_collapsibles,
    )
    write_cache(cache_json, cache_html, url, final_url, html, links)
    if verbose:
        print(f"[FETCH] {url}", file=sys.stderr)
    return final_url, links, False


async def crawl_links(
    start_url: str,
    max_depth: int,
    same_domain_only: bool,
    cache_dir: Path,
    ttl_seconds: int,
    timeout_ms: int,
    wait_extra_ms: int,
    delay_ms: int,
    center_only: bool,
    expand_collapsibles: bool,
    interactive_confirm: bool,
    verbose: bool,
) -> dict[str, Any]:
    start_url = normalize_url(start_url)

    visited: set[str] = set()
    queue = deque([(start_url, 0)])
    pages: dict[str, dict[str, Any]] = {}
    all_links: set[str] = set()
    approved_for_crawl: set[str] = set()
    rejected_for_crawl: set[str] = set()

    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        try:
            while queue:
                current_url, depth = queue.popleft()
                if current_url in visited:
                    continue
                visited.add(current_url)

                try:
                    final_url, links, from_cache = await get_links_with_cache(
                        browser=browser,
                        url=current_url,
                        cache_dir=cache_dir,
                        ttl_seconds=ttl_seconds,
                        timeout_ms=timeout_ms,
                        wait_extra_ms=wait_extra_ms,
                        center_only=center_only,
                        expand_collapsibles=expand_collapsibles,
                        verbose=verbose,
                    )
                except Exception as exc:  # pragma: no cover
                    pages[current_url] = {
                        "depth": depth,
                        "final_url": current_url,
                        "links_count": 0,
                        "from_cache": False,
                        "error": str(exc),
                    }
                    continue

                unique_links = sorted(set(links))
                all_links.update(unique_links)
                pages[current_url] = {
                    "depth": depth,
                    "final_url": final_url,
                    "links_count": len(unique_links),
                    "from_cache": from_cache,
                }

                if depth >= max_depth:
                    continue

                links_for_next_depth = unique_links
                if interactive_confirm and depth < max_depth:
                    print(
                        f"\nСтраница: {current_url}\nНайдено ссылок: {len(unique_links)}",
                        file=sys.stderr,
                    )
                    picked: list[str] = []
                    for idx, link in enumerate(unique_links, start=1):
                        if link in approved_for_crawl:
                            picked.append(link)
                            continue
                        if link in rejected_for_crawl:
                            continue
                        answer = input(f"[{idx}/{len(unique_links)}] Парсить дальше? {link} [y/N/a/q]: ").strip().lower()
                        if answer == "q":
                            links_for_next_depth = picked
                            break
                        if answer == "a":
                            approved_for_crawl.add(link)
                            picked.append(link)
                            continue
                        if answer == "y":
                            approved_for_crawl.add(link)
                            picked.append(link)
                        else:
                            rejected_for_crawl.add(link)
                    else:
                        links_for_next_depth = picked

                for link in links_for_next_depth:
                    if same_domain_only and not same_domain(link, start_url):
                        continue
                    if link not in visited:
                        queue.append((link, depth + 1))

                # Даже при обходе из кэша держим равномерный темп.
                if delay_ms > 0:
                    await asyncio.sleep(delay_ms / 1000.0)
        finally:
            await browser.close()

    return {
        "start_url": start_url,
        "max_depth": max_depth,
        "same_domain_only": same_domain_only,
        "center_only": center_only,
        "expand_collapsibles": expand_collapsibles,
        "interactive_confirm": interactive_confirm,
        "ttl_seconds": ttl_seconds,
        "pages_scanned": len(pages),
        "unique_links_count": len(all_links),
        "pages": pages,
        "links": sorted(all_links),
    }


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Скрапинг ссылок с JS-рендером и кэшированием."
    )
    parser.add_argument("--url", required=True, help="Стартовая страница")
    parser.add_argument(
        "--max-depth",
        type=int,
        default=0,
        help="Глубина обхода ссылок: 0 = только стартовая страница",
    )
    parser.add_argument(
        "--same-domain-only",
        action="store_true",
        default=True,
        help="Ограничить обход тем же доменом (по умолчанию включено)",
    )
    parser.add_argument(
        "--allow-external",
        action="store_true",
        help="Разрешить обход внешних доменов (перекрывает --same-domain-only)",
    )
    parser.add_argument(
        "--center-only",
        action="store_true",
        default=True,
        help="Брать ссылки только из основного/центрального контентного блока (по умолчанию включено)",
    )
    parser.add_argument(
        "--all-page-links",
        action="store_true",
        help="Брать ссылки со всей страницы (отключает --center-only)",
    )
    parser.add_argument(
        "--expand-collapsibles",
        action="store_true",
        default=True,
        help="Пытаться развернуть аккордеоны/спойлеры в основном блоке (по умолчанию включено)",
    )
    parser.add_argument(
        "--no-expand-collapsibles",
        action="store_true",
        help="Не пытаться раскрывать выпадающие блоки",
    )
    parser.add_argument(
        "--cache-dir",
        type=Path,
        default=DEFAULT_CACHE_DIR,
        help=f"Директория кэша (default: {DEFAULT_CACHE_DIR})",
    )
    parser.add_argument(
        "--cache-ttl-days",
        type=float,
        default=7.0,
        help="Сколько дней считать кэш свежим",
    )
    parser.add_argument(
        "--timeout-ms",
        type=int,
        default=30000,
        help="Таймаут загрузки страницы",
    )
    parser.add_argument(
        "--wait-extra-ms",
        type=int,
        default=1500,
        help="Доп. ожидание после networkidle для JS-блоков",
    )
    parser.add_argument(
        "--delay-ms",
        type=int,
        default=700,
        help="Пауза между страницами",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=None,
        help="Куда сохранить JSON-результат",
    )
    parser.add_argument(
        "--links-only",
        action="store_true",
        help="Печатать только ссылки (по одной в строке) в stdout",
    )
    parser.add_argument("--verbose", action="store_true", help="Логи в stderr")
    parser.add_argument(
        "--interactive-confirm",
        action="store_true",
        help="Перед обходом вглубь спрашивать по каждой ссылке, идти ли дальше",
    )
    parser.add_argument(
        "--cache-extract-text",
        action="store_true",
        help="Не скрапить сеть, а извлечь текст из уже сохраненных HTML в кэше",
    )
    parser.add_argument(
        "--cache-text-output",
        type=Path,
        default=None,
        help="Куда сохранить тексты из кэша (JSON)",
    )
    parser.add_argument(
        "--cache-text-plain-output",
        type=Path,
        default=None,
        help="Куда сохранить тексты из кэша (plain text)",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    same_domain_only = False if args.allow_external else args.same_domain_only
    center_only = False if args.all_page_links else args.center_only
    expand_collapsibles = False if args.no_expand_collapsibles else args.expand_collapsibles
    ttl_seconds = int(max(0.0, args.cache_ttl_days) * 24 * 60 * 60)
    cache_dir = args.cache_dir.resolve()
    cache_dir.mkdir(parents=True, exist_ok=True)

    if args.cache_extract_text:
        texts = extract_texts_from_cache(cache_dir)
        if args.cache_text_output:
            out_json = args.cache_text_output.resolve()
            out_json.parent.mkdir(parents=True, exist_ok=True)
            out_json.write_text(
                json.dumps(texts, ensure_ascii=False, indent=2),
                encoding="utf-8",
            )
            print(f"Сохранено JSON: {out_json}", file=sys.stderr)

        if args.cache_text_plain_output:
            out_txt = args.cache_text_plain_output.resolve()
            out_txt.parent.mkdir(parents=True, exist_ok=True)
            chunks: list[str] = []
            for idx, item in enumerate(texts, start=1):
                chunks.append(
                    f"{'=' * 70}\n"
                    f"CACHED PAGE #{idx}\n"
                    f"URL: {item.get('url') or '-'}\n"
                    f"HTML: {item.get('html_file')}\n"
                    f"{'=' * 70}\n\n"
                    f"{item.get('text', '')}\n"
                )
            out_txt.write_text("\n".join(chunks), encoding="utf-8")
            print(f"Сохранено TXT: {out_txt}", file=sys.stderr)

        if not args.cache_text_output and not args.cache_text_plain_output:
            print(json.dumps(texts, ensure_ascii=False, indent=2))
        return

    result = asyncio.run(
        crawl_links(
            start_url=args.url,
            max_depth=max(0, args.max_depth),
            same_domain_only=same_domain_only,
            cache_dir=cache_dir,
            ttl_seconds=ttl_seconds,
            timeout_ms=args.timeout_ms,
            wait_extra_ms=max(0, args.wait_extra_ms),
            delay_ms=max(0, args.delay_ms),
            center_only=center_only,
            expand_collapsibles=expand_collapsibles,
            interactive_confirm=args.interactive_confirm,
            verbose=args.verbose,
        )
    )

    if args.output is not None:
        out = args.output.resolve()
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_text(json.dumps(result, ensure_ascii=False, indent=2), encoding="utf-8")
        print(f"Сохранено: {out}", file=sys.stderr)

    if args.links_only:
        for link in result["links"]:
            print(link)
    else:
        print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
