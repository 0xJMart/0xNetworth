package main

import (
	"fmt"
	"log"
	"os"

	"0xnetworth/backend/internal/integrations/coinbase"
)

func main() {
	// Get API keys from environment
	apiKeyName := os.Getenv("COINBASE_API_KEY_NAME")
	apiPrivateKey := os.Getenv("COINBASE_API_PRIVATE_KEY")

	if apiKeyName == "" || apiPrivateKey == "" {
		log.Fatal("Error: COINBASE_API_KEY_NAME and COINBASE_API_PRIVATE_KEY environment variables must be set")
	}

	// Create client
	client, err := coinbase.NewClient(apiKeyName, apiPrivateKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Test JWT generation for Advanced Trade API
	method := "GET"
	path := "/api/v3/brokerage/accounts"

	fmt.Println("=== Testing JWT Generation (CDP API v2 Format) ===")
	fmt.Printf("Method: %s\n", method)
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("API Key Name: %s\n", apiKeyName)
	fmt.Println()

	// Generate JWT
	jwt, err := client.GenerateJWT(method, path)
	if err != nil {
		log.Fatalf("Failed to generate JWT: %v", err)
	}

	fmt.Println("✓ JWT generated successfully!")
	fmt.Println()
	fmt.Println("Generated JWT:")
	fmt.Println(jwt)
	fmt.Println()

	// Show how to test with curl
	fmt.Println("=== Test with curl ===")
	fmt.Println("You can test this JWT with the following curl command:")
	fmt.Println()
	fmt.Printf("export JWT='%s'\n", jwt)
	fmt.Println()
	fmt.Printf("curl -L -X %s 'https://api.coinbase.com%s' \\\n", method, path)
	fmt.Println("  -H \"Authorization: Bearer $JWT\" \\")
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Println("  -H \"Accept: application/json\"")
	fmt.Println()

	// Also test by making an actual request
	fmt.Println("=== Testing with actual API request ===")
	accounts, err := client.GetAccounts()
	if err != nil {
		fmt.Printf("⚠ Request failed: %v\n", err)
		fmt.Println()
		fmt.Println("This could mean:")
		fmt.Println("  - API keys are invalid or expired")
		fmt.Println("  - API keys don't have required permissions")
		fmt.Println("  - Network connectivity issue")
		fmt.Println()
		fmt.Println("However, if JWT generation succeeded above, the JWT format is correct!")
	} else {
		fmt.Printf("✓ Success! Retrieved %d accounts\n", len(accounts))
	}
}

