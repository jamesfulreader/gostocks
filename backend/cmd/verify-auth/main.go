package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const baseURL = "http://localhost:8080/api"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

type AddPortfolioRequest struct {
	Symbol string `json:"symbol"`
}

func main() {
	email := "test_go@example.com"
	password := "password123"

	// 1. Register
	fmt.Println("1. Registering user...")
	regReq := RegisterRequest{Email: email, Password: password}
	if err := sendRequest("POST", "/register", regReq, ""); err != nil {
		fmt.Printf("Registration failed (might already exist): %v\n", err)
	} else {
		fmt.Println("Registration successful")
	}

	// 2. Login
	fmt.Println("\n2. Logging in...")
	loginReq := LoginRequest{Email: email, Password: password}
	var loginResp LoginResponse
	respBody, err := sendRequestWithResponse("POST", "/login", loginReq, "")
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		fmt.Printf("Failed to parse login response: %v\n", err)
		os.Exit(1)
	}
	token := loginResp.Token
	fmt.Printf("Login successful, Token: %s...\n", token[:20])

	// 3. Get Portfolio (Empty)
	fmt.Println("\n3. Getting Portfolio (expect empty)...")
	if err := sendRequest("GET", "/portfolio", nil, token); err != nil {
		fmt.Printf("Failed to get portfolio: %v\n", err)
	}

	// 3.5 Fetch Quote (to populate DB)
	fmt.Println("\n3.5 Fetching Quote for AAPL (to populate DB)...")
	if err := sendRequest("GET", "/quote?symbol=AAPL", nil, token); err != nil {
		fmt.Printf("Failed to fetch quote: %v\n", err)
	}

	// 4. Add to Portfolio
	fmt.Println("\n4. Adding AAPL to Portfolio...")
	addReq := AddPortfolioRequest{Symbol: "AAPL"}
	if err := sendRequest("POST", "/portfolio", addReq, token); err != nil {
		fmt.Printf("Failed to add to portfolio: %v\n", err)
	} else {
		fmt.Println("Added AAPL successfully")
	}

	// 5. Get Portfolio (Should have AAPL)
	fmt.Println("\n5. Getting Portfolio (expect AAPL)...")
	if err := sendRequest("GET", "/portfolio", nil, token); err != nil {
		fmt.Printf("Failed to get portfolio: %v\n", err)
	}

	// 6. Remove from Portfolio
	fmt.Println("\n6. Removing AAPL from Portfolio...")
	if err := sendRequest("DELETE", "/portfolio?symbol=AAPL", nil, token); err != nil {
		fmt.Printf("Failed to remove from portfolio: %v\n", err)
	} else {
		fmt.Println("Removed AAPL successfully")
	}

	// 7. Get Portfolio (Should be empty)
	fmt.Println("\n7. Getting Portfolio (expect empty)...")
	if err := sendRequest("GET", "/portfolio", nil, token); err != nil {
		fmt.Printf("Failed to get portfolio: %v\n", err)
	}
}

func sendRequest(method, endpoint string, body interface{}, token string) error {
	_, err := sendRequestWithResponse(method, endpoint, body, token)
	return err
}

func sendRequestWithResponse(method, endpoint string, body interface{}, token string) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(respBytes))
	}

	fmt.Printf("Response (%d): %s\n", resp.StatusCode, string(respBytes))
	return respBytes, nil
}
