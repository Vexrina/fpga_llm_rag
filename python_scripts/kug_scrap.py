#!/usr/bin/env python3
"""
Извлечение текста со сканированных страниц PDF (layoutparser + Detectron2 PubLayNet + Tesseract).

Рекомендуемый запуск для плотных таблиц + структурированных блоков (стр. 29 и 55):

  pyenv activate lp_env
  python extract_scanned_pages.py --pdf pdf_files/kug.pdf --pages 29 55 \\
      --page-mode 29 full --page-mode 55 layout

Страницы — 1-based (как в просмотрщике PDF).
"""
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

import cv2
import layoutparser as lp
import numpy as np
import pytesseract
import torch
from pdf2image import convert_from_path
from pytesseract import Output

# Корень: python_scripts/
ROOT = Path(__file__).resolve().parent
CONFIG = ROOT / "test_scraps" / "faster_rcnn_R_50_FPN_3x-config.yml"
WEIGHTS = ROOT / "models" / "publaynet" / "model_final.pth"
DEFAULT_PDF = ROOT / "pdf_files" / "kug.pdf"

LABEL_MAP = {0: "text", 1: "title", 2: "list", 3: "table", 4: "figure"}
# oem 3: LSTM + legacy; без жёсткой OTSU на всей странице текст читается лучше
TESS_TEXT = "--oem 3 --psm 6 -c preserve_interword_spaces=1"
TESS_TABLE = "--oem 3 --psm 6 -c preserve_interword_spaces=1"
MIN_OCR_DIM = 1600


def _build_extra_config(score_thresh: float) -> list:
    return [
        "MODEL.ROI_HEADS.SCORE_THRESH_TEST",
        score_thresh,
    ]


def load_model(config_path: Path, weights_path: Path, score_thresh: float):
    return lp.Detectron2LayoutModel(
        config_path=str(config_path),
        model_path=str(weights_path),
        extra_config=_build_extra_config(score_thresh),
        label_map=LABEL_MAP,
    )


def pdf_pages_to_pil(pdf_path: Path, page_numbers: list[int], dpi: int):
    """page_numbers — 1-based, как в PDF viewer."""
    out = []
    for p in page_numbers:
        imgs = convert_from_path(
            str(pdf_path), dpi=dpi, first_page=p, last_page=p
        )
        if not imgs:
            raise RuntimeError(f"Страница {p} не получена из PDF")
        out.append(imgs[0])
    return out


def preprocess_for_ocr(img_pil):
    """
    Градации серого + шумоподавление + CLAHE. Без глобальной OTSU — она часто
    «съедает» мелкий текст и линии таблиц на сканах.
    """
    img = np.array(img_pil.convert("RGB"))
    gray = cv2.cvtColor(img, cv2.COLOR_RGB2GRAY)
    denoised = cv2.fastNlMeansDenoising(gray, h=10, templateWindowSize=7, searchWindowSize=21)
    clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8, 8))
    enhanced = clahe.apply(denoised)
    coords = np.column_stack(np.where(enhanced < 250))
    if len(coords) < 10:
        return enhanced
    angle = cv2.minAreaRect(coords)[-1]
    if angle < -45:
        angle = -(90 + angle)
    else:
        angle = -angle
    # Только небольшой реальный перекос скана. minAreaRect по всем тёмным пикселям
    # на широких многоколоночных страницах даёт ~±90° — поворот на такой угол
    # ломает макет и теряет верхние строки (напр. коды 09.04.04 перед «Виды деятельности»).
    if abs(angle) < 0.3 or abs(angle) > 5.0:
        return enhanced
    h, w = enhanced.shape
    m = cv2.getRotationMatrix2D((w // 2, h // 2), angle, 1)
    return cv2.warpAffine(enhanced, m, (w, h), flags=cv2.INTER_CUBIC, borderMode=cv2.BORDER_REPLICATE)


def pad_and_upscale(crop: np.ndarray) -> np.ndarray:
    """Светлая рамка + масштаб для мелкого шрифта."""
    if crop.ndim == 2:
        c = cv2.copyMakeBorder(crop, 12, 12, 12, 12, cv2.BORDER_CONSTANT, value=255)
    else:
        c = crop
    h, w = c.shape[:2]
    m = max(h, w)
    if m >= MIN_OCR_DIM:
        return c
    scale = MIN_OCR_DIM / m
    return cv2.resize(c, None, fx=scale, fy=scale, interpolation=cv2.INTER_CUBIC)


def sort_reading_order(layout):
    """Сверху вниз, при близком Y — слева направо."""
    blocks = list(layout)

    def key(b):
        x1, y1, x2, y2 = b.coordinates
        return (round(y1 / 30) * 30, x1)

    return sorted(blocks, key=key)


def ocr_text_block(img_gray, block) -> str:
    x1, y1, x2, y2 = map(int, block.coordinates)
    h, w = img_gray.shape[:2]
    x1, y1 = max(0, x1), max(0, y1)
    x2, y2 = min(w, x2), min(h, y2)
    if x2 <= x1 or y2 <= y1:
        return ""
    crop = img_gray[y1:y2, x1:x2]
    crop = pad_and_upscale(crop)
    text = pytesseract.image_to_string(
        crop, lang="rus+eng", config=TESS_TEXT
    )
    lines = [ln.strip() for ln in text.splitlines() if ln.strip()]
    return "\n".join(lines)


def ocr_table_block(img_gray, block) -> str:
    """Таблица: строки по line_num Tesseract, ячейки по большим разрывам по X."""
    x1, y1, x2, y2 = map(int, block.coordinates)
    h, w = img_gray.shape[:2]
    x1, y1 = max(0, x1), max(0, y1)
    x2, y2 = min(w, x2), min(h, y2)
    if x2 <= x1 or y2 <= y1:
        return ""
    crop = img_gray[y1:y2, x1:x2]
    crop = pad_and_upscale(crop)
    data = pytesseract.image_to_data(
        crop,
        lang="rus+eng",
        output_type=Output.DICT,
        config=TESS_TABLE,
    )
    rows: dict[int, list[tuple[float, str]]] = {}
    n = len(data["text"])
    for i in range(n):
        t = (data["text"][i] or "").strip()
        if not t:
            continue
        try:
            conf = int(data["conf"][i])
        except (TypeError, ValueError):
            conf = -1
        if conf != -1 and conf < 20:
            continue
        line_id = data["line_num"][i]
        left = data["left"][i]
        cx = left + data["width"][i] / 2.0
        rows.setdefault(line_id, []).append((cx, t))

    if not rows:
        return pytesseract.image_to_string(
            crop, lang="rus+eng", config=TESS_TABLE
        ).strip()

    widths = []
    for i in range(n):
        if not (data["text"][i] or "").strip():
            continue
        try:
            if int(data["conf"][i]) < 20:
                continue
        except (TypeError, ValueError):
            continue
        widths.append(data["width"][i])
    median_w = float(np.median(widths)) if widths else 20.0

    sorted_lines = []
    for line_id in sorted(rows.keys()):
        words = sorted(rows[line_id], key=lambda z: z[0])
        if not words:
            continue
        xs = [w[0] for w in words]
        gaps = [xs[j + 1] - xs[j] for j in range(len(xs) - 1)]
        split_thresh = max(median_w * 2.5, 25.0)
        line_parts: list[str] = []
        buf: list[str] = []
        for j, (_, word) in enumerate(words):
            buf.append(word)
            if j < len(gaps) and gaps[j] > split_thresh:
                line_parts.append(" ".join(buf))
                buf = []
        if buf:
            line_parts.append(" ".join(buf))
        sorted_lines.append(" | ".join(line_parts) if line_parts else " ".join(w[1] for w in words))

    return "\n".join(sorted_lines)


def ocr_figure_block(img_gray, block) -> str:
    return ocr_text_block(img_gray, block)


def clean_line(s: str) -> str:
    s = re.sub(r"[^\S\n]+", " ", s)
    return s.strip()


def ocr_full_page(img_ocr: np.ndarray, psm: int = 3) -> str:
    """Целиком страница: для плотных таблиц, когда разбиение по блокам даёт мусор."""
    cfg = f"--oem 3 --psm {psm} -c preserve_interword_spaces=1"
    return pytesseract.image_to_string(img_ocr, lang="rus+eng", config=cfg).strip()


def parse_page(img_pil, lp_model, img_rgb: np.ndarray, img_ocr: np.ndarray):
    layout = lp_model.detect(img_rgb)
    layout = [b for b in layout if b.width > 40 and b.height > 12]
    layout = sort_reading_order(layout)

    parts: list[str] = []
    for b in layout:
        t = b.type if isinstance(b.type, str) else LABEL_MAP.get(b.type, "text")
        if t in ("text", "title", "list"):
            raw = ocr_text_block(img_ocr, b)
            raw = clean_line(raw)
            if raw:
                prefix = f"## {t}\n" if t in ("title",) else ""
                parts.append(prefix + raw)
        elif t == "table":
            raw = ocr_table_block(img_ocr, b)
            raw = "\n".join(clean_line(x) for x in raw.splitlines() if x.strip())
            if raw:
                parts.append("### table\n" + raw)
        elif t == "figure":
            raw = ocr_figure_block(img_ocr, b)
            raw = clean_line(raw)
            if raw:
                parts.append("### figure\n" + raw)

    return "\n\n".join(parts)


def main():
    ap = argparse.ArgumentParser(description="OCR страниц PDF через layout + Tesseract")
    ap.add_argument(
        "--pdf",
        type=Path,
        default=DEFAULT_PDF,
        help="Путь к PDF",
    )
    ap.add_argument(
        "--pages",
        type=int,
        nargs="+",
        default=[29, 55],
        help="Номера страниц (1-based)",
    )
    ap.add_argument(
        "--dpi",
        type=int,
        default=400,
        help="DPI рендера страницы (выше — лучше мелкий текст, дольше)",
    )
    ap.add_argument(
        "--out",
        "-o",
        type=Path,
        default=None,
        help="Файл для сохранения (UTF-8). По умолчанию: extracted_pages_<n>.txt рядом с PDF",
    )
    ap.add_argument(
        "--score-thresh",
        type=float,
        default=0.35,
        help="Порог уверенности Detectron2 (ниже — больше регионов)",
    )
    ap.add_argument(
        "--strategy",
        choices=("layout", "full"),
        default="layout",
        help="layout: PubLayNet + OCR по блокам; full: только Tesseract на всю страницу (плотные таблицы)",
    )
    ap.add_argument(
        "--full-psm",
        type=int,
        default=3,
        help="При --strategy full: PSM Tesseract (3 — авто, 4 — одна колонка)",
    )
    ap.add_argument(
        "--append-full-page",
        action="store_true",
        help="После layout добавить полный OCR страницы для сравнения",
    )
    ap.add_argument(
        "--page-mode",
        action="append",
        nargs=2,
        metavar=("PAGE", "MODE"),
        default=[],
        help="Переопределить режим для страницы: MODE = layout или full. Пример: --page-mode 29 full --page-mode 55 layout",
    )
    args = ap.parse_args()

    pdf_path = args.pdf.resolve()
    if not pdf_path.is_file():
        print(f"Файл не найден: {pdf_path}", file=sys.stderr)
        sys.exit(1)
    if not CONFIG.is_file():
        print(f"Конфиг не найден: {CONFIG}", file=sys.stderr)
        sys.exit(1)
    if not WEIGHTS.is_file():
        print(f"Веса не найдены: {WEIGHTS}", file=sys.stderr)
        sys.exit(1)

    page_modes: dict[int, str] = {}
    for p, m in args.page_mode:
        page_modes[int(p)] = m
        if m not in ("layout", "full"):
            print(f"Недопустимый MODE: {m} (нужно layout или full)", file=sys.stderr)
            sys.exit(1)

    need_layout = args.strategy == "layout" or any(
        page_modes.get(p, args.strategy) == "layout" for p in args.pages
    )

    print(f"Устройство: {'cuda' if torch.cuda.is_available() else 'cpu'}", file=sys.stderr)
    lp_model = None
    if need_layout:
        lp_model = load_model(CONFIG, WEIGHTS, args.score_thresh)

    images = pdf_pages_to_pil(pdf_path, args.pages, args.dpi)

    chunks = []
    for page_no, img_pil in zip(args.pages, images):
        img_rgb = np.array(img_pil.convert("RGB"))
        img_ocr = preprocess_for_ocr(img_pil)
        strat = page_modes.get(page_no, args.strategy)
        if strat == "full":
            body = ocr_full_page(img_ocr, psm=args.full_psm)
            body = "\n".join(clean_line(x) for x in body.splitlines() if x.strip())
        else:
            body = parse_page(img_pil, lp_model, img_rgb, img_ocr)
        if args.append_full_page and strat == "layout":
            extra = ocr_full_page(img_ocr, psm=args.full_psm)
            extra = "\n".join(clean_line(x) for x in extra.splitlines() if x.strip())
            body = (
                body
                + "\n\n---\n### полный OCR страницы (дополнительно)\n\n"
                + extra
            )
        chunks.append(f"{'=' * 72}\nСТРАНИЦА {page_no}\n{'=' * 72}\n\n{body}")

    text = "\n\n".join(chunks) + "\n"

    out_path = args.out
    if out_path is None:
        out_path = pdf_path.parent / f"extracted_pages_{'_'.join(map(str, args.pages))}.txt"
    out_path.write_text(text, encoding="utf-8")
    print(text)
    print(f"\nСохранено: {out_path}", file=sys.stderr)


if __name__ == "__main__":
    main()
