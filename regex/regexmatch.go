package regex

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	MappingTree *Tree
)

func Init(hostfile string) {
	// Open the file
	file, err := os.Open(hostfile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	frontMatchMap := make(map[string]string)

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var count int16 = 0
	for scanner.Scan() {
		line := scanner.Text()
		count++

		// Split the line by whitespace
		fields := strings.Fields(line)
		if len(fields) == 2 {
			ip := fields[0]
			domain := fields[1]
			if strings.HasPrefix(domain, "*") {
				frontMatchMap[domain] = ip
			}
		}
	}
	log.Printf("Read hostfile %d lines\n", count)

	// Check for any scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading hostfile:", err)
		return
	}

	// Print the domain and IP mapping
	/*
		for domain, ip := range frontMatchMap {
			fmt.Printf("Domain: [%s], IP: [%s]\n", domain, ip)
		}
	*/

	MappingTree = ConstructTreeReverse(frontMatchMap)
	// PrintTree(MappingTree)
}
