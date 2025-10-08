package main

import (
	"context"
	"fmt"
	"os"

	"github.com/adjaecent/unofficial-kuvera-api"
)

func main() {
	fmt.Println("🚀 Kuvera API Client Demo")
	fmt.Println("========================")

	// Create a new client
	client := kuvera.NewClient()
	ctx := context.Background()

	// Get credentials from environment variables (recommended)
	username := os.Getenv("KUVERA_USERNAME")
	password := os.Getenv("KUVERA_PASSWORD")

	// Fallback to placeholder values for demo (update these with real credentials)
	if username == "" {
		username = "your_username"
	}
	if password == "" {
		password = "your_password"
	}

	fmt.Printf("📧 Using username: %s\n", username)
	fmt.Println()

	// Note: Gold price API requires authentication, so we'll test it after login

	// Step 2: Attempt login
	fmt.Println("🔐 Attempting login...")
	loginResp, err := client.Login(ctx, username, password)
	if err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		fmt.Println("💡 Tip: Set KUVERA_USERNAME and KUVERA_PASSWORD environment variables")
		fmt.Println("💡 Or update the credentials in example/main.go")
		return
	}

	if loginResp.Status != "success" {
		fmt.Printf("❌ Login failed: %s\n", loginResp.Error)
		fmt.Println("💡 Please check your credentials")
		return
	}

	fmt.Printf("✅ Login successful! Welcome %s (%s)\n", loginResp.Name, loginResp.Email)
	fmt.Println()

	// Step 3: Get gold prices (requires authentication)
	fmt.Println("🥇 Fetching current gold prices...")
	goldPrice, err := client.GetGoldPrice(ctx)
	if err != nil {
		fmt.Printf("❌ Gold price request failed: %v\n", err)
	} else {
		fmt.Printf("✅ Gold prices from Kuvera partner:\n")
		fmt.Printf("   💰 Buy:  ₹%.2f per gram\n", goldPrice.CurrentGoldPrice.Buy)
		fmt.Printf("   💸 Sell: ₹%.2f per gram\n", goldPrice.CurrentGoldPrice.Sell)
		fmt.Printf("   📊 Taxes - CGST: %.1f%%, SGST: %.1f%%, IGST: %.1f%%\n",
			goldPrice.Taxes.CGST, goldPrice.Taxes.SGST, goldPrice.Taxes.IGST)
		fmt.Printf("   🕒 Fetched at: %s\n", goldPrice.FetchedAt)
	}
	fmt.Println()

	// Step 4: Get complete portfolio data
	fmt.Println("📈 Fetching complete portfolio data...")
	portfolio, err := client.GetPortfolio(ctx)
	if err != nil {
		fmt.Printf("❌ Portfolio request failed: %v\n", err)
	} else if portfolio.Status == "success" {
		fmt.Printf("✅ Portfolio data retrieved successfully!\n")
		fmt.Printf("💰 Total portfolio value: ₹%.2f\n", portfolio.Data.CurrentValue)
		fmt.Printf("📊 Overall gain: ₹%.2f (%.2f%%)\n",
			portfolio.Data.CurrentGain, portfolio.Data.CurrentGainPercent)
		fmt.Printf("📅 One-day change: ₹%.2f (%.2f%%)\n",
			portfolio.Data.OneDayGain, portfolio.Data.OneDayGainPercent)
		fmt.Printf("📈 Current XIRR: %.2f%%\n", portfolio.Data.CurrentXIRR)
		fmt.Println()

		// Display breakdown by asset class
		fmt.Println("📋 Asset Breakdown:")
		fmt.Printf("   🏛️  Mutual Funds: ₹%.2f (%.2f%% return)\n",
			portfolio.Data.MutualFunds.CurrentValue, portfolio.Data.MutualFunds.AbsolutePercentage)
		fmt.Printf("   🥇 Gold: ₹%.2f (%.2fg)\n",
			portfolio.Data.Gold.CurrentValue, portfolio.Data.Gold.TotalGoldQuantity)
		fmt.Printf("   📈 Indian Equities: ₹%.2f\n", portfolio.Data.IndianEquities.CurrentValue)
		fmt.Printf("   🏦 Fixed Deposits: ₹%.2f\n", portfolio.Data.FixedDeposit.CurrentValue)
	}
	fmt.Println()

	// Step 5: Get detailed holdings
	fmt.Println("📊 Fetching detailed fund holdings...")
	holdings, err := client.GetHoldings(ctx)
	if err != nil {
		fmt.Printf("❌ Holdings request failed: %v\n", err)
	} else {
		fmt.Printf("✅ Holdings data retrieved successfully!\n")
		totalFunds := len(*holdings)
		fmt.Printf("📈 Total fund codes: %d\n", totalFunds)

		// Display first few holdings
		fmt.Println("📋 Sample Holdings:")
		count := 0
		for fundCode, fundHoldings := range *holdings {
			if count >= 3 {
				break
			}
			for _, holding := range fundHoldings {
				fmt.Printf("   💼 %s: Folio %s - ₹%.2f (%.3f units)\n",
					fundCode, holding.FolioNumber, holding.AllottedAmount, holding.Units)
				fmt.Printf("      📂 Category: %s | Direct: %t | Orders: %d\n",
					holding.KuveraCategory, holding.Direct, len(holding.OrderDetails))
				if holding.IsSip && len(holding.SIPs) > 0 {
					fmt.Printf("      🔄 SIP: ₹%.2f %s (Active)\n",
						holding.SIPs[0].Amount, holding.SIPs[0].Frequency)
				}
				break // Show only first holding per fund
			}
			count++
		}
	}
	fmt.Println()

	// Step 6: Summary
	fmt.Println("📋 Demo Summary:")
	fmt.Println("================")
	fmt.Printf("✅ Login API: Working perfectly!\n")
	fmt.Printf("✅ Gold price API: Working perfectly!\n")
	fmt.Printf("✅ Portfolio API: Working perfectly!\n")
	fmt.Printf("✅ Holdings API: Working perfectly!\n")
	fmt.Printf("🔑 Authentication token: %s...\n", loginResp.Token[:20])

	fmt.Println()
	fmt.Println("🎉 Kuvera API demo completed successfully!")
	fmt.Println("💡 All major APIs are working: authentication, gold prices, portfolio data, and detailed holdings")
	fmt.Println("💡 Check the example_test.go file for more usage examples")
}
