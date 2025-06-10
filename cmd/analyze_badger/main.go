// cmd/analyze_badger/main.go

/*
 * This tool provides a quick way to analyze multiple BadgerDB directories
 * and summarize the contents in a human-readable format for debugging
 */
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
)

func main() {

	// Point to the BadgerDB directories as command line arguments
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <badger_dir1> <badger_dir2> ...\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	// Form: nodeID -> bucket (nodeID|name|labels) -> count
	type statmap map[string]map[string]int
	stats := make(statmap)

	// Iterate over each directory provided
	for _, dir := range os.Args[1:] {
		nodeID := filepath.Base(dir)
		stats[nodeID] = make(map[string]int)

		// Load the BadgerDB at the given directory
		opts := badger.DefaultOptions(dir).WithReadOnly(true)
		db, err := badger.Open(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open BadgerDB at %s: %v\n", dir, err)
			continue
		}
		defer db.Close()

		// Iterate through all items in the BadgerDB and count occurrences of each bucket (nodeID|name|labels)
		db.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				k := string(it.Item().Key())
				// Expect k == nodeID|name|labels|ts
				parts := strings.SplitN(k, "|", 4)
				if len(parts) >= 3 {
					bucket := fmt.Sprintf("%s|%s|%s", parts[0], parts[1], parts[2])
					stats[nodeID][bucket]++
				}
			}
			return nil
		})
	}

	// Print a per-node table
	for node, buckets := range stats {
		fmt.Printf("\n--- Node: %s ---\n", node)
		fmt.Printf("%-50s  %s\n", "MetricKey (nodeID|name|labels)", "Count")

		// Sorted output for determinism
		keys := make([]string, 0, len(buckets))
		for k := range buckets {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Print each bucket and its count
		for _, bucket := range keys {
			fmt.Printf("%-50s  %d\n", bucket, buckets[bucket])
		}
	}

	// Global cross-reference: which nodes have each bucket
	bucketMap := make(map[string][]string)
	for node, buckets := range stats {
		for bucket := range buckets {
			bucketMap[bucket] = append(bucketMap[bucket], node)
		}
	}
	fmt.Printf("\n--- Global Replica Map (bucket -> nodes) ---\n")

	// Sorted output
	allBuckets := make([]string, 0, len(bucketMap))
	for bucket := range bucketMap {
		allBuckets = append(allBuckets, bucket)
	}
	sort.Strings(allBuckets)
	for _, bucket := range allBuckets {
		nodes := bucketMap[bucket]
		sort.Strings(nodes)
		fmt.Printf("%-50s: %s\n", bucket, strings.Join(nodes, ", "))
	}
}
