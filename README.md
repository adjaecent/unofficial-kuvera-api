# Unofficial Kuvera API Go Library (Read-Only)

[![Go Reference](https://pkg.go.dev/badge/github.com/adjaecent/unofficial-kuvera-api.svg)](https://pkg.go.dev/github.com/adjaecent/unofficial-kuvera-api)

An unofficial **read-only** Go client library for the [Kuvera](https://kuvera.in) investment platform API.

## 📚 Documentation

- **API Reference**: [pkg.go.dev](https://pkg.go.dev/github.com/adjaecent/unofficial-kuvera-api)
- **Interactive API Documentation**: [OpenAPI Specification](https://adjaecent.github.io/unofficial-kuvera-api/api)
- **OpenAPI YAML**: [Raw Specification](https://adjaecent.github.io/unofficial-kuvera-api/openapi.yaml)

> **🔍 Read-Only Library**: For data retrieval and analysis only. Cannot place trades or modify accounts.

> **⚠️ Disclaimer**: Unofficial library, not affiliated with Kuvera. Use at your own risk.

## 🚀 Features (Read-Only Data Access)

- ✅ **User Authentication** - Login with username/password to get access tokens
- ✅ **Portfolio Data** - Retrieve complete portfolio including mutual funds, gold, equities, and FDs
- ✅ **Holdings Details** - Get detailed fund holdings with transaction history and SIP information
- ✅ **Gold Prices** - Get current gold buy/sell prices and tax information (auth required)

## 📦 Installation

```bash
go get github.com/adjaecent/unofficial-kuvera-api
```

## 🏃‍♂️ Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/adjaecent/unofficial-kuvera-api"
)

func main() {
    // Create a new client
    client := kuvera.NewClient()
    ctx := context.Background()

    // Login to get access token
    resp, err := client.Login(ctx, "your_username", "your_password")
    if err != nil {
        log.Fatal("Login failed:", err)
    }
    fmt.Printf("✓ Login successful! Welcome %s\n", resp.Name)

    // Get gold prices (requires authentication)
    goldPrice, err := client.GetGoldPrice(ctx)
    if err != nil {
        log.Fatal("Gold price failed:", err)
    }
    fmt.Printf("🥇 Gold buy: ₹%.2f, sell: ₹%.2f per gram\n",
        goldPrice.CurrentGoldPrice.Buy, goldPrice.CurrentGoldPrice.Sell)

    // Retrieve complete portfolio data
    portfolio, err := client.GetPortfolio(ctx)
    if err != nil {
        log.Fatal("Failed to get portfolio:", err)
    }

    fmt.Printf("📈 Total portfolio value: ₹%.2f\n", portfolio.Data.CurrentValue)
    fmt.Printf("📊 Overall gain: %.2f%%\n", portfolio.Data.CurrentGainPercent)

    // Display asset breakdown
    fmt.Printf("🏛️ Mutual funds: ₹%.2f\n", portfolio.Data.MutualFunds.CurrentValue)
    fmt.Printf("🥇 Gold: ₹%.2f\n", portfolio.Data.Gold.CurrentValue)
    fmt.Printf("📈 Indian equities: ₹%.2f\n", portfolio.Data.IndianEquities.CurrentValue)

    // Get detailed holdings
    holdings, err := client.GetHoldings(ctx)
    if err != nil {
        log.Fatal("Failed to get holdings:", err)
    }

    fmt.Printf("📊 Total fund codes: %d\n", len(*holdings))
    for fundCode, fundHoldings := range *holdings {
        for _, holding := range fundHoldings {
            fmt.Printf("💼 %s: ₹%.2f (%.3f units)\n",
                fundCode, holding.AllottedAmount, holding.Units)
            break // Show only first holding per fund
        }
        break // Show only first fund
    }
}
```

## 🔧 Examples

### Running the Examples

#### 1. **Working Example** (`example/main.go`)
- **Purpose**: Demonstrates real API usage with actual credentials
- **Usage**: Update credentials in the file or set environment variables, then run:
  ```bash
  export KUVERA_USERNAME="your_username"
  export KUVERA_PASSWORD="your_password"
  go run example/main.go
  ```

#### 2. **Documentation Examples** (`example_test.go`)
- **Purpose**: Provides godoc examples and API documentation
- **Usage**: View in documentation or run tests:
  ```bash
  go test -run Example
  godoc -http=:6060  # View at http://localhost:6060
  ```

## 📖 Local Development

Run `godoc -http=:6060` and visit http://localhost:6060 for local documentation.

## 🔒 Security Considerations

- **Never commit credentials** to version control
- **Use environment variables** for sensitive information:
  ```go
  username := os.Getenv("KUVERA_USERNAME")
  password := os.Getenv("KUVERA_PASSWORD")
  ```
- **Validate all responses** before using data for investment decisions
- **Test thoroughly** with small amounts before scaling

## 🛠️ Development

### Prerequisites

- Go 1.25.0+ (see `go.mod`)
- Valid Kuvera account credentials

### Building

```bash
# Build the library
go build

# Run tests
go test

# Generate documentation
godoc -http=:6060

# Run example
go run example/main.go
```
