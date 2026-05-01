package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type createOrderRequest struct {
	Products []orderProduct `json:"products"`
}

type orderProduct struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type productResponse struct {
	ID        string    `json:"id"`
	ProductID string    `json:"productId"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type orderResponse struct {
	ID            string            `json:"id"`
	Products      []productResponse `json:"products"`
	Status        string            `json:"status"`
	FailureReason string            `json:"failureReason"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
}

type apiError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type orderResult struct {
	Index  int
	ID     string
	Status string
	Error  string
}

func main() {
	gatewayURL := flag.String("url", "http://localhost:8080", "API gateway URL")
	productID := flag.String("product-id", "2f26929d-ef0f-45d1-a60d-5b5bec402d53", "Product ID to order")
	concurrency := flag.Int("concurrency", 3, "Number of concurrent requests")
	timeout := flag.Duration("timeout", 30*time.Second, "Max time to wait for saga completion")
	flag.Parse()

	fmt.Println("=== Concurrency Race Condition Test ===")
	fmt.Printf("Gateway:     %s\n", *gatewayURL)
	fmt.Printf("Product ID:  %s\n", *productID)
	fmt.Printf("Concurrency: %d\n", *concurrency)
	fmt.Println()

	orderIDs := sendConcurrentOrders(*gatewayURL, *productID, *concurrency)
	if len(orderIDs) == 0 {
		fmt.Println("ERROR: No orders were created successfully.")
		os.Exit(1)
	}
	fmt.Println()

	finalResults := waitForCompletion(*gatewayURL, orderIDs, *timeout)
	fmt.Println()

	printReport(finalResults, *concurrency)
}

func sendConcurrentOrders(gatewayURL, productID string, count int) []string {
	fmt.Printf("Sending %d concurrent order requests...\n", count)

	type httpResult struct {
		index   int
		orderID string
		status  string
		errMsg  string
	}

	results := make([]httpResult, count)
	var wg sync.WaitGroup

	barrier := make(chan struct{})

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-barrier

			body := createOrderRequest{
				Products: []orderProduct{
					{ProductID: productID, Quantity: 1},
				},
			}

			jsonBody, err := json.Marshal(body)
			if err != nil {
				results[idx] = httpResult{index: idx, errMsg: fmt.Sprintf("marshal error: %v", err)}
				return
			}

			resp, err := http.Post(gatewayURL+"/api/orders", "application/json", bytes.NewReader(jsonBody))
			if err != nil {
				results[idx] = httpResult{index: idx, errMsg: fmt.Sprintf("HTTP error: %v", err)}
				return
			}
			defer resp.Body.Close()

			respBody, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				var apiErr apiError
				_ = json.Unmarshal(respBody, &apiErr)
				results[idx] = httpResult{
					index:  idx,
					errMsg: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, apiErr.Error),
				}
				return
			}

			var orderResp orderResponse
			if err := json.Unmarshal(respBody, &orderResp); err != nil {
				results[idx] = httpResult{
					index:  idx,
					errMsg: fmt.Sprintf("parse error: %v", err),
				}
				return
			}

			results[idx] = httpResult{
				index:   idx,
				orderID: orderResp.ID,
				status:  orderResp.Status,
			}
		}(i)
	}

	time.Sleep(100 * time.Millisecond)
	close(barrier)
	wg.Wait()

	var orderIDs []string
	for i, r := range results {
		if r.orderID != "" {
			fmt.Printf("  Request %d: created (id: %s) status: %s\n", i+1, shortID(r.orderID), r.status)
			orderIDs = append(orderIDs, r.orderID)
		} else {
			fmt.Printf("  Request %d: FAILED - %s\n", i+1, r.errMsg)
		}
	}

	return orderIDs
}

func waitForCompletion(gatewayURL string, orderIDs []string, timeout time.Duration) []orderResult {
	fmt.Printf("Waiting for saga completion (timeout: %v)...\n", timeout)

	results := make([]orderResult, len(orderIDs))
	terminalStatuses := map[string]bool{
		"confirmed": true,
		"failed":    true,
		"cancelled": true,
	}

	var wg sync.WaitGroup
	for i, id := range orderIDs {
		wg.Add(1)
		go func(idx int, orderID string) {
			defer wg.Done()
			deadline := time.Now().Add(timeout)

			for time.Now().Before(deadline) {
				result := pollOrder(gatewayURL, orderID)
				if terminalStatuses[result.Status] {
					results[idx] = result
					return
				}
				time.Sleep(500 * time.Millisecond)
			}
			results[idx] = orderResult{
				Index:  idx,
				ID:     orderID,
				Status: "timeout",
				Error:  fmt.Sprintf("did not reach terminal state within %v", timeout),
			}
		}(i, id)
	}

	wg.Wait()

	for _, r := range results {
		switch r.Status {
		case "confirmed":
			fmt.Printf("  Order %s: %s\n", shortID(r.ID), green("CONFIRMED"))
		case "failed":
			fmt.Printf("  Order %s: %s (%s)\n", shortID(r.ID), red("FAILED"), r.Error)
		case "cancelled":
			fmt.Printf("  Order %s: %s\n", shortID(r.ID), yellow("CANCELLED"))
		case "timeout":
			fmt.Printf("  Order %s: %s\n", shortID(r.ID), yellow("TIMEOUT"))
		default:
			fmt.Printf("  Order %s: %s\n", shortID(r.ID), r.Status)
		}
	}

	return results
}

func pollOrder(gatewayURL, orderID string) orderResult {
	resp, err := http.Get(gatewayURL + "/api/orders/" + orderID)
	if err != nil {
		return orderResult{Index: 0, ID: orderID, Status: "error", Error: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var orderResp orderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return orderResult{Index: 0, ID: orderID, Status: "error", Error: "parse error"}
	}

	return orderResult{
		Index:  0,
		ID:     orderResp.ID,
		Status: orderResp.Status,
		Error:  orderResp.FailureReason,
	}
}

func printReport(results []orderResult, concurrency int) {
	confirmed := 0
	failed := 0
	cancelled := 0
	timedOut := 0
	other := 0

	for _, r := range results {
		switch r.Status {
		case "confirmed":
			confirmed++
		case "failed":
			failed++
		case "cancelled":
			cancelled++
		case "timeout":
			timedOut++
		default:
			other++
		}
	}

	fmt.Println("=== Results ===")
	fmt.Printf("Orders created:   %d\n", len(results))
	fmt.Printf("Orders confirmed: %d\n", confirmed)
	fmt.Printf("Orders failed:    %d\n", failed)
	fmt.Printf("Orders cancelled: %d\n", cancelled)
	fmt.Printf("Orders timed out: %d\n", timedOut)
	fmt.Println()

	if confirmed > 1 {
		fmt.Printf("%s: Race condition detected! %d orders confirmed but only 1 item was in stock.\n",
			red("FAIL"), confirmed)
		fmt.Println()
		fmt.Println("Root cause: inventory-service/internal/app/command/reserve_stock.go")
		fmt.Println("  - FindByProductIDs() reads stock without locks")
		fmt.Println("  - HasSufficientStock() check is not atomic with DeductStock()")
		fmt.Println("  - No SELECT FOR UPDATE, no transactions, no distributed locks")
		os.Exit(1)
	} else if confirmed == 1 {
		fmt.Printf("%s: Only 1 order confirmed. Race condition did NOT manifest this run.\n",
			green("PASS"))
		fmt.Println()
		fmt.Println("NOTE: This does NOT mean the system is safe.")
		fmt.Println("The race condition exists in the code but may not trigger due to")
		fmt.Println("message processing timing. Run multiple times or add artificial delays")
		fmt.Println("to increase the chance of observing the bug.")
		os.Exit(0)
	} else {
		fmt.Printf("%s: No orders confirmed. All were rejected/cancelled.\n",
			yellow("INCONCLUSIVE"))
		fmt.Println("The product may not exist or may already have 0 stock.")
		os.Exit(2)
	}
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func red(s string) string    { return fmt.Sprintf("\033[31m%s\033[0m", s) }
func green(s string) string  { return fmt.Sprintf("\033[32m%s\033[0m", s) }
func yellow(s string) string { return fmt.Sprintf("\033[33m%s\033[0m", s) }
