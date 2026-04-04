# Albion Profit Calculator - Development Plan

## Overview
Small incremental steps, ~1 hour per milestone. Complete Phase 1 (albiondata-client) before Phase 2 (calculator).

---

## Phase 1: albiondata-client Modifications

### Milestone 1.1: Environment Setup - Docker & PostgreSQL
**Goal:** Get PostgreSQL running locally in Docker
**Time:** ~45 min

- [ ] Download and install Docker Desktop for Windows
- [ ] Verify Docker is running (`docker --version`)
- [ ] Create `docker-compose.local.yml` for PostgreSQL
- [ ] Start PostgreSQL container
- [ ] Test connection using a database tool (pgAdmin or command line)
- [ ] Create initial database schema (market_prices table)

**Verification:** Can connect to PostgreSQL and see empty table

---

### Milestone 1.2: Go Environment Setup
**Goal:** Set up Go and verify you can build the existing client
**Time:** ~30-45 min

- [ ] Download and install Go (latest stable, 1.22+)
- [ ] Verify Go is installed (`go version`)
- [ ] Set up Go environment variables if needed
- [ ] Navigate to `c:\dev\albiondata-client`
- [ ] Run `go mod download` to fetch dependencies
- [ ] Build the client (`go build`)
- [ ] Run the client and verify it starts (systray icon appears)

**Verification:** Client builds and runs, shows systray icon

---

### Milestone 1.3: Understand the Codebase
**Goal:** Map out where changes need to be made
**Time:** ~30 min

- [ ] Read through key files:
  - `client/uploader.go` - existing upload interface
  - `client/uploader_http.go` - HTTP uploader implementation
  - `client/uploader_nats.go` - NATS uploader implementation
  - `client/dispatcher.go` - where uploads are triggered
  - `client/config.go` - configuration handling
- [ ] Identify the upload interface pattern
- [ ] Note where market data structures are defined
- [ ] Create notes on modification points

**Verification:** Understand where to add PostgreSQL uploader

---

### Milestone 1.4: Create PostgreSQL Uploader - Basic Structure
**Goal:** Add new uploader file that compiles (no functionality yet)
**Time:** ~45 min

- [ ] Create feature branch: `git checkout -b feature/local-database`
- [ ] Add `lib/pq` PostgreSQL driver to `go.mod`
- [ ] Create `client/uploader_postgres.go` with stub implementation
- [ ] Implement the Uploader interface with empty methods
- [ ] Verify project still builds
- [ ] Commit: "Add PostgreSQL uploader skeleton"

**Verification:** Project compiles with new file

---

### Milestone 1.5: Add Database Configuration
**Goal:** Add PostgreSQL config options
**Time:** ~30-45 min

- [ ] Modify `client/config.go` to add database settings
- [ ] Update `config.yaml.example` with database section
- [ ] Create your local `config.yaml` with database credentials
- [ ] Load and parse database config on startup
- [ ] Add connection string builder
- [ ] Commit: "Add database configuration"

**Verification:** Config loads without errors

---

### Milestone 1.6: Implement PostgreSQL Connection
**Goal:** Connect to database on startup
**Time:** ~45 min

- [ ] Import PostgreSQL driver in uploader_postgres.go
- [ ] Add connection logic in initialization
- [ ] Test connection on startup
- [ ] Handle connection errors gracefully
- [ ] Add reconnection logic (if connection drops)
- [ ] Commit: "Implement PostgreSQL connection"

**Verification:** Client connects to PostgreSQL on startup (check logs)

---

### Milestone 1.7: Implement Market Data Insert
**Goal:** Save market offer data to database
**Time:** ~1 hour

- [ ] Study existing market data structures (`lib/market.go`)
- [ ] Write INSERT/UPSERT query for market_prices table
- [ ] Implement `SendToDatabase()` method
- [ ] Handle the "latest price only" logic (UPSERT)
- [ ] Add timestamp for data freshness
- [ ] Run client, browse market in-game
- [ ] Verify data appears in database
- [ ] Commit: "Implement market data insert"

**Verification:** Browse market in-game → data appears in PostgreSQL

---

### Milestone 1.8: Implement Market History Insert
**Goal:** Save daily volume/history data
**Time:** ~45 min

- [ ] Study market history structures (`lib/marketHistory.go`)
- [ ] Create/update schema for history volume data
- [ ] Implement history data insert
- [ ] Test by viewing item history in-game
- [ ] Verify history data saved
- [ ] Commit: "Implement market history insert"

**Verification:** View item history in-game → volume data in database

---

### Milestone 1.9: Add WebSocket Server - Basic Setup
**Goal:** Start a WebSocket server alongside the client
**Time:** ~45 min

- [ ] Add `gorilla/websocket` dependency (already in vendor)
- [ ] Create `client/ws_server.go` for WebSocket server
- [ ] Start WebSocket server on configurable port
- [ ] Handle basic client connections
- [ ] Add to config: websocket port setting
- [ ] Commit: "Add WebSocket server skeleton"

**Verification:** WebSocket server starts, can connect with browser console

---

### Milestone 1.10: WebSocket Notifications
**Goal:** Push notifications when new data arrives
**Time:** ~45 min

- [ ] Define notification message format (JSON)
- [ ] Hook into dispatcher to detect new data
- [ ] Broadcast to all connected WebSocket clients
- [ ] Include: item_id, city, timestamp, data_type
- [ ] Test with simple WebSocket client
- [ ] Commit: "Implement WebSocket notifications"

**Verification:** Browse market → WebSocket message received

---

### Milestone 1.11: Disable Public Uploads
**Goal:** Make public uploads optional/disabled by default
**Time:** ~30 min

- [ ] Add `public_upload.enabled` config option
- [ ] Modify dispatcher to check this setting
- [ ] Skip HTTP/NATS uploaders when disabled
- [ ] Set default to `false` in config
- [ ] Test: verify no data sent to public servers
- [ ] Commit: "Make public uploads optional"

**Verification:** Network traffic shows no calls to public servers

---

### Milestone 1.12: Final Testing & Polish
**Goal:** End-to-end testing of all modifications
**Time:** ~1 hour

- [ ] Clean build from scratch
- [ ] Full test: start client → browse markets → check database → check WebSocket
- [ ] Test error scenarios (DB down, reconnect)
- [ ] Review all commits
- [ ] Update README with new configuration options
- [ ] Merge to main branch (or keep on feature branch)
- [ ] Tag release: v1.0.0-local

**Verification:** All features work together reliably

---

## Phase 1 Complete Checklist

- [ ] PostgreSQL running in Docker
- [ ] Client saves market prices to local database
- [ ] Client saves daily volume data
- [ ] WebSocket pushes notifications on new data
- [ ] Public uploads disabled
- [ ] Configuration documented
- [ ] Feature branch ready

---

## Phase 2: albion-profit-calculator

**Location:** `c:\dev\albion-profit-calculator`
**Stack:** Python 3.11.9 + FastAPI + HTML/CSS/JS

---

### Milestone 2.1: Python Project Setup
**Goal:** Create project structure with dependencies
**Time:** ~30-45 min

- [ ] Create directory `c:\dev\albion-profit-calculator`
- [ ] Initialize git repository
- [ ] Create virtual environment (`python -m venv venv`)
- [ ] Create `requirements.txt`:
  - fastapi
  - uvicorn
  - psycopg2-binary (PostgreSQL driver)
  - websockets
  - pydantic
  - python-dotenv
- [ ] Install dependencies (`pip install -r requirements.txt`)
- [ ] Create basic folder structure:
  ```
  albion-profit-calculator/
  ├── app/
  │   ├── __init__.py
  │   ├── main.py
  │   ├── config.py
  │   ├── database.py
  │   ├── models/
  │   ├── routers/
  │   ├── services/
  │   └── static/
  ├── requirements.txt
  ├── .env
  └── README.md
  ```
- [ ] Create basic FastAPI app in `main.py`
- [ ] Verify server runs (`uvicorn app.main:app --reload`)
- [ ] Commit: "Initial project setup"

**Verification:** Visit `http://localhost:8000` → see "Hello World" or docs page

---

### Milestone 2.2: Database Connection & Models
**Goal:** Connect to PostgreSQL, create data models
**Time:** ~45 min

- [ ] Create `app/config.py` for environment variables
- [ ] Create `.env` file with database credentials
- [ ] Create `app/database.py` with connection pool
- [ ] Create `app/models/market.py`:
  - MarketPrice model (matches market_prices table)
  - MarketHistory model (matches market_history table)
- [ ] Create `app/models/items.py`:
  - Item definitions (raw resources, refined materials)
  - Tier/enchantment enums
- [ ] Test database connection on startup
- [ ] Commit: "Add database connection and models"

**Verification:** Server starts and logs "Connected to database"

---

### Milestone 2.3: Basic API Endpoints
**Goal:** Create endpoints to query market data
**Time:** ~45 min

- [ ] Create `app/routers/market.py`
- [ ] Endpoints:
  - `GET /api/prices/{item_id}` - get price for item (all cities)
  - `GET /api/prices/{item_id}/{city}` - get price for item in city
  - `GET /api/history/{item_id}` - get volume history
  - `GET /api/stale` - get list of stale items (>2 hours)
- [ ] Add data freshness field to responses
- [ ] Test endpoints with browser/curl
- [ ] Commit: "Add market data API endpoints"

**Verification:** API returns data from database

---

### Milestone 2.4: WebSocket Client - Receive Notifications
**Goal:** Connect to albiondata-client WebSocket
**Time:** ~45 min

- [ ] Create `app/services/ws_client.py`
- [ ] Connect to albiondata-client WebSocket on startup
- [ ] Parse incoming notifications (new market data)
- [ ] Log received notifications
- [ ] Handle reconnection if connection drops
- [ ] Run as background task
- [ ] Commit: "Add WebSocket client for real-time updates"

**Verification:** Browse market in-game → see log message in calculator server

---

### Milestone 2.5: Basic Web UI Structure
**Goal:** Create HTML/CSS/JS foundation
**Time:** ~1 hour

- [ ] Create `app/static/` folder structure:
  ```
  static/
  ├── index.html
  ├── css/
  │   └── style.css
  └── js/
      └── app.js
  ```
- [ ] Set up FastAPI static file serving
- [ ] Create basic layout:
  - Header with title
  - Sidebar for navigation/inputs
  - Main area for tables/results
- [ ] Add basic CSS styling (clean, readable)
- [ ] Set up JavaScript fetch wrapper for API calls
- [ ] Commit: "Add basic web UI structure"

**Verification:** Visit `http://localhost:8000` → see styled page layout

---

### Milestone 2.6: Refining Data Models
**Goal:** Define raw→refined material mappings
**Time:** ~45 min

- [ ] Create `app/data/refining.py`:
  - Raw resource item IDs (ore, hide, fiber, wood, stone)
  - Refined material item IDs (bars, leather, cloth, planks, blocks)
  - Raw → Refined mapping per tier
  - Input quantities per tier (e.g., T5 bar = 2x T5 ore + 1x T4 bar)
- [ ] Create `app/data/cities.py`:
  - City names and IDs
  - City bonus mappings
  - Portal costs between cities
- [ ] Create helper functions to look up mappings
- [ ] Commit: "Add refining data models"

**Verification:** Unit test or manual check that mappings are correct

---

### Milestone 2.7: Resource Return Rate Calculator
**Goal:** Calculate return rates based on spec/focus/city
**Time:** ~45 min

- [ ] Create `app/services/return_rate.py`
- [ ] Implement base return rate formula:
  - Base rate by spec level
  - Focus bonus multiplier
  - City bonus multiplier
- [ ] Create function: `calculate_return_rate(spec, use_focus, city, material_type)`
- [ ] Return effective materials returned per craft
- [ ] Add unit tests for known values
- [ ] Commit: "Add resource return rate calculator"

**Verification:** Test with known spec/focus values → matches in-game

---

### Milestone 2.8: Profit Calculation Service
**Goal:** Calculate profit for refining operations
**Time:** ~1 hour

- [ ] Create `app/services/calculator.py`
- [ ] Implement profit calculation:
  - Input costs (raw materials from market)
  - Output value (refined materials to market)
  - Resource return bonus (materials saved)
  - Tax/fees (setup fee, sales tax)
  - Net profit per item
  - Profit per focus point
- [ ] Create function: `calculate_refining_profit(item, city, spec, focus, prices)`
- [ ] Handle missing price data (mark as "needs refresh")
- [ ] Commit: "Add profit calculation service"

**Verification:** Manual calculation matches service output

---

### Milestone 2.9: User Settings Storage
**Goal:** Save user's spec levels and preferences
**Time:** ~30-45 min

- [ ] Create `app/models/user_settings.py`
- [ ] Store in local file or database:
  - Spec levels for each refining skill
  - Available focus points
  - Current city location
  - Preferences (sort order, filters)
- [ ] Create API endpoints:
  - `GET /api/settings` - get current settings
  - `POST /api/settings` - update settings
- [ ] Add settings panel to UI
- [ ] Commit: "Add user settings storage"

**Verification:** Change settings → refresh page → settings persist

---

### Milestone 2.10: Portal Cost Calculator
**Goal:** Factor travel costs into recommendations
**Time:** ~30 min

- [ ] Create `app/data/portals.py`:
  - Portal cost matrix (city-to-city costs)
- [ ] Create function: `get_portal_cost(from_city, to_city)`
- [ ] Integrate into profit calculation:
  - Calculate net profit after travel
  - Find optimal sell city considering portal costs
- [ ] Commit: "Add portal cost calculations"

**Verification:** Recommendations change based on starting city

---

### Milestone 2.11: Profit Display UI
**Goal:** Show profit calculations in web UI
**Time:** ~1 hour

- [ ] Create profit table component:
  - Columns: Item, Buy Cost, Sell Price, Return Rate, Taxes, Net Profit, Profit/Focus
  - Sortable columns
  - Color coding (green=profitable, red=loss)
- [ ] Add city comparison view:
  - Show same item across all cities
  - Highlight best refining city
  - Highlight best selling city
- [ ] Add filters:
  - By material type
  - By tier
  - By enchantment level
  - Hide unprofitable
- [ ] Commit: "Add profit display UI"

**Verification:** See formatted profit table with real data

---

### Milestone 2.12: Stale Data Detection & Prompts
**Goal:** Identify old data and prompt user
**Time:** ~45 min

- [ ] Create `app/services/freshness.py`:
  - Check timestamp vs 2-hour threshold
  - Identify items needed for calculation but stale
- [ ] Add "Data Status" panel to UI:
  - List of items with stale data
  - Suggested markets to check
  - "Last updated" timestamp per item
- [ ] Highlight stale data in profit tables (yellow/warning)
- [ ] Smart suggestions: prioritize items most impacting profit
- [ ] Commit: "Add stale data detection and prompts"

**Verification:** Items older than 2 hours show warning + suggestion

---

### Milestone 2.13: Real-time UI Updates
**Goal:** Update UI when new data arrives via WebSocket
**Time:** ~45 min

- [ ] Add WebSocket endpoint in FastAPI for browser
- [ ] Create `js/websocket.js` for browser connection
- [ ] When new data arrives:
  - Update relevant rows in profit table
  - Remove item from "stale" list
  - Show brief notification toast
- [ ] Recalculate profits automatically
- [ ] Commit: "Add real-time UI updates"

**Verification:** Browse market in-game → UI updates without refresh

---

### Milestone 2.14: Final Testing & Polish
**Goal:** End-to-end testing, bug fixes, polish
**Time:** ~1 hour

- [ ] Full workflow test:
  1. Start albiondata-client
  2. Start calculator server
  3. Open browser UI
  4. Browse markets in-game
  5. See real-time updates
  6. Use profit calculations
- [ ] Fix any bugs found
- [ ] Improve UI styling/UX
- [ ] Add loading states
- [ ] Error handling for edge cases
- [ ] Create README with setup instructions
- [ ] Commit: "Final polish and documentation"

**Verification:** Full workflow works smoothly

---

## Phase 2 Complete Checklist

- [ ] FastAPI server running
- [ ] Connected to PostgreSQL database
- [ ] Receiving WebSocket notifications from albiondata-client
- [ ] All refining materials mapped
- [ ] Profit calculations accurate
- [ ] City bonuses applied correctly
- [ ] Portal costs factored in
- [ ] User settings saved
- [ ] Stale data detection working
- [ ] Real-time UI updates
- [ ] Clean, usable web interface

---

## Database Schema Reference

### market_prices
```sql
CREATE TABLE market_prices (
    id SERIAL PRIMARY KEY,
    item_id VARCHAR(128) NOT NULL,
    city VARCHAR(64) NOT NULL,
    quality INTEGER DEFAULT 1,
    buy_price_max BIGINT,          -- Highest buy order
    buy_price_min BIGINT,
    sell_price_min BIGINT,         -- Lowest sell order
    sell_price_max BIGINT,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(item_id, city, quality)
);

CREATE INDEX idx_market_prices_item ON market_prices(item_id);
CREATE INDEX idx_market_prices_city ON market_prices(city);
CREATE INDEX idx_market_prices_updated ON market_prices(updated_at);
```

### market_history (daily volume)
```sql
CREATE TABLE market_history (
    id SERIAL PRIMARY KEY,
    item_id VARCHAR(128) NOT NULL,
    city VARCHAR(64) NOT NULL,
    quality INTEGER DEFAULT 1,
    date DATE NOT NULL,
    avg_price BIGINT,
    item_count INTEGER,            -- Daily volume
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(item_id, city, quality, date)
);

CREATE INDEX idx_market_history_item ON market_history(item_id);
CREATE INDEX idx_market_history_date ON market_history(date);
```

---

## Quick Commands Reference

### Docker
```powershell
# Start PostgreSQL
docker-compose -f docker-compose.local.yml up -d

# Stop PostgreSQL
docker-compose -f docker-compose.local.yml down

# View logs
docker-compose -f docker-compose.local.yml logs -f

# Connect to PostgreSQL CLI
docker exec -it albion-postgres psql -U albion -d albion
```

### Go
```powershell
# Download dependencies
go mod download

# Build
go build -o albiondata-client.exe

# Run
.\albiondata-client.exe
```

### Git
```powershell
# Create feature branch
git checkout -b feature/local-database

# Commit changes
git add .
git commit -m "message"

# Switch back to main
git checkout main
```

---

## Next Steps

When ready to start, begin with **Milestone 1.1: Environment Setup - Docker & PostgreSQL**

---

*Plan created: April 3, 2026*
