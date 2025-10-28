# DeFi Rates

A DeFi yield rate comparison site built with Go and HTMX. Compare yield rates across different DeFi protocols in real-time.

## Features

- **Real-time Yield Data**: Automatically fetches and updates yield rates from DeFi protocols
- **Multi-Protocol Support**: Currently supports Pendle with plans to expand to more protocols
- **Advanced Filtering**: Filter by asset, chain, APY range, and TVL
- **Responsive Design**: Clean, modern UI that works on desktop and mobile
- **Fast & Lightweight**: Built with Go and HTMX for optimal performance
- **No Database Setup Required**: Uses SQLite for zero-configuration data storage

## Current Protocol Support

### Pendle
- Fetches all active markets across 10 supported chains
- **Supported chains**: Ethereum, Arbitrum, Optimism, Base, BSC, Mantle, Zora, Sonic, Taiko, Berachain
- Displays implied APY, TVL, maturity dates, and pool information
- Direct links to pool pages on Pendle app
- Automatic expiry filtering (excludes expired markets)

### Coming Soon
The midterm goal is to integrate all protocols listed on [OpenYield](https://www.openyield.com).

## Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: HTMX + HTML templates
- **Database**: SQLite3
- **Styling**: Custom CSS with responsive design
- **Data Source**: Pendle API v2

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

The server will start on `http://localhost:8080` with pre-loaded sample data showing various Pendle yield opportunities.

**Note**: The Pendle API may be rate-limited or require authentication for direct access. Use the `-load-sample` flag to see the application functionality with realistic sample data.

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

### Pendle API Access Issues

If you see warnings like "failed to fetch markets" or "Access denied", this is normal in certain network environments:

**Common causes:**
- Network firewalls blocking cryptocurrency/DeFi APIs
- Corporate proxy restrictions
- API rate limiting
- WAF (Web Application Firewall) rules

**Solution:**
Use the sample data mode which provides realistic Pendle yield opportunities across multiple chains without requiring external API access:

```bash
./defirates -load-sample
```

This demonstrates all features including:
- Multi-chain filtering
- APY and TVL range filtering
- Asset-specific searches
- Real-time HTMX updates
- Responsive UI

### Testing API Access

To test if the Pendle API is accessible from your environment:

```bash
curl -s "https://api-v2.pendle.finance/api/core/v1/1/markets/active"
```

- **200 OK + JSON data**: API is accessible, real data fetching will work
- **403 Forbidden**: API is blocked, use `-load-sample` mode
- **400 Bad Request**: API is accessible but check error message for details

## Project Structure

```
defirates/
├── cmd/
│   └── server/                  # Application entry point
│       └── main.go
├── internal/
│   ├── api/                     # External API clients
│   │   ├── pendle.go           # Pendle API client
│   │   ├── fetcher.go          # Data fetching service
│   │   ├── pendle_test.go      # API client unit tests
│   │   └── integration_test.go # End-to-end integration tests
│   ├── database/                # Database layer
│   │   ├── database.go         # SQLite operations
│   │   └── database_test.go    # Database unit tests
│   ├── handlers/                # HTTP handlers
│   │   ├── handlers.go
│   │   ├── handlers_test.go    # Handler/template tests
│   │   └── templates/          # HTML templates
│   │       ├── index.html
│   │       └── table.html
│   └── models/                  # Data models
│       ├── yield.go
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
- `min_apy`: Minimum APY percentage
- `max_apy`: Maximum APY percentage
- `min_tvl`: Minimum Total Value Locked in USD
- `sort_by`: Sort field ("apy", "tvl", "updated_at")
- `sort_order`: Sort order ("asc", "desc")

**Response:**
- Full HTML page on initial load
- Table fragment on HTMX requests (for dynamic updates)

## How It Works

1. **Data Fetching**: On startup, the application fetches yield data from Pendle's API
2. **Database Storage**: Data is stored in SQLite with automatic upserts to prevent duplicates
3. **Periodic Updates**: A background goroutine refreshes data at the configured interval
4. **Real-time Filtering**: HTMX enables instant filtering without page reloads
5. **Responsive UI**: Clean, modern interface adapts to all screen sizes

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
- `maturity_date`: Expiry date for fixed-term yields
- `pool_name`: Pool identifier
- `external_url`: Link to protocol's pool page
- `updated_at`: Last update timestamp
- `created_at`: Creation timestamp

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Roadmap

- [x] Pendle integration
- [x] Comprehensive test suite (59.2% coverage)
- [x] Multi-chain support (10 chains)
- [x] Real-time filtering with HTMX
- [ ] Add more protocols (Aave, Compound, etc.)
- [ ] Historical data tracking and charts
- [ ] Email/webhook notifications for high yields
- [ ] API endpoint for programmatic access
- [ ] User accounts and watchlists
- [ ] Mobile app

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Pendle Finance](https://www.pendle.finance) for their excellent API
- [HTMX](https://htmx.org) for making dynamic UIs simple
- [OpenYield](https://www.openyield.com) for DeFi protocol inspiration

## Support

For issues, questions, or suggestions, please open an issue on GitHub.
