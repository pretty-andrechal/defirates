# DeFi Rates

A DeFi yield rate comparison site built with Go and HTMX. Compare yield rates across multiple DeFi protocols in real-time with automatic updates.

## Features

- **Real-time Web Updates**: Server-Sent Events (SSE) push updates to your browser automatically
- **Multi-Protocol Support**: Supports Pendle and Beefy Finance with plans to expand
- **Advanced Filtering**: Filter by asset, chain, APY range, TVL, and categories
- **Categories**: View and filter by yield categories (PT, LRT, LST, Stablecoin, etc.)
- **Responsive Design**: Clean, modern UI that works on desktop and mobile
- **Fast & Lightweight**: Built with Go and HTMX for optimal performance
- **No Database Setup Required**: Uses SQLite for zero-configuration data storage
- **Live Update Indicators**: Visual feedback when data refreshes in real-time

## Current Protocol Support

### Pendle
- Fetches all active markets across 10 supported chains
- **Supported chains**: Ethereum, Arbitrum, Optimism, Base, BSC, Mantle, Zora, Sonic, Taiko, Berachain
- Displays implied APY, TVL, maturity dates, and pool information
- Categories from market categoryIds
- Direct links to pool pages on Pendle app
- Automatic expiry filtering (excludes expired markets)

### Beefy Finance
- Comprehensive vault coverage across 23 chains
- **Supported chains**: Ethereum, Arbitrum, Optimism, Base, BSC, Polygon, Avalanche, Fantom, Cronos, Aurora, Celo, Moonbeam, Moonriver, Metis, Emerald, Kava, Fuse, Harmony, Canto, zkSync, Polygon zkEVM, and more
- Displays total APY with detailed breakdown, TVL, and platform information
- Multiple asset categories per vault (shows underlying assets)
- Direct links to vaults on Beefy app
- Only shows active vaults (excludes retired/EOL vaults)

### Coming Soon
The midterm goal is to integrate all protocols listed on [OpenYield](https://www.openyield.com).

## Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: HTMX + HTML templates + Server-Sent Events (SSE)
- **Database**: SQLite3
- **Styling**: Custom CSS with responsive design
- **Data Sources**: Pendle API v2, Beefy Finance API

## Getting Started

### Prerequisites

- Go 1.21 or higher
- GCC compiler (required for SQLite3)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/pretty-andrechal/defirates.git
cd defirates
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o defirates ./cmd/server
```

### Running the Application

Start the server with sample data for demonstration:
```bash
./defirates -load-sample
```

The server will start on `http://localhost:8080` with pre-loaded sample data:
- **14 Pendle yield rates** across Ethereum, Arbitrum, Optimism, Base, and Mantle
- **10 Beefy vault yields** across Ethereum, Arbitrum, BSC, Polygon, Avalanche, Optimism, and Base
- Total of **24 yield rates** demonstrating all features including categories, filtering, and real-time updates

**Important**: Both Pendle and Beefy APIs may return 403 errors from certain environments due to Cloudflare/WAF blocking. This is normal and doesn't affect the application functionality. Use `-load-sample` for testing, or see [API_ACCESS_TROUBLESHOOTING.md](./API_ACCESS_TROUBLESHOOTING.md) for details and solutions.

### Command-Line Options

Customize the server behavior with these flags:

```bash
./defirates -port 3000 -db mydata.db -fetch-interval 10m -load-sample
```

Available options:
- `-port`: HTTP port (default: 8080)
- `-db`: SQLite database path (default: defirates.db)
- `-fetch-interval`: Data refresh interval (default: 5m)
- `-load-sample`: Load sample data for demonstration (recommended for first run)

### Development Mode

For development with auto-reload, you can use `go run`:

```bash
go run ./cmd/server -load-sample -fetch-interval 1m
```

## Troubleshooting

### API Access Issues

If you see warnings like "failed to fetch markets/vaults" or "Access denied", this is normal in certain network environments:

**Common causes:**
- Network firewalls blocking cryptocurrency/DeFi APIs
- Corporate proxy restrictions
- API rate limiting
- WAF (Web Application Firewall) rules

**Solution:**
Use the sample data mode which provides realistic yield opportunities across multiple chains without requiring external API access:

```bash
./defirates -load-sample
```

This demonstrates all features including:
- Multi-chain filtering
- APY and TVL range filtering
- Asset and category filtering
- Real-time SSE updates
- Protocol filtering (Pendle/Beefy)
- Responsive UI

### Testing API Access

To test if the APIs are accessible from your environment:

**Pendle API:**
```bash
curl -s "https://api-v2.pendle.finance/api/core/v1/1/markets/active"
```

**Beefy Finance API:**
```bash
curl -s "https://api.beefy.finance/vaults/ethereum"
curl -s "https://api.beefy.finance/apy/breakdown"
```

- **200 OK + JSON data**: API is accessible, real data fetching will work
- **403 Forbidden**: API is blocked, use `-load-sample` mode
- **400 Bad Request**: API is accessible but check error message for details

**For detailed troubleshooting information**, including root cause analysis, deployment recommendations, and production considerations, see [API_ACCESS_TROUBLESHOOTING.md](./API_ACCESS_TROUBLESHOOTING.md).

## Project Structure

```
defirates/
├── cmd/
│   └── server/                  # Application entry point
│       └── main.go
├── internal/
│   ├── api/                     # External API clients
│   │   ├── pendle.go           # Pendle API client
│   │   ├── beefy.go            # Beefy Finance API client
│   │   ├── fetcher.go          # Multi-protocol data fetching service
│   │   ├── sample_data.go      # Sample data for demonstration
│   │   ├── pendle_test.go      # Pendle API unit tests
│   │   ├── beefy_test.go       # Beefy API unit tests
│   │   └── integration_test.go # End-to-end integration tests
│   ├── database/                # Database layer
│   │   ├── database.go         # SQLite operations
│   │   └── database_test.go    # Database unit tests
│   ├── handlers/                # HTTP handlers & SSE
│   │   ├── handlers.go         # HTTP request handlers
│   │   ├── events.go           # Server-Sent Events manager
│   │   ├── handlers_test.go    # Handler/template tests
│   │   ├── events_test.go      # SSE manager tests
│   │   └── templates/          # HTML templates
│   │       ├── index.html      # Main page with SSE client
│   │       ├── table.html      # Yield rates table
│   │       └── rows.html       # Table rows template (for updates)
│   └── models/                  # Data models
│       ├── yield.go            # YieldRate and FilterParams models
│       └── models_test.go      # Model tests
├── static/
│   └── css/                    # Stylesheets
│       └── style.css
├── go.mod
├── go.sum
├── run_tests.sh                # Test suite runner
├── TESTING_GUIDE.md            # Comprehensive testing documentation
└── README.md
```

## API Endpoints

### `GET /`
Main page with yield rates table and filters.

**Query Parameters:**
- `asset`: Filter by asset (e.g., "ETH", "USDC")
- `chain`: Filter by blockchain (e.g., "Ethereum", "Arbitrum")
- `categories`: Filter by category (e.g., "PT", "LRT", "Stablecoin")
- `min_apy`: Minimum APY percentage
- `max_apy`: Maximum APY percentage
- `min_tvl`: Minimum Total Value Locked in USD
- `sort_by`: Sort field ("apy", "tvl", "updated_at")
- `sort_order`: Sort order ("asc", "desc")

**Response:**
- Full HTML page on initial load
- Table fragment on HTMX requests (for dynamic updates)

### `GET /events`
Server-Sent Events (SSE) endpoint for real-time updates.

**Response:**
- Event stream with `text/event-stream` content type
- Sends `connected` event on initial connection
- Sends `update` events when data is refreshed
- Clients automatically reconnect on disconnect

### `GET /api/rates`
Fetch specific yield rates by IDs (used for live updates).

**Query Parameters:**
- `ids`: Comma-separated list of rate IDs (e.g., "1,5,23,45")

**Response:**
- HTML table rows for the requested rates
- Used by SSE client to update visible rows in-place

## How It Works

1. **Multi-Protocol Data Fetching**: On startup, fetches yield data from Pendle and Beefy Finance APIs
2. **Database Storage**: Data is stored in SQLite with automatic upserts to prevent duplicates
3. **Periodic Updates**: A background goroutine refreshes data at the configured interval (default 5 minutes)
4. **Real-time Push Updates**: Server-Sent Events broadcast data updates to all connected browsers
5. **In-Place Updates**: JavaScript fetches fresh data for visible rows and updates them without re-filtering
6. **Smart Filtering**: HTMX enables instant filtering without page reloads; updates respect current filters
7. **Responsive UI**: Clean, modern interface adapts to all screen sizes

## Testing

The project includes a comprehensive test suite covering all layers of the application.

### Running Tests

**Quick test run:**
```bash
go test ./...
```

**Comprehensive test suite with coverage and race detection:**
```bash
./run_tests.sh
```

This script runs:
- Unit tests for all packages (API, Database, Handlers, Models)
- Integration tests (end-to-end API flows)
- Race condition detection
- Test coverage reporting (currently **59.2%** overall)

**Run specific test suites:**
```bash
# API unit tests only (skips integration tests)
go test -v ./internal/api -short

# Integration tests only
go test -v ./internal/api -run Integration

# Database tests
go test -v ./internal/database

# Handler/template tests
go test -v ./internal/handlers

# Model tests
go test -v ./internal/models
```

**Note for macOS users:** The test script automatically suppresses harmless linker warnings about `LC_DYSYMTAB` that appear when running tests with the race detector. These warnings don't affect test functionality.

### Test Coverage

- **API Package**: 49.1% coverage (unit tests + integration tests)
- **Database Package**: 86.4% coverage (CRUD operations, filtering, sorting)
- **Handlers Package**: 60.4% coverage (HTTP handlers, template rendering)
- **Models Package**: 100% coverage (struct marshaling, JSON tags)
- **Overall**: 59.2% coverage

### Testing Documentation

For detailed information about the test suite, test scenarios, debugging, and CI/CD integration, see [TESTING_GUIDE.md](TESTING_GUIDE.md).

## Adding More Protocols

To add a new protocol:

1. Create a new API client in `internal/api/`
2. Implement the fetching logic in `fetcher.go`
3. Update the database models if needed
4. Add the protocol to the periodic fetch routine

Example structure:
```go
type AaveClient struct {
    // Client implementation
}

func (f *Fetcher) FetchAndStoreAaveData() error {
    // Fetching logic
}
```

## Database Schema

### `protocols` table
- `id`: Primary key
- `name`: Protocol name (unique)
- `url`: Protocol website
- `description`: Protocol description
- `created_at`: Timestamp

### `yield_rates` table
- `id`: Primary key
- `protocol_id`: Foreign key to protocols
- `asset`: Asset symbol (e.g., "ETH")
- `chain`: Blockchain name
- `apy`: Annual Percentage Yield
- `tvl`: Total Value Locked in USD
- `maturity_date`: Expiry date for fixed-term yields (nullable)
- `pool_name`: Pool identifier
- `categories`: Comma-separated categories (e.g., "PT, LRT, Liquidity")
- `external_url`: Link to protocol's pool page
- `updated_at`: Last update timestamp
- `created_at`: Creation timestamp

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Roadmap

- [x] Pendle integration (10 chains)
- [x] Beefy Finance integration (23 chains)
- [x] Comprehensive test suite with coverage reporting
- [x] Multi-chain support (30+ chains total)
- [x] Real-time filtering with HTMX
- [x] Server-Sent Events (SSE) for live updates
- [x] Categories support with filtering
- [ ] Add more protocols (Aave, Compound, Yearn, etc.)
- [ ] Historical data tracking and charts
- [ ] Email/webhook notifications for high yields
- [ ] REST API endpoint for programmatic access
- [ ] User accounts and watchlists
- [ ] Mobile app

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Pendle Finance](https://www.pendle.finance) for their excellent API
- [Beefy Finance](https://beefy.finance) for their comprehensive vault API
- [HTMX](https://htmx.org) for making dynamic UIs simple
- [OpenYield](https://www.openyield.com) for DeFi protocol inspiration

## Support

For issues, questions, or suggestions, please open an issue on GitHub.
