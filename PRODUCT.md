# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

People who want to get to know Matteo professionally, understand his work, and potentially buy services from him. This includes professional contacts arriving through NFC contact sharing.

## Product Purpose

mrokk.com introduces Matteo's professional background and work, with an emphasis on cloud and data engineering. It helps visitors assess whether his experience fits their needs and find a way to contact him about working together or buying services.

Success means visitors understand what Matteo does, can examine his work, and can easily start a relevant conversation. No quantitative targets have been established.

## Positioning

This is Matteo's personal professional website. Its substance is his own cloud/data engineering experience and work, connected to direct personal contact channels and NFC contact sharing.

A more specific commercial differentiator or service proposition remains undecided; do not invent one.

## Operating Context

- Visitors can arrive through the website directly or through an NFC interaction used to share contact information.
- The intended journey is to learn about Matteo, explore relevant work and experience, and make contact.
- The existing site separates the introduction (`/`), CV and portfolio (`/me`), blog (`/blog` and individual posts), and public contact links (`/contacts`).
- The user requested an Italian version alongside English. English keeps its existing URLs; Italian pages use `/it/`. A shared IT/EN link switches between equivalent pages, and navigation retains the selected language.
- An existing NFC joke page is served at `/mattrolled`, with `/rickrolled` as an alias. It combines a personal video with contact links. NFC contact information is a confirmed priority; a required NFC destination has not been specified.

## Capabilities and Constraints

Confirmed priorities are cloud/data engineering, selected work, and NFC contact information.

Existing implementation facts, rather than requirements to retain every implementation choice:

- The site uses Go, the chi router, server-rendered HTML templates, CSS, and a small JavaScript file. There is no frontend framework.
- `content.yaml` holds professional content and contact links. Blog posts can use Markdown files in `blog/`.
- Templates and static assets are embedded in the Go binary. Content is loaded at startup; updates require restarting the server, and embedded asset changes require rebuilding it.
- Current contact channels are email, LinkedIn, and Telegram. The public site also includes light/dark themes and a blog RSS feed.
- `/ifconfig` and `/ip` provide a network echo utility; `/healthz` provides a health check.
- Local development uses `go run .`, serving port 8080 by default. Cloud Run deployment configuration is present; this does not establish the current production deployment state.

Open decisions:

- Specific paid service offerings, scope, pricing, availability, and the inquiry or purchasing process.
- Which completed projects should be featured as evidence for prospective clients.
- The NFC destination and whether its current joke/video experience should change.
- Any additional accessibility standard or audience-specific access requirements.

## Brand Commitments

The existing identity is mrokk / Matteo, with mrokk.com as the website domain. The user confirmed an emphasis on professional cloud/data engineering work and NFC contact information. No additional binding voice or visual direction was established during initialization.

## Evidence on Hand

- `content.yaml`: existing biography, employment history, selected work, education, skills, certification links, volunteering, and contact information. Treat these as the site's recorded content; do not expand them into unsupported claims or assume time-sensitive details remain current indefinitely.
- `static/images/ista_enhanced.jpeg`: existing portrait.
- `static/images/favicon.png`: existing site icon.
- `static/videos/Make_me_a_video_of_me_dancing.mp4`: existing personal video for the NFC joke page.
- `blog/lorem.md`: placeholder blog content, not evidence of published technical expertise.
- Existing selected-work entries include a sound-feature recognition project and a general cloud/data exploration entry. Detailed commercial case studies, testimonials, client approvals, quantified outcomes, and paid service terms have not been established in this initialization and must not be fabricated.

## Product Principles

1. Help visitors understand Matteo's cloud/data engineering expertise and whether it fits their needs.
2. Support professional claims with actual experience, work, and verifiable evidence.
3. Make relevant contact straightforward, including for people arriving through NFC.
4. Support service inquiries with truthful information; leave unconfirmed offers and commercial terms undecided.
