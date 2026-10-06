# Clearlien landing page

`index.html` is the deployable early-access page. It is one self-contained file with no build tools needed to host it.

- Edit `src/head.html` (title, fonts, styles) and `src/body.html` (content, form, script).
- Run `./build.sh` to regenerate `index.html`.

## Put it live (about 10 minutes)

**Option A: Netlify (sign-ups work with no extra setup)**

1. Create a free Netlify account, then drag the `landing/` folder onto app.netlify.com/drop.
2. In Site settings, go to Forms and enable form detection, then redeploy. Sign-ups from the `early-access` form show up under Forms. Turn on email notifications there.
3. Add your domain under Domain management.

**Option B: any static host (Vercel, Cloudflare Pages, GitHub Pages) plus Formspree**

1. Create a form at formspree.io and copy its endpoint, for example `https://formspree.io/f/abcdwxyz`.
2. In `src/body.html`, set `var CONFIG = { endpoint: "https://formspree.io/f/abcdwxyz" };` and run `./build.sh`.
3. Deploy `index.html`.

## What each sign-up records

- Name, work email, firm, role, institution type, collateral types, collateral size band, and whether they want a design-partner check.
- `utm_source`, `utm_medium`, `utm_campaign`, `utm_content`, `utm_term`. This is the first set of UTM tags the visitor arrived with, remembered across visits, so you can see which ad or post produced each lead.
- `referrer`.
- A hidden `website` field that traps spam bots.

## Before you run ads

- Buy the domain and run a trademark search for the name (USPTO trademark search).
- Add a short privacy policy page and link it in the footer. LinkedIn Lead Gen Forms require one, and so does GDPR for UK and EU visitors.
- Add privacy-friendly analytics (Plausible or Fathom) by pasting their script tag into the `<head>` in `build.sh`.
- Replace the sample figures in the hero report only if you have real ones. They are labeled as sample data today.

The claude.ai artifact link is a preview only. On the preview, the form shows a confirmation without sending anything.
