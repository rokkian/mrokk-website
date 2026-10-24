---
name: mrokk — Gravity-rain garden
description: A quiet painted garden with rounded lettering and a plain text identity.
colors:
  garden: "#263e37"
  cloud: "#e7e8df"
  soft-cloud: "#c2cfc3"
  rain-blue: "#9bc6ff"
  rain-coral: "#ffa49a"
  stem: "#6a8272"
  control-ink: "#233c34"
  selection: "#475e7a"
  day-garden: "#e5ede2"
  day-cloud: "#233c34"
  day-soft-cloud: "#445f50"
  day-rain-blue: "#295d98"
  day-rain-coral: "#a34134"
  day-stem: "#889d8a"
  day-control-ink: "#e5ede2"
  day-selection: "#bed6dc"
typography:
  display:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "clamp(68px, 7vw, 96px)"
    fontWeight: 400
    lineHeight: 1.05
    letterSpacing: "-.04em"
  title:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "clamp(23px, 2.5vw, 30px)"
    fontWeight: 500
    lineHeight: 1.35
    letterSpacing: "-.025em"
  body:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "18px"
    fontWeight: 500
    lineHeight: 1.75
  label:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "16px"
    fontWeight: 600
    lineHeight: 1.4
  navigation:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "15px"
    fontWeight: 500
  note:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "14px"
    fontWeight: 500
    lineHeight: 1.8
  caption:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "12px"
    fontWeight: 500
  reading-title:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "64px"
    fontWeight: 400
    lineHeight: 1.05
    letterSpacing: "-.04em"
  article-title:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "56px"
    fontWeight: 400
    lineHeight: 1.15
    letterSpacing: "-.04em"
  profile-title:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "48px"
    fontWeight: 400
    lineHeight: 1.05
    letterSpacing: "-.04em"
  reading-copy:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "15px"
    fontWeight: 500
    lineHeight: 1.75
  row-title:
    fontFamily: '"Quicksand", sans-serif'
    fontSize: "23px"
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: "-.025em"
rounded:
  pill: "40px"
  motion: "24px"
  circle: "50%"
  skip: "8px"
  media: "16px"
spacing:
  tight: "12px"
  gutter: "24px"
  open: "32px"
  content-start: "32px"
components:
  destination-primary:
    backgroundColor: "{colors.cloud}"
    textColor: "{colors.control-ink}"
    typography: "{typography.label}"
    rounded: "{rounded.pill}"
    padding: "15px 22px"
  destination-primary-hover:
    backgroundColor: "{colors.rain-blue}"
    textColor: "{colors.garden}"
  destination-secondary:
    backgroundColor: "{colors.garden}"
    textColor: "{colors.cloud}"
    typography: "{typography.label}"
    rounded: "{rounded.pill}"
    padding: "15px 22px"
  destination-secondary-hover:
    backgroundColor: "{colors.rain-coral}"
    textColor: "{colors.garden}"
  theme-control:
    textColor: "{colors.cloud}"
    rounded: "{rounded.circle}"
    padding: "11px"
    size: "44px"
  theme-control-hover:
    backgroundColor: "{colors.cloud}"
    textColor: "{colors.garden}"
  motion-control:
    backgroundColor: "{colors.garden}"
    textColor: "{colors.soft-cloud}"
    typography: "{typography.caption}"
    rounded: "{rounded.motion}"
    padding: "10px 12px"
  motion-control-hover:
    textColor: "{colors.cloud}"
  navigation:
    textColor: "{colors.cloud}"
    typography: "{typography.navigation}"
    padding: "12px 0"
  contact-email:
    backgroundColor: "{colors.cloud}"
    textColor: "{colors.control-ink}"
    rounded: "{rounded.media}"
    padding: "28px 32px"
  skip-link:
    backgroundColor: "{colors.cloud}"
    textColor: "{colors.garden}"
    rounded: "{rounded.skip}"
    padding: "12px 18px"
---

# Design System: mrokk — Gravity-rain garden

## Overview

**Creative North Star: "Gravity-rain garden"**

A quiet painted garden pairs dusk-green ground and cloud-cream lettering with gentle blue and coral accents. Rounded typography, a plain text identity, and the original painted artwork give the interface a warm, personal character. The daylight variant preserves the same forms and materials while reversing the surrounding contrast.

This system applies to the whole public website: homepage, CV, blog index and articles, contacts, and NFC page. The user explicitly requested a uniform style across these sections. `static/site.css` owns shared tokens, typography, header, footer, focus and theme controls; `templates/partials.html` supplies their shared markup. `static/home.css` and `static/garden.js` own the homepage scene, while `static/styles.css` owns the reading layouts. Homepage composition belongs to [.impeccable/surfaces/templates-home-html.md](.impeccable/surfaces/templates-home-html.md); internal reading layouts belong to [.impeccable/surfaces/templates-index-html.md](.impeccable/surfaces/templates-index-html.md).

**Key Characteristics:**

- Painted atmosphere with calm, readable interface elements.
- One rounded type family with an airy display and compact controls.
- Flat surfaces, fine strokes, and pill-shaped destinations.
- Dusk and daylight palettes with equivalent semantic roles.
- Restrained optional motion, visible keyboard focus, and working direct links.

## Colors

The palette uses muted garden neutrals and two small, clear accents; the artwork supplies its own painted colors.

### Primary

- **Rain blue** (`rain-blue`): keyboard focus and the primary destination hover.

### Secondary

- **Rain coral** (`rain-coral`): the secondary destination hover. It is an interface accent as well as a color present in the scene.

### Neutral

- **Garden** (`garden`): the page ground and quiet control surfaces.
- **Cloud** (`cloud`): primary lettering and the filled destination surface.
- **Soft cloud** (`soft-cloud`): supporting prose, notes, footer, and motion-control lettering.
- **Stem** (`stem`): fine control outlines, footer divider, and scrollbar thumb.
- **Control ink** (`control-ink`): lettering on the cloud-filled destination.
- **Selection** (`selection`): selected-text ground, paired with cloud lettering.

The unsuffixed frontmatter tokens are the default dusk palette and map directly to the CSS custom properties with the same names. Each `day-` token replaces its corresponding custom property under `data-theme="light"`. Component references describe dusk; use the same semantic substitution for daylight, including hover and focus states.

**The Same Materials Rule.** Theme changes replace semantic palette values while preserving geometry, typography, and artwork.

## Typography

**Display Font:** Quicksand, with sans-serif fallback.

**Body Font:** Quicksand, with sans-serif fallback.

**Character:** Soft, rounded lettering carries both the large personal name and practical interface text. The self-hosted variable Latin font supports weights 300–700; the shipped hierarchy uses regular display lettering and firmer supporting text. Font synthesis is disabled.

### Hierarchy

- **Display:** the `display` role, with tight tracking and a generous name scale.
- **Title:** the `title` role, used for the professional role line.
- **Body:** the `body` role, muted and comfortably spaced; desktop prose is limited to 470px, reduced to 440px at the intermediate breakpoint.
- **Label:** the `label` role for substantial destination links.
- **Navigation:** the `navigation` role for the compact header links.
- **Note / Caption:** the `note` role for supporting evidence; the `caption` role for footer and motion-control text.

At 700px and below, display becomes `clamp(64px, 17vw, 88px)`, the title is 23px, body is 16px with 1.8 line height, destination labels are 14px, navigation is 13px, notes are 12px, and footer text is 11px. At 370px and below, the title becomes 21px and navigation links become 12px. These are responsive overrides of the recorded roles, not a second type system.

## Layout

The homepage uses a flex column with a 1600px maximum width and at least one small viewport height. Horizontal inset grows with `clamp(24px, 5.5vw, 88px)`. The repeated gutter and open spacing tokens separate header items, content, actions, and footer; spacing stays generous around readable text and tighter inside controls.

The desktop header is 80px tall. Each page starts its content with the shared `--content-start` gap of 32px, reduced to 20px on mobile. The wordmark, introduction, destination links, and footer share one left edge. Main content and the decorative scene share an isolated stacking context; the scene remains behind the interface and ignores pointer events. At 1050px and below, the text area widens. At 700px and below, the page uses 24px side insets, a 72px header, full-width introduction, and a scene in normal flow below it. All three navigation links remain visible at this size, including at 320px. At 370px and below, side insets become 20px and destination links stack. A 1500px minimum-width rule enlarges the desktop scene and main area.

The source supports a minimum viewport width of 320px. Exact scene positioning and cloud coordinates are artwork implementation details, not reusable layout tokens.

Internal pages share the homepage’s outer insets and left alignment, capped at 1180px. The compact CV uses a 280px profile column and flexible work-history column with a 64px gutter. At 1100px these become 240px and 40px; at 700px the profile and work history stack. A 64px portrait sits beside the 48px personal name. The profile remains sticky only on desktop viewports wider than 1100px and taller than 1000px; otherwise it scrolls with the page so the expanded certification list stays reachable. Contacts use a split composition: personal introduction and portrait on the left, a prominent email destination and two social links on the right; at 700px these stack. The blog index pairs 150px metadata with the article headline and a functional arrow; article back navigation and metadata share one compact toolbar. The NFC page pairs the 16:9 video with contact text and links, stacking on phones. Long prose stays within 65–68ch. Internal copy is 15–17px with 1.7–1.8 leading, while page titles are 64px (48px on phones), article titles 56px (40px on phones), and section titles 30px (26px on phones).

## Elevation & Depth

The interface uses no box shadows. Depth comes from the painted image, tonal contrast, and overlap within the isolated scene. Interface surfaces remain flat; a fine stem-colored border distinguishes the outlined destination and theme control. The footer has a matching hairline divider.

**The Quiet Interface Rule.** Keep readable text and controls still while optional motion stays within the decorative scene.

## Shapes

Destination links use broad pill corners. The theme control is circular, the motion control is a smaller pill, and the focus-revealed skip link has modest rounding. Hand-painted silhouettes belong to the artwork; the name and wordmark stay unadorned. Functional icons are inline stroked SVGs, with stroke weight adjusted to their size; do not replace them with font glyphs.

## Components

### Destination links

Substantial, rounded links pair text with a fine arrow. The filled variant uses cloud and control ink; its hover uses rain blue and garden. The outlined variant uses garden and cloud with a stem border; its hover uses rain coral and garden, including the border. Both have a 58px minimum height on desktop, a 54px minimum height at the mobile breakpoint, and the frontmatter padding and label roles. Mobile padding becomes `14px 18px`. Color transitions take 160ms with `ease-out`; there is no lift or scale state.

### Navigation and footer links

Header links have 44px minimum height and underline on hover. Footer links use soft cloud at rest, then cloud and an underline on hover. The wordmark uses plain rounded lettering. Navigation spacing contracts on mobile rather than becoming a menu overlay.

The header includes a plain text IT/EN link, set at 12px with weight 600 and the same 44px minimum height. It opens the equivalent page in the other language; its `lang`, `hreflang`, accessible label and title name that destination. Language is encoded in page URLs, with English on the original paths and Italian under `/it/`. Shared templates localize visible copy, screen-reader labels, theme controls and rain controls. The page `lang`, canonical URL and alternate-language metadata match the content. The same theme and visual tokens apply to both languages; layouts allow Italian text to wrap, including the longer contact heading.

The browser favicon uses the user-supplied ivory `m` on deep green from `static/images/favicon-new.jpeg`. Shared head markup serves 16px, 32px, and 48px PNG variants plus a 180px Apple touch icon on every page. It accompanies the shared visual identity on every route.

### Reading sections and links

The CV uses native details/summary controls: the latest role starts expanded, while earlier roles show role, company, dates and location until opened. About, skills and credentials are compact profile disclosures. Certifications start expanded by default, with a 48px user-supplied badge beside each credential; each optional image path is configured by `certificates[].icon` in `content.yaml`. Badge images preserve their aspect ratio and have empty alt text because the adjacent credential name supplies the link label. The controls retain keyboard activation and work without JavaScript; print CSS reveals all descriptions. Descriptions render Markdown as semantic paragraphs and lists without changing professional facts. Summary focus uses the same rain-blue treatment as navigation.

Contacts give email a cloud-filled destination with control-ink lettering and a 16px radius. Its hover uses rain blue and garden lettering. Social channels remain open text links below it. The contact portrait/name link leads to the CV. Other reading content uses fine stem dividers, soft-cloud supporting text, blue actions and plain wrapping tags. A blue underline and `aria-current="page"` mark the current primary navigation destination. Long article content supports headings, lists, links, code, blockquotes and horizontal scrolling for tables or code blocks. Portraits and videos keep 16px corners with no tilt, decorative frame or shadow.

### Theme control

A 44px circular outlined button exchanges sun and moon SVGs. Hover reverses the cloud/garden contrast. It is initially hidden until JavaScript initializes it. Dusk is the default; a saved `light` value restores daylight. Storage access failures leave the toggle functional for the current page. The pressed state and accessible label update with the theme.

### Motion control and painted scene

The small text-and-icon pill exposes pause/resume with an updated pressed state and accessible label. It is shown only after the canvas is available and is hidden for reduced motion. The artwork remains a single native-alpha WebP; its user-supplied JPEG and the accepted droplet sprite retain provenance under `.impeccable/assets/`.

Only the blue cloud emits six animated painted drops; the coral cloud and fox remain still. Pointer movement gently steers the rain; touch pointer movement is ignored. The two destination links steer it left and right on both pointer entry and keyboard focus. Pause freezes the current state. Reduced motion draws a still state with neutral steering and disables CSS transitions. Hidden documents stop the animation loop. The sidecar records observed timing and interpolation; it does not promote image-relative emitter coordinates into global tokens.

### Keyboard focus and skip link

All links and buttons receive a rain-blue outline (2px, with 6px offset) on `:focus-visible`. A cloud-filled skip link moves into view when focused. Real anchors keep navigation and destinations working without JavaScript; script-dependent controls stay hidden in that state.

## Do's and Don'ts

### Do:

- **Do** use the shared tokens, navigation, footer, and theme behavior across all public pages.
- **Do** preserve semantic contrast in both dusk and daylight, including control hover and focus states.
- **Do** keep interface text and controls still while optional movement stays in the painted scene.
- **Do** retain visible keyboard focus, direct anchors, reduced-motion stillness, and pause/resume behavior.
- **Do** preserve the user-approved artwork and raster provenance when preparing future variants.

### Don't:

- **Don't** reintroduce separate page palettes, fonts, or theme implementations.
- **Don't** turn illustration coordinates, homepage copy, or destination strategy into global visual tokens.
- **Don't** add static rain to the prepared illustration or animate the coral cloud and fox; the accepted blue-cloud rain is the scene's moving layer.
- **Don't** add shadows or moving text to this flat, quiet interface without an approved visual change.
- **Don't** add decorative circles, orbit marks, or dots beside the name or wordmark; the user explicitly rejected these identity ornaments.
- **Don't** invent cards, inputs, chips, or additional component variants as if they were part of the shipped system.
