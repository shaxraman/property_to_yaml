package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type propLine struct {
	Keys  []string
	Value string
}

// orderedNode хранит порядок (для маршалинга в YAML)
type orderedNode struct {
	Pairs     []propLine
	SortAlpha bool
}

func (o orderedNode) MarshalYAML() (interface{}, error) {
	return buildTree(o.Pairs, o.SortAlpha), nil
}

func buildTree(pairs []propLine, sortAlpha bool) *yaml.Node {
	// Если все ключи пустые (значение на этом уровне)
	if len(pairs) == 1 && len(pairs[0].Keys) == 0 {
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: pairs[0].Value,
		}
	}

	type pair struct {
		Key   string
		Lines []propLine
	}
	var ordered []pair
	seen := map[string]bool{}
	for _, p := range pairs {
		if len(p.Keys) == 0 {
			continue
		}
		k := p.Keys[0]
		if !seen[k] {
			seen[k] = true
			var group []propLine
			for _, pp := range pairs {
				if len(pp.Keys) > 0 && pp.Keys[0] == k {
					group = append(group, propLine{Keys: pp.Keys[1:], Value: pp.Value})
				}
			}
			ordered = append(ordered, pair{Key: k, Lines: group})
		}
	}
	if sortAlpha {
		sort.Slice(ordered, func(i, j int) bool {
			return ordered[i].Key < ordered[j].Key
		})
	}

	// Собираем Ordered map для YAML
	node := yaml.Node{
		Kind:    yaml.MappingNode,
		Content: []*yaml.Node{},
	}
	for _, p := range ordered {
		keyNode := yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: p.Key,
		}
		val := buildTree(p.Lines, sortAlpha)
		node.Content = append(node.Content, &keyNode, val)
	}
	return &node
}

func main() {
	var pretty bool
	var sortMode string

	flag.BoolVar(&pretty, "pretty", false, "Insert blank lines between top-level YAML blocks")
	flag.StringVar(&sortMode, "sort", "original", "Sort mode: 'original' or 'alpha'")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: property_to_yaml [--pretty] [--sort original|alpha] <filename>")
		return
	}
	inputFile := args[0]
	outputFile := strings.TrimSuffix(inputFile, ".properties") + ".yaml"

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	var props []propLine
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			parts = strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		keys := strings.Split(key, ".")
		props = append(props, propLine{Keys: keys, Value: val})
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating YAML file: %v\n", err)
		return
	}
	defer outFile.Close()

	sortAlpha := strings.ToLower(sortMode) == "alpha"

	encoder := yaml.NewEncoder(outFile)
	encoder.SetIndent(2)
	err = encoder.Encode(orderedNode{Pairs: props, SortAlpha: sortAlpha})
	if err != nil {
		fmt.Printf("Error encoding YAML: %v\n", err)
		return
	}
	encoder.Close()

	if pretty {
		// yamlfmt-style: insert blank lines before each top-level key except the first
		// Read the generated YAML file, insert blank lines, write back
		yamlPath := outputFile
		yamlFile, err := os.Open(yamlPath)
		if err != nil {
			fmt.Printf("Error opening YAML file for formatting: %v\n", err)
			return
		}
		defer yamlFile.Close()
		var lines []string
		scanner = bufio.NewScanner(yamlFile)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading YAML file: %v\n", err)
			return
		}
		// Insert blank line before each top-level key except the first
		var result []string
		topLevelKeyCount := 0
		for i := 0; i < len(lines); i++ {
			line := lines[i]
			trimmed := strings.TrimSpace(line)
			isTop := trimmed == line && strings.HasSuffix(trimmed, ":") && len(trimmed) > 0
			if isTop {
				topLevelKeyCount++
				if topLevelKeyCount > 1 {
					// Insert blank line before this top-level key
					result = append(result, "")
				}
			}
			result = append(result, line)
		}
		// Write back
		yamlFileWrite, err := os.Create(yamlPath)
		if err != nil {
			fmt.Printf("Error rewriting YAML file: %v\n", err)
			return
		}
		defer yamlFileWrite.Close()
		for i, l := range result {
			if i > 0 {
				fmt.Fprintln(yamlFileWrite)
			}
			fmt.Fprint(yamlFileWrite, l)
		}
	}

	fmt.Printf("Converted '%s' to '%s' (sort: %s, pretty: %v)\n", inputFile, outputFile, sortMode, pretty)
}