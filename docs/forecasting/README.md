# Stock Forecasting Plan

This document outlines a phased plan to add stock-price forecasting while preserving the current architecture (domain -> service -> repository -> API) and the shared graph-schema contract used by API and `cmd/ask-cypher`.

## Strategy at a Glance

- Use a quantitative baseline model as the primary predictor (not LLM-only prediction).
- Use knowledge graph (KG) and GraphRAG as contextual feature sources and explanation support.
- Keep all forecasting artifacts read-only relative to core supply-chain facts.
- Treat outputs as analytics with uncertainty, never financial advice.

## Scope and Principles

- Forecast only companies with a public market ticker.
- Enforce explicit parent-ticker mapping for subsidiaries (for example, Google DeepMind -> GOOGL).
- Prefer calibrated, backtested models over persuasive but unverified narrative outputs.
- Store forecast outputs and model metadata in Neo4j for API and NL query use.
- Return prediction intervals and quality metadata with every forecast response.

## Phase 0: Product Definition

Goal: agree on forecasting behavior and success criteria before implementation.

Deliverables:

- Supported horizons: 1d, 7d, 30d.
- Initial ticker universe and eligibility rules.
- Parent-child ticker mapping policy.
- Output contract: point forecast, confidence interval, model metadata, freshness metadata.
- Evaluation contract: MAE, RMSE, directional accuracy, and baseline-comparison rules.

Exit criteria:

- Written decisions for horizons, ticker coverage, mapping policy, and evaluation metrics.

## Phase 1: Data Model and Schema

Goal: add graph structures for market history, forecasts, and reliability metadata.

Graph additions:

- Label `StockPrice` with properties: `date`, `close`, `adjusted_close`, `volume`, `source`, `created_at`.
- Label `Forecast` with properties: `as_of`, `horizon_days`, `predicted_close`, `lower`, `upper`, `model`, `mae`, `rmse`, `directional_accuracy`, `version`, `trained_on_from`, `trained_on_to`, `created_at`.
- Optional label `ForecastRun` with properties: `run_id`, `started_at`, `finished_at`, `status`, `error`, `provider`, `model_version`.
- Relationship `(Company)-[:HAS_STOCK_PRICE]->(StockPrice)`.
- Relationship `(Company)-[:HAS_FORECAST]->(Forecast)`.
- Optional relationship `(Forecast)-[:GENERATED_BY]->(ForecastRun)`.
- Optional property on `Company`: `ticker`.

Data constraints and indexes (required):

- Unique or merge-safe key for forecast identity: company + `as_of` + `horizon_days` + `version`.
- Index on `Company.ticker`.
- Index on `StockPrice.date` and lookup path from company to date.

Repository updates:

- Update `internal/graphschema/schema.go` with new labels, relationships, and properties.
- Add repository methods for upsert/read of `StockPrice`, `Forecast`, and optionally `ForecastRun`.

Exit criteria:

- New schema is queryable from Neo4j Browser.
- `cmd/ask-cypher` can reference the new labels/properties via shared schema safely.

## Phase 2: Ingestion and Baseline Training

Goal: create a repeatable, idempotent pipeline that ingests history and writes forecasts.

Implementation:

- Add `cmd/train-forecast` command.
- Pull daily historical prices from one provider.
- Normalize/validate prices (split/dividend-aware via `adjusted_close` when available).
- Train baseline per ticker and horizon (moving average, linear trend, Holt-Winters).
- Select best baseline per ticker/horizon using rolling validation.
- Persist history (`StockPrice`) and outputs (`Forecast`) with model-quality metadata.

Operational behavior:

- Idempotent writes: same identity updates existing forecast, does not duplicate.
- Log per-ticker input size, fit quality, and write counts.
- Guardrails for insufficient history and invalid provider payloads.

Exit criteria:

- One command refreshes end-to-end history and forecasts for all enabled tickers.
- Forecast nodes exist with interval and quality metadata.

## Phase 3: API and Query Experience

Goal: expose forecasts and history to API clients and natural-language querying.

API additions:

- `GET /api/v1/stocks/:name/history?from=&to=`
- `GET /api/v1/forecast/:name?horizon_days=`
- Optional admin endpoint: `POST /api/v1/forecast/run`

Code layout:

- Add forecast domain types under `internal/domain/`.
- Add service methods under `internal/service/`.
- Add repository methods under `internal/repository/`.
- Add handlers/routes under `internal/api/http/`.

`cmd/ask-cypher` support:

- Extend shared schema prompt to include stock/forecast entities.
- Add few-shot examples for safe forecast-related Cypher queries.

Exit criteria:

- API returns forecast, interval, model metadata, and freshness metadata.
- Natural-language forecast questions produce valid, schema-grounded Cypher.

## Phase 4: Evaluation and Reliability

Goal: measure quality, calibrate uncertainty, and operate reliably.

Evaluation:

- Backtest each ticker/horizon on rolling windows.
- Track MAE, RMSE, directional accuracy, and interval coverage.
- Compare champion baseline vs candidate upgrades.
- Run ablations: with and without KG/news-derived features.

Reliability:

- Freshness checks for latest price date and forecast date.
- Alerts for ingestion failures and stale outputs.
- Health surfacing for missing tickers, data gaps, and low-confidence predictions.

Exit criteria:

- Quantitative report exists per ticker/horizon.
- Freshness and health checks are production-ready.

## Phase 5: KG, GraphRAG, and News Enrichment

Goal: improve contextual awareness and risk sensitivity without replacing numeric forecasting.

Design principles:

- LLM is not the sole predictor; it is used for extraction, synthesis, and explanation.
- Convert news into structured, time-stamped graph facts and features.
- Keep extracted facts separate from inferred narratives.

Graph additions (optional but recommended):

- Label `NewsEvent` with properties: `event_id`, `event_time`, `event_type`, `summary`, `source`, `source_quality`, `confidence`, `created_at`.
- Label `Signal` with properties: `as_of`, `signal_type`, `value`, `decay_half_life`, `window_days`, `version`.
- Relationships:
  - `(Company)-[:MENTIONED_IN]->(NewsEvent)`
  - `(Company)-[:HAS_SIGNAL]->(Signal)`
  - Optional causal/context links from `NewsEvent` to impacted companies.

Pipeline behavior:

- Stream news, deduplicate, and score source quality.
- Extract entities/relations/events with LLM + deterministic validators.
- Write only high-confidence structured facts/signals.
- Apply temporal decay to event signals.
- Feed signals as features into forecast training/inference.

Exit criteria:

- News-to-signal pipeline is live and measurable.
- KG/news features show statistically meaningful lift over baseline in backtests.

## Phase 6: Advanced Models (Optional)

Goal: improve forecast quality only when evidence justifies added complexity.

Candidate upgrades:

- ARIMA/Prophet and other robust statistical models.
- Gradient-boosted or hybrid models with market + KG/news features.
- Regime-aware modeling (earnings windows, volatility regimes, macro shock flags).

Exit criteria:

- New model beats baseline/champion consistently on agreed metrics.
- Model remains stable across retraining cycles.

## Suggested Delivery Sequence

1. Phase 0 + Phase 1 in one PR (contract + schema).
2. Phase 2 in one PR (ingest + baseline training command).
3. Phase 3 in one PR (API + ask-cypher support).
4. Phase 4 in one PR (evaluation + reliability).
5. Phase 5 in one PR (KG/news enrichment and ablation evidence).
6. Phase 6 only if metrics justify complexity.

## Risks and Mitigations

- Private-company mismatch: enforce ticker-based eligibility and mapping policy.
- Provider limits/outages: cache last good data and expose freshness state.
- Corporate action distortions: prefer `adjusted_close` for modeling.
- Noisy news stream: dedupe, quality-score sources, confidence-threshold writes.
- Model overconfidence: always return interval bounds and calibration metrics.
- LLM hallucination risk: require schema validation and deterministic extraction checks.
- Scope creep: gate advanced modeling behind measurable lift.
