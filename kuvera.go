// Package kuvera provides an unofficial Go client library for the Kuvera API.
//
// Kuvera is a platform that allows investing in mutual funds and ETFs in India. This library
// provides a simple interface to interact with Kuvera's REST API for authentication,
// mutual fund data, and gold prices.
//
// # Basic Usage
//
//	client := kuvera.NewClient()
//
//	// Login to get access token
//	resp, err := client.Login("username", "password")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get mutual fund details
//	funds, err := client.GetMutualFunds()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get gold prices
//	goldPrice, err := client.GetGoldPrice()
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Authentication
//
// All API calls except Login and GetGoldPrice require authentication. The Client automatically
// stores and includes the access token from a successful login in subsequent requests.
//
// # Error Handling
//
// All methods return detailed error information. Network errors, JSON parsing
// errors, and API errors are wrapped with descriptive messages.
package kuvera

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BaseURL is the base URL for the Kuvera API.
const (
	BaseURL = "https://api.kuvera.in"
	DefaultTimeout = 30 * time.Second
	DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:143.0) Gecko/20100101 Firefox/143.0"
)

// Common errors
var (
	ErrNotAuthenticated = errors.New("not authenticated: please login first")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmptyUsername = errors.New("username cannot be empty")
	ErrEmptyPassword = errors.New("password cannot be empty")
)

// APIError represents an error response from the Kuvera API.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     string `json:"error,omitempty"`
}

func (e *APIError) Error() string {
	if e.Err != "" {
		return fmt.Sprintf("API error %d: %s - %s", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
}

// KuveraClient defines the interface for Kuvera API operations.
type KuveraClient interface {
	// Login authenticates with username/password and returns user info and JWT token
	Login(ctx context.Context, username, password string) (*LoginResponse, error)
	// GetPortfolio retrieves complete portfolio data including all investments (requires authentication)
	GetPortfolio(ctx context.Context) (*PortfolioResponse, error)
	// GetHoldings retrieves detailed holdings information for all funds (requires authentication)
	GetHoldings(ctx context.Context) (*HoldingsResponse, error)
	// GetGoldPrice retrieves current gold buy/sell prices (requires authentication)
	GetGoldPrice(ctx context.Context) (*GoldPriceResponse, error)
}

// ClientOption is a function that configures a Client.
type ClientOption func(*clientConfig)

// clientConfig holds configuration for the client.
type clientConfig struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// WithBaseURL sets a custom base URL for the API.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *clientConfig) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *clientConfig) {
		c.httpClient = client
	}
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *clientConfig) {
		c.userAgent = userAgent
	}
}

// WithTimeout sets a custom timeout for requests.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
	}
}

// Client represents a Kuvera API client with authentication and HTTP configuration.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	userAgent   string
	accessToken string
	sessionID   string
}

// LoginRequest represents the request payload for user authentication.
type LoginRequest struct {
	// Email is the user's login email
	Email string `json:"email"`
	// Password is the user's login password
	Password string `json:"password"`
	// V is the version parameter
	V string `json:"v"`
}

// LoginResponse represents the response from the login API endpoint.
type LoginResponse struct {
	// Status indicates if the login was successful ("success" or "error")
	Status string `json:"status"`
	// Name is the user's full name
	Name string `json:"name"`
	// Email is the user's email address
	Email string `json:"email"`
	// Profile contains additional profile information
	Profile interface{} `json:"profile"`
	// NewUser indicates if this is a new user
	NewUser bool `json:"new_user"`
	// Token is the JWT token used for authenticated API calls
	Token string `json:"token"`
	// Error contains error message if login failed
	Error string `json:"error,omitempty"`
}

// GoldData represents gold investment details.
type GoldData struct {
	// OneDayChange is the one-day change in value
	OneDayChange float64 `json:"one_day_change"`
	// CurrentValue is the current value of gold holdings
	CurrentValue float64 `json:"current_value"`
	// TotalInvested is the total amount invested in gold
	TotalInvested float64 `json:"total_invested"`
	// XIRR is the extended internal rate of return
	XIRR string `json:"xirr"`
	// TotalGoldQuantity is the total quantity of gold in grams
	TotalGoldQuantity float64 `json:"total_gold_quantity"`
	// Kuvera contains Kuvera-specific gold data
	Kuvera GoldKuveraData `json:"kuvera"`
	// Imported contains imported gold data
	Imported GoldImportedData `json:"imported"`
}

// GoldKuveraData represents Kuvera-specific gold investment data.
type GoldKuveraData struct {
	// Quantity is the quantity held through Kuvera
	Quantity float64 `json:"quantity"`
	// OneDayChange is the one-day change in Kuvera gold
	OneDayChange float64 `json:"one_day_change"`
	// InvestedValue is the amount invested through Kuvera
	InvestedValue float64 `json:"invested_value"`
	// CurrentValue is the current value of Kuvera gold
	CurrentValue float64 `json:"current_value"`
	// ProfitAmount is the profit/loss amount
	ProfitAmount float64 `json:"profit_amount"`
	// XIRR is the extended internal rate of return
	XIRR string `json:"xirr"`
}

// GoldImportedData represents imported gold investment data.
type GoldImportedData struct {
	// Quantity is the imported gold quantity
	Quantity float64 `json:"quantity"`
	// OneDayChange is the one-day change in imported gold value
	OneDayChange float64 `json:"one_day_change"`
	// InvestedValue is the invested value of imported gold
	InvestedValue float64 `json:"invested_value"`
	// CurrentValue is the current value of imported gold
	CurrentValue float64 `json:"current_value"`
	// ProfitAmount is the profit/loss amount
	ProfitAmount float64 `json:"profit_amount"`
	// XIRR is the extended internal rate of return
	XIRR float64 `json:"xirr"`
}

// IndianEquitiesData represents Indian equities investment data.
type IndianEquitiesData struct {
	// OneDayChange is the one-day change in value
	OneDayChange float64 `json:"one_day_change"`
	// CurrentValue is the current value of Indian equities
	CurrentValue float64 `json:"current_value"`
	// TotalInvested is the total amount invested
	TotalInvested float64 `json:"total_invested"`
	// OneDayChangePercentage is the one-day change percentage
	OneDayChangePercentage float64 `json:"one_day_change_percentage"`
}

// MutualFundsData represents mutual funds investment data.
type MutualFundsData struct {
	// OneDayChange is the one-day change in value
	OneDayChange float64 `json:"one_day_change"`
	// CurrentValue is the current value of mutual funds
	CurrentValue float64 `json:"current_value"`
	// TotalInvested is the total amount invested
	TotalInvested float64 `json:"total_invested"`
	// XIRRPercentage is the XIRR percentage
	XIRRPercentage float64 `json:"xirr_percentage"`
	// AbsolutePercentage is the absolute return percentage
	AbsolutePercentage float64 `json:"absolute_percentage"`
}

// FDDetails represents fixed deposit details.
type FDDetails struct {
	// AccountID is the account identifier
	AccountID int `json:"account_id"`
	// Invested is the amount invested
	Invested string `json:"invested"`
	// CurrentValue is the current value
	CurrentValue float64 `json:"current_value"`
	// OneDayChange is the one-day change
	OneDayChange float64 `json:"one_day_change"`
	// KuveraCode is the Kuvera partner code
	KuveraCode string `json:"kuvera_code"`
	// PartnerFriendlyID is the partner friendly identifier
	PartnerFriendlyID string `json:"partner_friendly_id"`
}

// FixedDepositData represents fixed deposit investment data.
type FixedDepositData struct {
	// CurrentValue is the current value of fixed deposits
	CurrentValue float64 `json:"current_value"`
	// TotalInvested is the total amount invested
	TotalInvested string `json:"total_invested"`
	// OneDayChange is the one-day change
	OneDayChange float64 `json:"one_day_change"`
	// XIRR is the extended internal rate of return
	XIRR float64 `json:"xirr"`
	// CurrentXIRR is the current XIRR
	CurrentXIRR float64 `json:"current_xirr"`
	// Interest contains interest information
	Interest interface{} `json:"interest"`
	// FDDetails contains details of individual FDs
	FDDetails []FDDetails `json:"fd_details"`
}

// PortfolioData represents the complete portfolio data.
type PortfolioData struct {
	// CurrentValue is the total current value of the portfolio
	CurrentValue float64 `json:"current_value"`
	// CurrentGain is the current gain/loss
	CurrentGain float64 `json:"current_gain"`
	// CurrentValueAssets is the current value of assets
	CurrentValueAssets float64 `json:"current_value_assets"`
	// CurrentGainPercent is the current gain percentage
	CurrentGainPercent float64 `json:"current_gain_percent"`
	// OneDayGain is the one-day gain/loss
	OneDayGain float64 `json:"one_day_gain"`
	// OneDayGainPercent is the one-day gain percentage
	OneDayGainPercent float64 `json:"one_day_gain_percent"`
	// Invested is the total amount invested
	Invested float64 `json:"invested"`
	// InvestedValueAssets is the invested value in assets
	InvestedValueAssets float64 `json:"invested_value_assets"`
	// CurrentXIRR is the current XIRR
	CurrentXIRR float64 `json:"current_xirr"`
	// AlltimeXIRR is the all-time XIRR
	AlltimeXIRR float64 `json:"alltime_xirr"`
	// AlltimeReturn is the all-time return
	AlltimeReturn float64 `json:"alltime_return"`
	// AlltimeAbsPercentage is the all-time absolute percentage
	AlltimeAbsPercentage float64 `json:"alltime_abs_percentage"`
	// AlltimeAbsReturn is the all-time absolute return
	AlltimeAbsReturn float64 `json:"alltime_abs_return"`
	// USEquities contains US equities data (empty object)
	USEquities map[string]interface{} `json:"us_equities"`
	// EPF contains EPF data (empty object)
	EPF map[string]interface{} `json:"epf"`
	// Gold contains gold investment data
	Gold GoldData `json:"gold"`
	// IndianEquities contains Indian equities data
	IndianEquities IndianEquitiesData `json:"indian_equities"`
	// MutualFunds contains mutual funds data
	MutualFunds MutualFundsData `json:"mutual_funds"`
	// SaveSmarts contains save smarts data (empty object)
	SaveSmarts map[string]interface{} `json:"save_smarts"`
	// FixedDeposit contains fixed deposit data
	FixedDeposit FixedDepositData `json:"fixed_deposit"`
}

// PortfolioResponse represents the response from the portfolio returns API endpoint.
type PortfolioResponse struct {
	// Status indicates if the request was successful
	Status string `json:"status"`
	// Data contains the portfolio data
	Data PortfolioData `json:"data"`
}

// OrderDetail represents a single order/transaction in a holding.
type OrderDetail struct {
	// Amount is the transaction amount
	Amount float64 `json:"amount"`
	// ReinvestAmount is the reinvestment amount (usually null)
	ReinvestAmount interface{} `json:"reinvest_amount"`
	// NAV is the Net Asset Value at the time of purchase
	NAV float64 `json:"nav"`
	// Units is the number of units purchased
	Units float64 `json:"units"`
	// OrderDate is the date of the order
	OrderDate string `json:"order_date"`
}

// SIPDetail represents SIP (Systematic Investment Plan) information.
type SIPDetail struct {
	// ID is the unique SIP identifier
	ID int `json:"id"`
	// PortfolioID is the portfolio identifier
	PortfolioID int `json:"portfolio_id"`
	// AMCAmfiCodeTo is the destination fund code
	AMCAmfiCodeTo string `json:"amc_amfi_code_to"`
	// AMCAmfiCodeFrom is the source fund code (usually null)
	AMCAmfiCodeFrom interface{} `json:"amc_amfi_code_from"`
	// FolioNo is the folio number
	FolioNo string `json:"folio_no"`
	// Amount is the SIP amount
	Amount float64 `json:"amount"`
	// Type is the transaction type (usually "sip")
	Type string `json:"type"`
	// Frequency is the SIP frequency (e.g., "Monthly")
	Frequency string `json:"frequency"`
	// StartDate is the SIP start date
	StartDate string `json:"start_date"`
	// EndDate is the SIP end date (usually null for ongoing)
	EndDate interface{} `json:"end_date"`
	// ISIN is the fund ISIN code
	ISIN string `json:"isin"`
	// IsUserAdded indicates if user added this SIP
	IsUserAdded interface{} `json:"isUserAdded"`
	// NoOfInstallments is the number of installments
	NoOfInstallments int `json:"no_of_installments"`
	// UpdatedAt is when the record was last updated
	UpdatedAt string `json:"updated_at"`
	// State is the current state of the SIP
	State string `json:"state"`
	// PortfolioCode is the portfolio code
	PortfolioCode string `json:"portfolio_code"`
	// BSEMessage is the message from BSE
	BSEMessage string `json:"bse_message"`
	// EmailStatus is the email status
	EmailStatus interface{} `json:"email_status"`
	// TxnRefNo is the transaction reference number
	TxnRefNo string `json:"txn_ref_no"`
	// InternalRefNo is the internal reference number
	InternalRefNo string `json:"internal_ref_no"`
	// OrderTriggerDate is when the order was triggered
	OrderTriggerDate string `json:"order_trigger_date"`
	// OrderPaymentStatus is the payment status
	OrderPaymentStatus interface{} `json:"order_payment_status"`
	// MandateID is the mandate identifier
	MandateID string `json:"mandate_id"`
	// SIPType is the type of SIP
	SIPType string `json:"sip_type"`
	// CreatedAt is when the SIP was created
	CreatedAt string `json:"created_at"`
	// BSESIPRegNo is the BSE SIP registration number
	BSESIPRegNo string `json:"bse_sip_reg_no"`
	// BSEOrderNo is the BSE order number
	BSEOrderNo string `json:"bse_order_no"`
	// FundHouse is the fund house name
	FundHouse string `json:"fund_house"`
	// LumpsumType is the lumpsum type
	LumpsumType interface{} `json:"lumpsum_type"`
	// SIPFirstOrderFlag indicates if this is the first order
	SIPFirstOrderFlag string `json:"sip_firstorderflag"`
	// BSEPlacedOrderDate is when the order was placed with BSE
	BSEPlacedOrderDate string `json:"bse_placed_order_date"`
	// GoalID is the goal identifier
	GoalID interface{} `json:"goal_id"`
	// Units is the number of units (usually null)
	Units interface{} `json:"units"`
	// AllUnitsFlag indicates if all units should be used
	AllUnitsFlag interface{} `json:"all_units_flag"`
	// PaymentGatewayID is the payment gateway identifier
	PaymentGatewayID interface{} `json:"payment_gateway_id"`
	// LockVersion is the lock version for concurrency control
	LockVersion int `json:"lock_version"`
	// UpsizeCode is the upsize code
	UpsizeCode string `json:"upsize_code"`
}

// Holding represents a single fund holding with all its details.
type Holding struct {
	// FolioNumber is the folio number for this holding
	FolioNumber string `json:"folioNumber"`
	// AllottedAmount is the total amount allotted/invested
	AllottedAmount float64 `json:"allottedAmount"`
	// LockFreeUnits is the number of lock-free units
	LockFreeUnits float64 `json:"lock_free_units"`
	// Units is the total number of units owned
	Units float64 `json:"units"`
	// XIRRDates contains the dates for XIRR calculation
	XIRRDates []string `json:"xirr_dates"`
	// XIRRValues contains the values for XIRR calculation
	XIRRValues []float64 `json:"xirr_values"`
	// IsSip indicates if this is a SIP investment
	IsSip bool `json:"isSip"`
	// KuveraCategory is the Kuvera categorization
	KuveraCategory string `json:"kuvera_category"`
	// Direct indicates if this is a direct fund
	Direct bool `json:"direct"`
	// OrderDetails contains all order/transaction details
	OrderDetails []OrderDetail `json:"order_details"`
	// Reason contains any reason (usually empty)
	Reason interface{} `json:"reason"`
	// ValidFlag indicates if the holding is valid
	ValidFlag string `json:"valid_flag"`
	// Source indicates the source of the holding
	Source string `json:"source"`
	// SIPs contains SIP details if applicable
	SIPs []SIPDetail `json:"sips,omitempty"`
}

// HoldingsResponse represents the response from the holdings API endpoint.
// The response is a map where keys are fund codes and values are arrays of holdings.
type HoldingsResponse map[string][]Holding

// GoldTaxes represents tax information for gold trading.
type GoldTaxes struct {
	// CGST is the Central Goods and Services Tax percentage
	CGST float64 `json:"cgst"`
	// SGST is the State Goods and Services Tax percentage
	SGST float64 `json:"sgst"`
	// IGST is the Integrated Goods and Services Tax percentage
	IGST float64 `json:"igst"`
}

// CurrentGoldPrice represents buy/sell prices for gold.
type CurrentGoldPrice struct {
	// Buy is the current buy price per gram
	Buy float64 `json:"buy"`
	// Sell is the current sell price per gram
	Sell float64 `json:"sell"`
}

// GoldPriceResponse represents the response from the gold price API endpoint.
type GoldPriceResponse struct {
	// Taxes contains tax information for gold trading
	Taxes GoldTaxes `json:"taxes"`
	// BlockID is a unique identifier for this price block
	BlockID string `json:"block_id"`
	// FetchedAt is when the price was fetched
	FetchedAt string `json:"fetched_at"`
	// CurrentGoldPrice contains the current buy/sell prices
	CurrentGoldPrice CurrentGoldPrice `json:"current_gold_price"`
}

// NewClient creates a new Kuvera API client with the given options.
//
// Default configuration:
//   - BaseURL: Official Kuvera API endpoint
//   - Timeout: 30 seconds
//   - UserAgent: unofficial-kuvera-api/1.0
//
// Example:
//
//	client := kuvera.NewClient()
//	resp, err := client.Login(ctx, "username", "password")
//
// With custom options:
//
//	client := kuvera.NewClient(
//		kuvera.WithTimeout(60*time.Second),
//		kuvera.WithUserAgent("my-app/1.0"),
//	)
func NewClient(options ...ClientOption) KuveraClient {
	config := &clientConfig{
		baseURL:   BaseURL,
		userAgent: DefaultUserAgent,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, option := range options {
		option(config)
	}

	return &Client{
		baseURL:    config.baseURL,
		httpClient: config.httpClient,
		userAgent:  config.userAgent,
	}
}

// makeRequest is an internal helper method that handles HTTP request creation and execution.
// It automatically adds all necessary headers including authentication.
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, payload interface{}) (*http.Response, error) {
	// Validate URL
	apiURL, err := url.JoinPath(c.baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to match browser request
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	// Don't set Accept-Encoding to avoid compression issues
	if payload != nil {
		req.Header.Set("Content-Type", "application/json;charset=utf-8")
	}
	req.Header.Set("Origin", "https://kuvera.in")
	req.Header.Set("Referer", "https://kuvera.in/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	// Add authentication headers if available
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	} else {
		req.Header.Set("Authorization", "Bearer")
	}
	if c.sessionID != "" {
		req.Header.Set("X-Session-ID", c.sessionID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// handleResponse is an internal helper method that processes HTTP responses.
// It handles response body reading, JSON unmarshaling, and status code validation.
func (c *Client) handleResponse(resp *http.Response, result interface{}, operation string) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Debug: Uncomment the lines below for troubleshooting API responses
	// fmt.Printf("DEBUG %s Response Status: %d\n", operation, resp.StatusCode)
	// fmt.Printf("DEBUG %s Response Body: %s\n", operation, string(body))

	// Try to parse as JSON first
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse response (body: %s): %w", string(body), err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		// Try to extract API error details
		var apiErr APIError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Code != 0 {
			return &apiErr
		}
		return fmt.Errorf("%s failed with status code: %d", operation, resp.StatusCode)
	}

	return nil
}

// Login authenticates the user with Kuvera and stores the access token for subsequent requests.
//
// The method sends a POST request to the authentication endpoint with the provided
// credentials. On successful authentication, the access token is automatically stored
// in the client and will be included in all subsequent API calls.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - username: The user's Kuvera username/email
//   - password: The user's Kuvera password
//
// Returns:
//   - LoginResponse: Contains access token, user ID, and any error details
//   - error: Any network, parsing, authentication, or validation errors
//
// Example:
//
//	ctx := context.Background()
//	client := kuvera.NewClient()
//	resp, err := client.Login(ctx, "user@example.com", "mypassword")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Logged in successfully. User ID: %s\n", resp.Data.UserID)
func (c *Client) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	// Input validation
	if strings.TrimSpace(username) == "" {
		return nil, ErrEmptyUsername
	}
	if strings.TrimSpace(password) == "" {
		return nil, ErrEmptyPassword
	}

	loginReq := LoginRequest{
		Email:    username,
		Password: password,
		V:        "1.239.2",
	}

	resp, err := c.makeRequest(ctx, "POST", "/api/v5/users/authenticate.json", loginReq)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}

	var loginResp LoginResponse

	// Handle response parsing
	if err := c.handleResponse(resp, &loginResp, "login"); err != nil {
		// Check for specific login error messages
		if loginResp.Error != "" || loginResp.Status != "success" {
			return &loginResp, ErrInvalidCredentials
		}
		return &loginResp, err
	}

	// Store access token in client for subsequent requests
	c.accessToken = loginResp.Token

	return &loginResp, nil
}

// GetPortfolio retrieves complete portfolio data including all investments.
//
// This method fetches comprehensive portfolio data including mutual funds,
// gold, fixed deposits, Indian equities, and overall portfolio performance.
// The user must be authenticated (logged in) before calling this method.
//
// Returns:
//   - PortfolioResponse: Contains complete portfolio data
//   - error: Authentication errors, network errors, or API errors
//
// Example:
//
//	portfolio, err := client.GetPortfolio(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Total portfolio value: ₹%.2f\n", portfolio.Data.CurrentValue)
//	fmt.Printf("Mutual funds value: ₹%.2f\n", portfolio.Data.MutualFunds.CurrentValue)
//	fmt.Printf("Overall gain: %.2f%%\n", portfolio.Data.CurrentGainPercent)
func (c *Client) GetPortfolio(ctx context.Context) (*PortfolioResponse, error) {
	if c.accessToken == "" {
		return nil, ErrNotAuthenticated
	}

	resp, err := c.makeRequest(ctx, "GET", "/api/v5/portfolio/returns.json", nil)
	if err != nil {
		return nil, fmt.Errorf("portfolio request failed: %w", err)
	}

	var portfolioResp PortfolioResponse
	if err := c.handleResponse(resp, &portfolioResp, "portfolio"); err != nil {
		return &portfolioResp, err
	}

	return &portfolioResp, nil
}

// GetHoldings retrieves detailed holdings information for all mutual funds.
//
// This method fetches comprehensive details for each fund holding including
// folio numbers, units owned, order details, SIP information, and transaction
// history. The user must be authenticated (logged in) before calling this method.
//
// Returns:
//   - HoldingsResponse: Contains detailed holdings information organized by fund code
//   - error: Authentication errors, network errors, or API errors
//
// Example:
//
//	holdings, err := client.GetHoldings(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for fundCode, fundHoldings := range holdings {
//		for _, holding := range fundHoldings {
//			fmt.Printf("Fund %s - Folio: %s, Units: %.3f, Amount: ₹%.2f\n",
//				fundCode, holding.FolioNumber, holding.Units, holding.AllottedAmount)
//		}
//	}
func (c *Client) GetHoldings(ctx context.Context) (*HoldingsResponse, error) {
	if c.accessToken == "" {
		return nil, ErrNotAuthenticated
	}

	resp, err := c.makeRequest(ctx, "GET", "/api/v3/portfolio/holdings.json", nil)
	if err != nil {
		return nil, fmt.Errorf("holdings request failed: %w", err)
	}

	var holdingsResp HoldingsResponse
	if err := c.handleResponse(resp, &holdingsResp, "holdings"); err != nil {
		return &holdingsResp, err
	}

	return &holdingsResp, nil
}

// GetGoldPrice retrieves the current gold price information from Kuvera's partner.
//
// This method fetches current gold buy/sell prices in INR per gram along with
// tax information (CGST, SGST, IGST). This endpoint requires authentication.
//
// Returns:
//   - GoldPriceResponse: Contains current gold buy/sell prices and tax info
//   - error: Authentication errors, network errors, or API errors
//
// Example:
//
//	goldPrice, err := client.GetGoldPrice(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Gold buy: ₹%.2f, sell: ₹%.2f per gram\n",
//		goldPrice.CurrentGoldPrice.Buy, goldPrice.CurrentGoldPrice.Sell)
func (c *Client) GetGoldPrice(ctx context.Context) (*GoldPriceResponse, error) {
	if c.accessToken == "" {
		return nil, ErrNotAuthenticated
	}

	// Add query parameters as required by the API
	endpoint := "/api/v3/gold/current_price.json?v=1.239.2&cached=true"
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("gold price request failed: %w", err)
	}

	var goldResp GoldPriceResponse
	if err := c.handleResponse(resp, &goldResp, "gold price"); err != nil {
		return &goldResp, err
	}

	return &goldResp, nil
}
