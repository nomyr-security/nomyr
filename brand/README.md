# Nomyr brand kit

Everything here is generated from one source: the Aperture mark (a block N with a circular cut through the diagonal and a lit core). SVGs are the masters; the PNGs are rendered from them at the sizes each platform asks for.

## What's in the box

```
svg/       masters — use these wherever vector is accepted
web/       favicons, touch icons, PWA manifest, OG image, <head> snippet
social/    avatars, banners and post images sized per platform
png/       transparent marks and lockups for docs, decks, press
```

### svg/

| File | Use |
|---|---|
| `nomyr-mark.svg` | The mark in ink (#0C3B43) for light backgrounds |
| `nomyr-mark-on-dark.svg` | Off-white ink (#ECF3F1) for dark backgrounds |
| `nomyr-mark-white.svg` | Pure white ink — photos, saturated grounds |
| `nomyr-mark-mono.svg` | Single-colour version (core is ink, not green) — embossing, one-colour print, monochrome UI |
| `nomyr-lockup.svg` / `nomyr-lockup-on-dark.svg` | Mark + wordmark. Text is outlined; no font dependency |
| `nomyr-app-icon.svg` | Rounded-square app icon with the teal face |

### web/

Drop the files into your site root and paste `head-snippet.html` into `<head>`. It covers the classic `.ico` (16/32/48 inside one file), the SVG favicon modern browsers prefer, the iOS touch icon, the Safari pinned-tab mask, the PWA manifest with a maskable icon, and the Open Graph / X card tags. Change the `og:image` URL to wherever you host `og-image.png`.

`favicon.svg` is the small-size variant: below 24 px the aperture cut is dropped and only the green core remains, so it still reads at 16 px. `apple-touch-icon.png` and the `android-chrome-*` files are square-cornered on purpose — iOS and Android apply their own masks. `maskable-icon-512x512.png` keeps the mark inside the 80 % safe zone so it survives circular masks.

### social/

| File | Platform |
|---|---|
| `avatar-dark-400.png`, `avatar-dark-1024.png` | Profile picture, dark. Mark sits inside the circle-safe zone. Use the 400 for X, 1024 elsewhere |
| `avatar-light-400.png`, `avatar-light-1024.png` | Same, light ground (for feeds where a dark disc looks heavy) |
| `x-header-1500x500.png` | X / Twitter header |
| `linkedin-company-cover-1128x191.png` | LinkedIn company page cover |
| `linkedin-personal-banner-1584x396.png` | LinkedIn personal profile banner |
| `linkedin-post-1200x627.png` | LinkedIn link/post image |
| `github-social-preview-1280x640.png` | GitHub repo → Settings → Social preview |
| `og-image-1200x630.png` | Generic link preview (same as `web/og-image.png`) |
| `post-square-1080.png` | Square post — Instagram, Mastodon, Bluesky, Slack announcements |

Banner content is set to the right because the profile picture overlaps the left edge on X and LinkedIn.

### png/

Transparent renders at 256 / 512 / 1024 (mark) and 1600 px wide (lockup). `nomyr-lockup-on-dark-1600-solid.png` is the dark lockup on its own #0E2429 ground for places that flatten transparency (some CMSs, email).

## Colours

| Token | Hex | Use |
|---|---|---|
| Ink | `#0C3B43` | Mark on light, headings, primary buttons |
| Accent | `#07C983` | The core. Never recolour it |
| On dark | `#ECF3F1` | Mark and text on dark grounds |
| Dark | `#0E2429` | Dark surfaces |
| Deep | `#081A1E` | Page background for dark marketing |
| Cyan | `#3FB5C4` | Texture lines and secondary signals only |
| Accent text | `#45DDA2` | Green text on dark (passes contrast; plain accent doesn't) |

Type: IBM Plex Sans (UI, headings), IBM Plex Mono (labels, code), Instrument Serif italic (the one emphasised phrase in headlines). All are open-source on Google Fonts.

## Rules of the mark

- Clear space: one stem width (11/64 of the mark's height) on every side.
- Minimum size: 16 px. Below 24 px use the small-size variant (no cut, core only) — `web/favicon.svg` is that variant.
- Do not recolour the core, rotate the mark, add a drop shadow, outline it, or set it in a container other than the supplied app icon.
- On busy photography use `nomyr-mark-white.svg` or place it on a solid ink/deep panel.
- The wordmark is IBM Plex Sans 700 at −3.5 % tracking. Don't retype it — use the lockup files.
- Tagline: "Every machine identity you have, and the human who answers for it." The italic phrase is always the second half.

## Regenerating

The kit is produced by a Node script that renders each asset with headless Chromium from the same `logo.js` that drives the product UI. If the mark changes there, re-run the script and every file here updates together.
