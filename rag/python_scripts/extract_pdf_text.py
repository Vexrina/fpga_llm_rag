#!/usr/bin/env python3
"""Simple PDF text extraction using pdfplumber."""

import argparse
import sys
import os

import pdfplumber


def extract_text(pdf_path: str, pages: str) -> str:
    print(f"DEBUG: Opening PDF at {pdf_path}", file=sys.stderr)
    print(f"DEBUG: PDF file exists: {os.path.exists(pdf_path)}", file=sys.stderr)
    print(f"DEBUG: PDF file size: {os.path.getsize(pdf_path)}", file=sys.stderr)
    
    with pdfplumber.open(pdf_path) as pdf:
        print(f"DEBUG: PDF has {len(pdf.pages)} pages", file=sys.stderr)
        
        if pages == "1":
            page_nums = list(range(1, len(pdf.pages) + 1))
        else:
            page_nums = [int(p) for p in pages.split(",")]
        
        print(f"DEBUG: Extracting pages: {page_nums}", file=sys.stderr)
        
        texts = []
        for page_num in page_nums:
            if 1 <= page_num <= len(pdf.pages):
                page = pdf.pages[page_num - 1]
                text = page.extract_text()
                if text:
                    texts.append(text)
        return "\n\n".join(texts)


def main():
    parser = argparse.ArgumentParser(description="Extract text from PDF")
    parser.add_argument("--pdf", required=True, help="Path to PDF file")
    parser.add_argument("--pages", default="1", help="Page numbers (comma-separated) or '1' for all")
    parser.add_argument("-o", "--output", required=True, help="Output file")
    
    args = parser.parse_args()
    
    try:
        text = extract_text(args.pdf, args.pages)
        with open(args.output, "w", encoding="utf-8") as f:
            f.write(text)
        print(f"Extracted text to {args.output}")
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()