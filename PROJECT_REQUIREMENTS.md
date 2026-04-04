# Albion Online Profit Calculator - Project Requirements

## Overview
A two-layer system to help earn silver in Albion Online through refining (Phase 1), with potential expansion to crafting, farming, and mount raising.

---

## Architecture

### Layer 1: Data Collection (albiondata-client)
- **Language:** Go (existing project)
- **Location:** `c:\dev\albiondata-client`
- **Modifications needed:**
  - Add PostgreSQL uploader to save market data locally
  - Remove public server uploads (NATS/HTTP to public servers disabled)
  - Add WebSocket server to notify analysis layer when new data arrives
  - Capture ALL market data: all items, all tiers (T2-T8), all enchantments (.0-.3), all cities

### Layer 2: Analysis & Calculator
- **Language:** Python 3.11.9
- **Framework:** FastAPI (backend) + simple HTML/CSS/JS (frontend)
- **Location:** `c:\dev\albion-profit-calculator` (separate project)
- **Run:** Directly on PC (not Docker)

---

## Database

- **Type:** PostgreSQL
- **Deployment:** Docker container (local)
- **Schema:** Latest price only (no history for now)
- **Future:** Migrateable to cloud (Supabase/Neon) when needed

### Data Stored
- Item ID
- City/Market location
- Buy order price (highest)
- Sell order price (lowest)
- Timestamp captured
- Daily volume (history selling amount per day)

---

## Communication

- **albiondata-client → Database:** Direct PostgreSQL insert
- **albiondata-client → Analysis Layer:** WebSocket push notification on new data
- **Analysis Layer → Database:** PostgreSQL queries
- **Analysis Layer → User:** Prompt to check in-game market if data > 2 hours old

---

## Refining Calculator Features

### Core Calculations
- [ ] Profit per item (single unit margin)
- [ ] Profit per focus point (efficiency metric)
- [ ] Daily profit potential (based on available focus)
- [ ] Cost breakdown (raw materials, taxes, fees)
- [ ] Tax calculations (setup fee, sales tax)

### City Optimization
- [ ] Best city to refine (considering city bonuses)
- [ ] Best city to sell (considering portal costs)
- [ ] Resource return rate by city
- [ ] Portal cost calculations (city-to-city)

### User Inputs
- [ ] Spec levels for each refining skill
- [ ] Available focus points
- [ ] Current city location

### Data Management
- [ ] Data freshness indicator per item/city
- [ ] Staleness threshold: **2 hours**
- [ ] "Check this market" suggestions in web UI when data > 2 hours old
- [ ] No automatic fallback to public API — prompt user to view item in-game instead

### Display
- [ ] Buy/sell order suggestions
- [ ] Volume indicators (daily selling amount)
- [ ] Comparison tables across cities

---

## Materials to Track (Refining)

### Raw Resources
- Ore (T2-T8, .0-.3)
- Hide (T2-T8, .0-.3)
- Fiber (T2-T8, .0-.3)
- Wood (T2-T8, .0-.3)
- Stone (T2-T8, .0-.3)

### Refined Materials
- Metal Bars (T2-T8, .0-.3)
- Leather (T2-T8, .0-.3)
- Cloth (T2-T8, .0-.3)
- Planks (T2-T8, .0-.3)
- Stone Blocks (T2-T8, .0-.3)

### Markets
- Caerleon
- Bridgewatch
- Fort Sterling
- Lymhurst
- Martlock
- Thetford
- Brecilien

---

## City Bonuses (Refining)

| City | Bonus Material |
|------|----------------|
| Fort Sterling | Hide → Leather |
| Lymhurst | Wood → Planks |
| Bridgewatch | Stone → Blocks |
| Martlock | Ore → Bars |
| Thetford | Fiber → Cloth |
| Caerleon | No bonus (but central) |
| Brecilien | No bonus |

---

## Travel Costs (Portal Network)

- Only city portals considered (safe travel)
- Fixed silver costs between cities
- Factor into "best sell city" calculation

---

## Tech Stack Summary

| Component | Technology |
|-----------|------------|
| Data Collector | Go (albiondata-client) |
| Database | PostgreSQL (Docker) |
| Backend API | Python FastAPI |
| Frontend | HTML/CSS/JavaScript |
| Real-time | WebSocket |
| Notifications | Web browser only |

---

## Setup Requirements

### Prerequisites
- [ ] Docker Desktop installed
- [ ] Python 3.11.9 installed
- [ ] Go environment (for building albiondata-client)

### Docker Services
```yaml
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: albion
      POSTGRES_USER: albion
      POSTGRES_PASSWORD: <secure_password>
    ports:
      - "5432:5432"
    volumes:
      - albion_data:/var/lib/postgresql/data
```

---

## Data Flow

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────────────┐
│  Albion Online  │────▶│ albiondata-  │────▶│    PostgreSQL       │
│  (Game Client)  │     │ client (Go)  │     │    (Docker)         │
└─────────────────┘     └──────┬───────┘     └──────────┬──────────┘
                               │                        │
                               │ WebSocket              │ Query
                               ▼                        ▼
                        ┌──────────────────────────────────────┐
                        │      albion-profit-calculator        │
                        │         (Python FastAPI)             │
                        └──────────────────┬───────────────────┘
                                           │
                                           ▼
                        ┌──────────────────────────────────────┐
                        │          Web Browser UI              │
                        │    (localhost - HTML/CSS/JS)         │
                        └──────────────────────────────────────┘
                                           │
                                           │ If data > 2hr old
                                           ▼
                        ┌──────────────────────────────────────┐
                        │   "Please check [item] in [city]     │
                        │    market for fresh price"           │
                        └──────────────────────────────────────┘
```

---

## Future Expansion (Phase 2+)

- [ ] Crafting profit calculator
- [ ] Farming profit calculator
- [ ] Mount raising profit calculator
- [ ] Price history tracking
- [ ] Cloud database migration
- [ ] Mobile-friendly web UI
- [ ] Share data with friends

---

## Configuration

### albiondata-client config additions
```yaml
database:
  host: localhost
  port: 5432
  name: albion
  user: albion
  password: <secure_password>

websocket:
  enabled: true
  port: 8080

public_upload:
  enabled: false  # Disabled - local only
```

### Analysis layer config
```yaml
database:
  connection_string: postgresql://albion:password@localhost:5432/albion

websocket:
  albiondata_client: ws://localhost:8080

staleness:
  threshold_hours: 2
  action: prompt_user  # Show message to check market in-game

server:
  host: localhost
  port: 3000
```

---

*Document created: April 3, 2026*
*Status: Requirements gathering complete - Ready for planning*
