# Four startups

Four unrelated companies, each built from scratch. Each has its own:
- research method and investment case;
- business model and acquisition plan;
- stack, landing page and waitlist.

None of them shares code, design or approach with another.

| | Category | The problem | Customer | Stack | Run |
|---|---|---|---|---|---|
| [**Singleton**](singleton/) | Web3 identity | Bots outnumber humans online, and age-assurance laws push apps toward ID uploads that get breached. Singleton is proof-of-human and age with linkable ring signatures: a stable pseudonym per app, unlinkable across apps, and nothing else revealed. | Apps and platforms (per verification) | Go, ristretto255, WebAssembly wallet, SQLite, Solidity | `cd singleton && go run ./cmd/singleton` → :4201 |
| [**Harbormaster**](harbormaster/) | AI for trade | Tariffs change by the month and refunds have hard deadlines. Harbormaster audits every customs entry, recovers overpaid duty and prices each SKU's landed cost. | Importers, from SMB to mid-market (SaaS plus a share of recovered refunds) | Python, FastAPI, SQLAlchemy, htmx, Claude | `cd harbormaster && .venv/bin/uvicorn harbormaster.main:app --port 4202` |
| [**Sundial**](sundial/) | Fintech / retirement income | 4.1M Americans turn 65 every year with savings but no paycheck. Sundial is an actuarial engine for Social Security timing, withdrawal order, Roth conversions and annuities, turned into a monthly paycheck. | Retirees and near-retirees (subscription, plus AUM-capped managed accounts) | Java 21, Spring Boot, Flyway, React + TypeScript | `cd sundial && gradle bootRun` → :4203 |
| [**Kept**](kept/) | Consumer health (wildcard) | About 1 in 9 US adults is on a GLP-1. Two-thirds stop within a year, most lost weight returns, and up to a third of the loss can be muscle. Kept handles dose, side effects, protein and strength, cost and the off-ramp, and shows employers whether the spend is working. | Members (freemium) and employers/health plans (per covered life) | Ruby 3.3, Rails 8, Turbo/Stimulus, SQLite | `cd kept && bin/rails db:prepare && bin/rails s -p 4204` |

Each project's own README explains its demo accounts, tests, configuration and Docker image.

Each project's `docs/` folder holds its research, investment case, and revenue and ads plan, each written in a format that suits that company:
- **Singleton:** threat model and protocol spec.
- **Harbormaster:** customer discovery and PRD.
- **Sundial:** actuarial research and investment memo.
- **Kept:** press release and FAQ, plus clinical evidence.
