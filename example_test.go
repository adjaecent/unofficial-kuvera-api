package kuvera_test

import (
	"context"
	"fmt"
	"log"

	"github.com/adjaecent/unofficial-kuvera-api"
)

// ExampleNewClient demonstrates how to create a new Kuvera client.
func ExampleNewClient() {
	client := kuvera.NewClient()
	fmt.Printf("Client created successfully: %T", client)
	// Output: Client created successfully: *kuvera.Client
}

// ExampleNewClient_withOptions demonstrates how to create a new Kuvera client with custom options.
func ExampleNewClient_withOptions() {
	client := kuvera.NewClient(
		kuvera.WithUserAgent("my-app/1.0"),
		kuvera.WithTimeout(60000000000), // 60 seconds in nanoseconds
	)
	fmt.Printf("Client created with options: %T", client)
	// Output: Client created with options: *kuvera.Client
}

// ExampleClient_Login demonstrates how to authenticate with Kuvera.
func ExampleClient_Login() {
	client := kuvera.NewClient()
	ctx := context.Background()

	// Note: Use real credentials in actual implementation
	resp, err := client.Login(ctx, "demo@example.com", "demopassword")
	if err != nil {
		log.Fatal(err)
	}

	if resp.Status == "success" {
		fmt.Printf("Login successful! Welcome %s", resp.Name)
	} else {
		fmt.Printf("Login failed: %s", resp.Error)
	}
}

// ExampleClient_GetPortfolio demonstrates how to retrieve complete portfolio data.
func ExampleClient_GetPortfolio() {
	client := kuvera.NewClient()
	ctx := context.Background()

	// First, authenticate
	_, err := client.Login(ctx, "demo@example.com", "demopassword")
	if err != nil {
		log.Fatal("Login failed:", err)
	}

	// Get portfolio data
	portfolio, err := client.GetPortfolio(ctx)
	if err != nil {
		log.Fatal("Failed to get portfolio:", err)
	}

	fmt.Printf("✓ Total portfolio value: ₹%.2f\n", portfolio.Data.CurrentValue)
	fmt.Printf("✓ Overall gain: %.2f%%\n", portfolio.Data.CurrentGainPercent)

	// Display asset breakdown
	fmt.Printf("📈 Mutual funds: ₹%.2f\n", portfolio.Data.MutualFunds.CurrentValue)
	fmt.Printf("🥇 Gold: ₹%.2f\n", portfolio.Data.Gold.CurrentValue)
	fmt.Printf("📊 Indian equities: ₹%.2f\n", portfolio.Data.IndianEquities.CurrentValue)
}

// ExampleClient_GetHoldings demonstrates how to retrieve detailed holdings data.
func ExampleClient_GetHoldings() {
	client := kuvera.NewClient()
	ctx := context.Background()

	// First, authenticate
	_, err := client.Login(ctx, "demo@example.com", "demopassword")
	if err != nil {
		log.Fatal("Login failed:", err)
	}

	// Get holdings data
	holdings, err := client.GetHoldings(ctx)
	if err != nil {
		log.Fatal("Failed to get holdings:", err)
	}

	fmt.Printf("✓ Total fund codes: %d\n", len(*holdings))

	// Display sample holdings
	count := 0
	for fundCode, fundHoldings := range *holdings {
		if count >= 2 {
			break
		}
		for _, holding := range fundHoldings {
			fmt.Printf("📈 %s: ₹%.2f (%.3f units)\n",
				fundCode, holding.AllottedAmount, holding.Units)
			break // Show only first holding per fund
		}
		count++
	}
}

// ExampleClient_GetGoldPrice demonstrates how to retrieve current gold prices.
func ExampleClient_GetGoldPrice() {
	client := kuvera.NewClient()
	ctx := context.Background()

	// First, authenticate (required for gold price API)
	_, err := client.Login(ctx, "demo@example.com", "demopassword")
	if err != nil {
		log.Fatal("Login failed:", err)
	}

	// Get gold price (requires authentication)
	goldPrice, err := client.GetGoldPrice(ctx)
	if err != nil {
		log.Fatal("Failed to get gold price:", err)
	}

	fmt.Printf("🥇 Gold buy: ₹%.2f per gram\n", goldPrice.CurrentGoldPrice.Buy)
	fmt.Printf("💸 Gold sell: ₹%.2f per gram\n", goldPrice.CurrentGoldPrice.Sell)
	fmt.Printf("📊 Taxes - CGST: %.1f%%, SGST: %.1f%%\n",
		goldPrice.Taxes.CGST, goldPrice.Taxes.SGST)
}

// ExampleClient_workflowExample demonstrates a complete workflow.
func ExampleClient_workflowExample() {
	client := kuvera.NewClient()
	ctx := context.Background()

	// Step 1: Login
	fmt.Println("🔐 Logging in...")
	loginResp, err := client.Login(ctx, "demo@example.com", "demopassword")
	if err != nil {
		log.Fatal("Login failed:", err)
	}
	fmt.Printf("✓ Login successful! Welcome %s\n", loginResp.Name)

	// Step 2: Get gold price (public data)
	fmt.Println("\n🥇 Fetching gold price...")
	goldPrice, err := client.GetGoldPrice(ctx)
	if err != nil {
		log.Fatal("Gold price failed:", err)
	}
	fmt.Printf("✓ Gold buy: ₹%.2f, sell: ₹%.2f per gram\n",
		goldPrice.CurrentGoldPrice.Buy, goldPrice.CurrentGoldPrice.Sell)

	// Step 3: Get portfolio data
	fmt.Println("\n📈 Fetching portfolio...")
	portfolio, err := client.GetPortfolio(ctx)
	if err != nil {
		log.Fatal("Portfolio failed:", err)
	}
	fmt.Printf("✓ Portfolio value: ₹%.2f\n", portfolio.Data.CurrentValue)

	// Step 4: Analyze portfolio performance
	fmt.Println("\n🏆 Portfolio performance:")
	fmt.Printf("  • Overall gain: %.2f%%\n", portfolio.Data.CurrentGainPercent)
	fmt.Printf("  • Mutual funds return: %.2f%%\n", portfolio.Data.MutualFunds.AbsolutePercentage)
	fmt.Printf("  • Current XIRR: %.2f%%\n", portfolio.Data.CurrentXIRR)
}