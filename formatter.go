package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func formatProxy(proxy string, format int) string {
	parts := strings.Split(proxy, ":")
	if len(parts) != 4 {
		return "Invalid proxy format: " + proxy
	}

	formattedProxy := fmt.Sprintf("%s:%s@%s:%s", parts[2], parts[3], parts[0], parts[1])

	switch format {
	case 1:
		return "socks5://" + formattedProxy
	case 2:
		return "http://" + formattedProxy
	default:
		return formattedProxy
	}
}

func main() {
	fmt.Println("Choice format input:")
	fmt.Println("1. socks5://")
	fmt.Println("2. http://")
	fmt.Println("3. none (just saved it to formatted.txt)")

	var choice int
	fmt.Print("Enter your choice (1-3): ")
	fmt.Scan(&choice)

	if choice < 1 || choice > 3 {
		fmt.Println("Invalid choice!")
		return
	}

	file, err := os.Open("proxy.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	outputFile, err := os.Create("formatted.txt")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxy := scanner.Text()
		formattedProxy := formatProxy(proxy, choice)

		_, err := writer.WriteString(formattedProxy + "\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	}

	writer.Flush()

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	fmt.Println("Process completed!")
}
