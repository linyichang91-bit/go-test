// main.go
package main

import (
	"fmt"
	"net/http"
)

// 可测试逻辑（重点）
func Add(a, b int) int {
	return a + b
}

func handler(w http.ResponseWriter, r *http.Request) {
	result := Add(1, 2)
	fmt.Fprintf(w, "result: %d", result)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
