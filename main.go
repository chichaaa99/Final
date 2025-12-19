package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type StatResult struct {
	Name  string
	Value string
	Error error
}

func calculateStat(numbers []float64, stat string) StatResult {
	var rCode string
	switch stat {
	case "mean":
		rCode = fmt.Sprintf("cat(mean(c(%s)), '\\n')", joinFloats(numbers, ","))
	case "median":
		rCode = fmt.Sprintf("cat(median(c(%s)), '\\n')", joinFloats(numbers, ","))
	case "sd":
		rCode = fmt.Sprintf("cat(sd(c(%s)), '\\n')", joinFloats(numbers, ","))
	case "var":
		rCode = fmt.Sprintf("cat(var(c(%s)), '\\n')", joinFloats(numbers, ","))
	case "min":
		rCode = fmt.Sprintf("cat(min(c(%s)), '\\n')", joinFloats(numbers, ","))
	case "max":
		rCode = fmt.Sprintf("cat(max(c(%s)), '\\n')", joinFloats(numbers, ","))
	default:
		return StatResult{Name: stat, Error: fmt.Errorf("unknown stat: %s", stat)}
	}

	cmd := exec.Command("Rscript", "-e", rCode)
	output, err := cmd.Output()
	if err != nil {
		return StatResult{Name: stat, Error: err}
	}
	return StatResult{Name: stat, Value: strings.TrimSpace(string(output))}
}

func joinFloats(nums []float64, sep string) string {
	strs := make([]string, len(nums))
	for i, num := range nums {
		strs[i] = strconv.FormatFloat(num, 'f', -1, 64)
	}
	return strings.Join(strs, sep)
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter numbers separated by spaces: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	parts := strings.Fields(input)
	numbers := make([]float64, 0, len(parts))
	for _, part := range parts {
		num, err := strconv.ParseFloat(part, 64)
		if err != nil {
			fmt.Printf("Invalid number: %s\n", part)
			return
		}
		numbers = append(numbers, num)
	}

	if len(numbers) == 0 {
		fmt.Println("No numbers provided")
		return
	}

	stats := []string{"mean", "median", "sd", "var", "min", "max"}

	var wg sync.WaitGroup
	results := make(chan StatResult, len(stats))

	for _, stat := range stats {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			results <- calculateStat(numbers, s)
		}(stat)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.Error != nil {
			fmt.Printf("Error calculating %s: %v\n", result.Name, result.Error)
		} else {
			fmt.Printf("%s: %s\n", result.Name, result.Value)
		}
	}
}
