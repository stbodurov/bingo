# Singleton — Threat-model-first research

> Method: instead of starting from a market-size slide, we started from the **adversary**.
> Who is attacking online platforms, what does an attack cost them, what does it cost the
> victim, and why does every current defense either fail or leak private data? The product
> spec falls out of the answers. Research date: October 2026.

---

## 1. The adversary got cheaper than the defender

| Signal | Data point | Source |
|---|---|---|
| Machines are now the majority of the web | Automated requests reached **57.5%** of HTML traffic vs 42.5% human (Cloudflare Radar, 3 Jun 2026) — first time in internet history | [Tom's Hardware](https://www.tomshardware.com/tech-industry/artificial-intelligence/bots-have-now-passed-human-traffic-online-cloudflare-boss-laments-says-agentic-traffic-wasnt-expected-to-eclipse-real-people-until-next-year), [TechTimes](https://www.techtimes.com/articles/317877/20260605/bot-traffic-passes-humans-online-cloudflare-says-agentic-ai-drove-575-share.htm) |
| Agent traffic is compounding | AI-agent requests on Cloudflare grew **>1,700% YoY**; human traffic fell up to **40%** in retail, software, IT and financial services | [Forbes](https://www.forbes.com/sites/jonmarkman/2026/10/01/cloudflare-says-ai-agent-traffic-grew-more-than-1700-in-the-past-year/), [Cloudflare blog](https://blog.cloudflare.com/agentic-internet-bot-report/) |
| Independent confirmation | Imperva 2026 Bad Bot Report: bots **53%** of all traffic in 2025 (51% in 2024) | [digitalapplied summary](https://www.digitalapplied.com/blog/ai-crawler-bot-traffic-statistics-2026-data-reference) |
| CAPTCHAs are solved | ETH Zürich solved **100%** of reCAPTCHA v2 challenges with an off-the-shelf YOLO model; humans and bots needed the same number of challenges | [Decrypt](https://decrypt.co/251107/ai-can-best-googles-bot-detection-system-swiss-researchers-find), [arXiv 2409.08831](https://rivista.ai/wp-content/uploads/2024/09/2409.08831v1.pdf) |
| Deepfakes defeat "look at the camera" | An Arup employee wired ~**$25.5M** after a video call with deepfaked executives | [World ID enterprise coverage](https://www.biometricupdate.com/202606/world-shifts-from-crypto-identity-experiment-to-enterprise-proof-of-humanity) |

**Implication:** any defense that asks "can you do something a computer can't?" (puzzles,
behavioural signals, liveness selfies) is on the wrong side of the cost curve. Generative
models get better at imitation every quarter; the defender's test does not.

## 2. What the attacks cost victims

| Harm | Size | Source |
|---|---|---|
| All US consumer fraud reported to FTC, 2025 | **$15.9B**, up from $12.5B | [FTC via Central Oregon Daily](https://www.centraloregondaily.com/news/consumer/ftc-romance-scams-1-billion-losses-2025/article_c32c7fc5-c3a9-4cdc-8f4b-c80293080267.html) |
| Social-media-originated scams | **8×** 2020 levels | [FTC data spotlight, Apr 2026](https://www.ftc.gov/news-events/data-visualizations/data-spotlight/2026/04/reported-losses-scams-social-media-eight-times-higher-2020) |
| Romance scams (first 9 months of 2025) | **$1.16B**, 55,604 reports, +22% YoY; ~60% started on social media | [FTC via Central Oregon Daily](https://www.centraloregondaily.com/news/consumer/ftc-romance-scams-1-billion-losses-2025/article_c32c7fc5-c3a9-4cdc-8f4b-c80293080267.html) |

Every one of these begins with an account that a human did not need to stand behind:
fake profiles, sock puppets, bot farms, recycled identities.

## 3. Regulators now *require* platforms to know something about users

| Jurisdiction | Obligation | Teeth | Source |
|---|---|---|---|
| Australia | Under-16s may not hold accounts on major social platforms (in force **10 Dec 2025**). Government ID **cannot be the only option** — a reasonable alternative must exist | Up to **AUD 49.5M** per serious breach; 4.7M under-16 accounts removed by mid-Dec 2025; investigations into Meta, Snap, TikTok, YouTube opened Mar 2026 | [eSafety](https://www.esafety.gov.au/about-us/industry-regulation/social-media-age-restrictions), [DLA Piper](https://privacymatters.dlapiper.com/2026/02/australias-social-media-ban-and-the-esafety-commissioners-social-media-minimum-age-regulatory-guidance/), [Wikipedia](https://en.wikipedia.org/wiki/Online_Safety_Amendment) |
| United Kingdom | Online Safety Act "highly effective age assurance" since **25 Jul 2025**; guidance asks platforms to confirm age **without collecting or storing personal data** unless necessary | Up to **10% of worldwide revenue**; 90+ services investigated, fines £80k–£1.35M so far | [Ofcom summaries](https://www.zyphe.com/resources/blog/age-verification-uk-online-safety-act), [cybernews](https://cybernews.com/uk/internet-censorship/get-around-uk-age-verification-requirements-2026/) |
| European Union | Every member state must offer an EU Digital Identity Wallet by end-2026 | — | [idscan](https://idscan.net/blog/the-state-of-digital-ids-in-2026-a-complete-guide/) |
| United States | 22 states/territories have TSA-accepted mobile IDs; Apple Wallet IDs live in 14–17 jurisdictions and expanding | — | [idscan](https://idscan.net/blog/the-state-of-digital-ids-in-2026-a-complete-guide/), [MacRumors](https://www.macrumors.com/2026/08/17/apple-wallet-ids-expanding-to-four-more-states/) |

**Implication:** "Are you human?" and "are you old enough?" are becoming the same
compliance question, asked on every signup, in every major market.

## 4. Today's defenses and why each one fails

| Defense | Failure mode | Evidence |
|---|---|---|
| CAPTCHA / behavioural scoring | Solved by vision models; punishes humans with friction; scores rely on cookies and browsing history (a privacy cost) | ETH Zürich study above |
| Phone/SMS verification | Numbers are rented by the thousand; SIM farms; excludes people without stable numbers | Industry common knowledge; no single authoritative dataset |
| Document + selfie KYC ($1.35–$2.30 per check) | Too expensive for every signup; creates honeypots of ID images; deepfake injection | [Sumsub pricing](https://sumsub.com/de/pricing2/), [Didit pricing review](https://didit.me/pt-PT/blog/new-pricing-november-2026/) |
| **The honeypot problem (the big one)** | Discord's support vendor breach exposed **~70,000 government ID photos** from users who verified their age (hackers claimed 2.1M). The Tea dating-safety app left **72,000 images** including photo IDs in an open bucket | [NBC News](https://www.nbcnews.com/tech/tech-news/70000-government-id-photos-exposed-discord-user-hack-rcna236714), [TechSpot](https://www.techspot.com/news/109793-hackers-they-stole-21-million-discord-photos-latest.html) |
| Iris/palm biometrics (World, Humanity Protocol) | Requires special hardware or a body-part scan; regulators have repeatedly acted against biometric collection; excludes people with eye/hand conditions | [Biometric Update](https://www.biometricupdate.com/202606/world-shifts-from-crypto-identity-experiment-to-enterprise-proof-of-humanity), [proofoftalk](https://proofoftalk.io/blog/proof-of-personhood-identity-agentic-age/) |

The EFF summed up the trap: *"Age verification systems are surveillance systems."*
Platforms are forced to choose between **compliance** and **not becoming a breach headline**.

## 5. Demand is proven — and so is the valuation

- **World** (Tools for Humanity): 18M+ verified humans, 450M+ verifications, integrations with
  Tinder, Zoom, Docusign, Okta, Vercel; now **charging apps per verification**; World Foundation
  raised **$525M**. ([Business Wire](https://www.businesswire.com/news/home/20260417530721/en/The-New-World-ID-Proof-of-Human-for-the-AI-Era-Scales-Across-the-Digital-Platforms-People-and-Businesses-Use-Every-Day), [HackerNoon](https://hackernoon.com/world-foundation-raises-$525-million-to-scale-proof-of-human-as-ai-transforms-the-internet))
- **Humanity Protocol** (palm scans + ZK): **$1.1B** valuation on a $20M round (Pantera, Jump). ([Reuters via StreetInsider](https://www.streetinsider.com/Reuters/Humanity+Protocol+valued+at+%241.1+billion+after+latest+fundraise/24249164.html))
- **YC Fall 2026 RFS — "Proving You're Human"** (Max Kolysh) explicitly asks for this company.
  ([superframeworks RFS tracker](https://superframeworks.com/articles/yc-fall-2026-rfs-indie-hacker-plays))
- **a16z crypto 2026**: "Know Your Agent" and privacy named as the long-term moats.
  ([Yahoo Finance](https://finance.yahoo.com/news/vc-firm-a16z-flags-stablecoins-102355858.html))

Two unicorn-scale players exist and **both require a body scan**. The market leader's own
critics point to the gap: *no hardware, no biometrics, no document honeypot*.

## 6. Threat model (formal)

### Assets
1. A platform's **one-account-per-human** invariant.
2. A platform's **age-gate** compliance evidence.
3. The **user's real-world identity** and cross-site activity (must never leak).

### Adversaries
| ID | Adversary | Capability | Goal |
|---|---|---|---|
| A1 | Bot farm operator | Unlimited compute, LLMs, CAPTCHA solvers, rented phones | Mass fake accounts |
| A2 | Sybil human | One real identity, wants N accounts (ban evasion, vote stuffing, airdrop/ticket farming) | Duplicate humans per app |
| A3 | Under-age user | Real but too young; may borrow a parent's document | Bypass age gate |
| A4 | Curious platform | Sees every proof presented to it; colludes with other platforms | De-anonymise / link users across apps |
| A5 | Breach attacker | Steals Singleton's or a platform's database | Harvest identities |
| A6 | Malicious/compromised attester | Can issue false attestations | Mint fake humans |
| A7 | Singleton itself (insider) | Operates the registry | Track users |

### Security goals → design decisions
| Goal | Design decision | Defeats |
|---|---|---|
| G1: each human holds at most one credential | Enrollment de-duplication on **keyed hashes** (HMAC-pepper) of document numbers and of name+DOB+country — never the raw values | A1, A2 |
| G2: one human ⇒ one stable pseudonym per app | **Scoped key image** `I = x·H(scope)`; the proof binds `I` to a valid member key | A2 |
| G3: apps cannot link users across apps | Different scopes ⇒ independent `H(scope)` ⇒ unlinkable `I` (DDH assumption on ristretto255) | A4 |
| G4: proofs reveal nothing beyond the predicate | **Linkable ring signature** over an anonymity bucket; no name, DOB, document, or key is disclosed | A4, A7 |
| G5: nothing worth stealing | Singleton stores public keys, keyed hashes and threshold dates — no images, no documents, no DOB | A5 |
| G6: attester misbehaviour is visible and reversible | Public attester registry, per-attester assurance levels, fraud rates; revocation of every key an attester minted; epoch roots anchored on-chain | A6, A7 |
| G7: age without birthdays | Store only *the date a key becomes 16/18/21*; sets are recomputed each epoch | A3 (with mDL/passport issuer signatures), A4 |
| G8: humans can deputise AI agents | Delegation statements signed **inside the same scope** — the app learns "this agent acts for pseudonym I", not who I is | Agent impersonation |

### Residual risks we accept (and disclose)
- **Bucket anonymity, not set anonymity (v1).** Ring signatures are O(n); we sign over a
  hash-prefix *bucket* of ≥64 members (stable across epochs, so intersection attacks reveal at
  most the bucket). v2 replaces the ring with a zk-SNARK (Semaphore-style) proof over the whole set.
- **Document lending.** A teenager can borrow a parent's phone. Mitigation: device-bound keys,
  liveness at enrollment for low-assurance paths, and one-credential-per-document (the parent
  cannot then enrol themselves without revoking the child's).
- **Insider at enrollment.** The registry sees which public key was enrolled with which keyed
  hash. It never sees proofs' signer (ring anonymity), and keyed hashes are useless without the
  HSM-held pepper.

## 7. What this research told us to build

1. **An acceptance network, not another biometric.** Accept the credentials people already have —
   mobile driver's licences (ISO 18013-5), ePassport chips (ICAO 9303), EU wallets, in-person
   attestation, vouching — and turn them into one reusable, private proof.
2. **One API for platforms** that answers three questions: *human? unique here? old enough?*
3. **A consumer pass** that holds the key, never the document.
4. **Public transparency** (anchored roots, attester registry, revocations) so trust does not
   depend on believing us.
5. **Agent delegation** so the agentic web has a way to say "a real person stands behind this bot".
