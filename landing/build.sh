#!/bin/sh
# Builds the deployable landing page (index.html) from src/.
# head.html holds <title>, fonts and styles; body.html holds markup and script.
set -e
cd "$(dirname "$0")"
{
  printf '<!doctype html>\n<html lang="en">\n<head>\n'
  printf '<meta charset="utf-8">\n<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">\n'
  printf '<meta name="description" content="Clearlien checks every loan and invoice behind your facilities against independent sources, every day, and flags assets already pledged to another lender.">\n'
  printf '<meta property="og:title" content="Clearlien: know the collateral is real, and pledged only to you">\n'
  printf '<meta property="og:description" content="Daily, asset-level collateral verification and a cross-lender registry for private credit, banks and factors.">\n'
  printf '<meta property="og:type" content="website">\n<meta name="twitter:card" content="summary">\n'
  printf '<meta name="theme-color" content="#10221D">\n'
  printf '<link rel="icon" href="data:image/svg+xml,%%3Csvg xmlns=%%27http://www.w3.org/2000/svg%%27 viewBox=%%270 0 32 32%%27%%3E%%3Ccircle cx=%%2716%%27 cy=%%2716%%27 r=%%2715%%27 fill=%%27%%2310221D%%27/%%3E%%3Cpath d=%%27M10.5 16.4l3.6 3.6 7.4-8%%27 fill=%%27none%%27 stroke=%%27%%23F3F6F4%%27 stroke-width=%%272.4%%27/%%3E%%3C/svg%%3E">\n'
  cat src/head.html
  printf '</head>\n<body>\n'
  cat src/body.html
  printf '</body>\n</html>\n'
} > index.html
echo "built $(wc -c < index.html) bytes -> landing/index.html"
