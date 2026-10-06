#!/bin/sh
# Builds the deployable landing page (index.html) from src/.
# head.html holds <title>, fonts and styles; body.html holds markup and script.
set -e
cd "$(dirname "$0")"
{
  printf '<!doctype html>\n<html lang="en">\n<head>\n'
  printf '<meta charset="utf-8">\n<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">\n'
  printf '<meta name="description" content="Applicants prove income, hours and balances straight from the payroll portal, gig app or bank that holds them. Cryptographic proofs, checked against your rules by AI.">\n'
  printf '<meta property="og:title" content="Fontis: stop reading documents, get the proof">\n'
  printf '<meta property="og:description" content="Income and hours verification from the source, for the applicants payroll databases cannot verify.">\n'
  printf '<meta property="og:type" content="website">\n<meta name="twitter:card" content="summary">\n'
  printf '<meta name="theme-color" content="#0D1A2C">\n'
  printf '<link rel="icon" href="data:image/svg+xml,%%3Csvg xmlns=%%27http://www.w3.org/2000/svg%%27 viewBox=%%270 0 28 28%%27%%3E%%3Crect width=%%2728%%27 height=%%2728%%27 rx=%%276%%27 fill=%%27%%230D1A2C%%27/%%3E%%3Ccircle cx=%%2714%%27 cy=%%2719%%27 r=%%272.6%%27 fill=%%27%%2384A6FF%%27/%%3E%%3Cpath d=%%27M8.5 15.5a7.5 7.5 0 0 1 11 0M5 11.5a12.5 12.5 0 0 1 18 0%%27 fill=%%27none%%27 stroke=%%27%%23F1F4F8%%27 stroke-width=%%272%%27 stroke-linecap=%%27round%%27/%%3E%%3C/svg%%3E">\n'
  cat src/head.html
  printf '</head>\n<body>\n'
  cat src/body.html
  printf '</body>\n</html>\n'
} > index.html
echo "built $(wc -c < index.html) bytes -> fontis/index.html"
