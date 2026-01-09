package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const docsDir = "./internal/clients/tradernet/docs"

func main() {
	fmt.Println("Converting Tradernet API Documentation to Markdown...")
	fmt.Println("====================================================\n")

	totalFiles := 0
	successCount := 0
	failedFiles := []string{}

	// Walk through all HTML files
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-HTML files
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		totalFiles++
		relPath := strings.TrimPrefix(path, docsDir+"/")
		fmt.Printf("Converting: %s... ", relPath)

		// Read HTML file
		htmlContent, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("FAILED (read error: %v)\n", err)
			failedFiles = append(failedFiles, relPath)
			return nil
		}

		// Convert to Markdown
		markdown := convertHTMLToMarkdown(string(htmlContent))

		// Write Markdown file
		mdPath := strings.TrimSuffix(path, ".html") + ".md"
		if err := os.WriteFile(mdPath, []byte(markdown), 0644); err != nil {
			fmt.Printf("FAILED (write error: %v)\n", err)
			failedFiles = append(failedFiles, relPath)
			return nil
		}

		// Remove old HTML file
		if err := os.Remove(path); err != nil {
			fmt.Printf("WARNING (could not remove HTML: %v)\n", err)
		}

		successCount++
		fmt.Println("OK")
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	// Update README
	if err := updateREADME(); err != nil {
		fmt.Printf("Error updating README: %v\n", err)
	} else {
		fmt.Println("\nREADME.md updated with .md links")
	}

	// Print summary
	fmt.Println("\n====================================================")
	fmt.Printf("Total files: %d\n", totalFiles)
	fmt.Printf("Successfully converted: %d\n", successCount)
	fmt.Printf("Failed: %d\n", len(failedFiles))

	if len(failedFiles) > 0 {
		fmt.Println("\nFailed files:")
		for _, file := range failedFiles {
			fmt.Printf("  - %s\n", file)
		}
	}
}

func convertHTMLToMarkdown(html string) string {
	var md strings.Builder

	// Clean up email protection
	html = regexp.MustCompile(`<a[^>]*__cf_email__[^>]*>.*?</a>`).ReplaceAllString(html, "[email protected]")
	html = regexp.MustCompile(`data-cfemail="[^"]*"`).ReplaceAllString(html, "")

	// Remove the outer div wrapper
	html = regexp.MustCompile(`(?s)^<div[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?s)</div>\s*$`).ReplaceAllString(html, "")

	// Split by major block elements while preserving them
	// Process line by line to maintain structure
	lines := strings.Split(html, "\n")

	inCodeBlock := false
	inTable := false
	inTableHeader := false
	inList := false
	codeContent := ""
	currentCodeLang := "json"

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines outside code blocks
		if trimmedLine == "" && !inCodeBlock {
			continue
		}

		// Handle h1
		if strings.Contains(trimmedLine, "<h1>") {
			text := extractText(trimmedLine)
			md.WriteString("# " + text + "\n\n")
			continue
		}

		// Handle h2
		if strings.Contains(trimmedLine, "<h2>") {
			text := extractText(trimmedLine)
			md.WriteString("## " + text + "\n\n")
			continue
		}

		// Handle h3
		if strings.Contains(trimmedLine, "<h3>") {
			text := extractText(trimmedLine)
			md.WriteString("### " + text + "\n\n")
			continue
		}

		// Handle h4
		if strings.Contains(trimmedLine, "<h4>") {
			text := extractText(trimmedLine)
			md.WriteString("#### " + text + "\n\n")
			continue
		}

		// Handle paragraphs
		if strings.Contains(trimmedLine, "<p>") && !strings.Contains(trimmedLine, "<span class=\"uk-button\"") {
			text := extractText(trimmedLine)
			if text != "" {
				md.WriteString(text + "\n\n")
			}
			continue
		}

		// Handle dfn (definitions)
		if strings.Contains(trimmedLine, "<dfn>") {
			text := extractText(trimmedLine)
			if text != "" {
				md.WriteString("**" + text + "**\n\n")
			}
			continue
		}

		// Handle code blocks
		if strings.Contains(trimmedLine, "<code") {
			inCodeBlock = true
			// Detect language from class
			if strings.Contains(line, "class=\"javascript") {
				currentCodeLang = "javascript"
			} else if strings.Contains(line, "class=\"python") {
				currentCodeLang = "python"
			} else if strings.Contains(line, "class=\"php") {
				currentCodeLang = "php"
			} else if strings.Contains(line, "class=\"bash") || strings.Contains(line, "cURL") {
				currentCodeLang = "bash"
			} else {
				currentCodeLang = "json"
			}
			continue
		}

		if strings.Contains(trimmedLine, "<pre>") {
			continue
		}

		if inCodeBlock {
			if strings.Contains(trimmedLine, "</pre>") || strings.Contains(trimmedLine, "</code>") {
				// End of code block
				if strings.Contains(trimmedLine, "</pre>") && !strings.Contains(trimmedLine, "</code>") {
					continue // Just closing pre, code tag still open
				}
				inCodeBlock = false
				if codeContent != "" {
					md.WriteString("```" + currentCodeLang + "\n")
					md.WriteString(cleanCode(codeContent))
					md.WriteString("\n```\n\n")
					codeContent = ""
				}
				currentCodeLang = "json"
				continue
			}
			// Accumulate code content - preserve original indentation
			codeContent += line + "\n"
			continue
		}

		// Handle tables
		if strings.Contains(trimmedLine, "<table") {
			inTable = true
			continue
		}

		if strings.Contains(trimmedLine, "</table>") {
			inTable = false
			md.WriteString("\n")
			continue
		}

		if inTable {
			if strings.Contains(trimmedLine, "<caption>") {
				text := extractText(trimmedLine)
				md.WriteString("**" + text + "**\n\n")
				continue
			}

			if strings.Contains(trimmedLine, "<thead>") {
				inTableHeader = true
				continue
			}

			if strings.Contains(trimmedLine, "</thead>") {
				inTableHeader = false
				continue
			}

			if strings.Contains(trimmedLine, "<tbody>") || strings.Contains(trimmedLine, "</tbody>") {
				continue
			}

			if strings.Contains(trimmedLine, "<tr>") {
				continue
			}

			if strings.Contains(trimmedLine, "</tr>") {
				md.WriteString("\n")
				// Add separator after header row
				if inTableHeader {
					md.WriteString("|---|---|---|---|\n")
				}
				continue
			}

			if strings.Contains(trimmedLine, "<td>") || strings.Contains(trimmedLine, "<th>") {
				text := extractText(trimmedLine)
				if text == "" {
					text = " "
				}
				md.WriteString("| " + text + " ")
				continue
			}
		}

		// Handle lists (ul/li for code examples)
		if strings.Contains(trimmedLine, "<ul") && strings.Contains(trimmedLine, "uk-tab") {
			// This is the tab switcher for examples, skip it
			inList = true
			md.WriteString("## Examples\n\n")
			continue
		}

		if inList && strings.Contains(trimmedLine, "</ul>") {
			inList = false
			continue
		}

		if inList && strings.Contains(trimmedLine, "<li>") {
			// Check for language heading in next lines
			for j := i + 1; j < len(lines) && j < i+5; j++ {
				nextLine := strings.TrimSpace(lines[j])
				if strings.Contains(nextLine, "<h3>Browser</h3>") {
					md.WriteString("### Browser (JavaScript)\n\n")
					i = j
					break
				} else if strings.Contains(nextLine, "<h3>NodeJS</h3>") {
					md.WriteString("### NodeJS\n\n")
					i = j
					break
				} else if strings.Contains(nextLine, "<h3>Python</h3>") {
					md.WriteString("### Python\n\n")
					i = j
					break
				} else if strings.Contains(nextLine, "<h3>PHP</h3>") {
					md.WriteString("### PHP\n\n")
					i = j
					break
				} else if strings.Contains(nextLine, "<h3>cURL</h3>") {
					md.WriteString("### cURL\n\n")
					i = j
					break
				} else if strings.Contains(nextLine, "<h3>JS (jQuery)</h3>") {
					md.WriteString("### JavaScript (jQuery)\n\n")
					i = j
					break
				}
			}
			continue
		}

		// Skip remaining HTML tags
		if strings.Contains(trimmedLine, "<br>") || strings.Contains(trimmedLine, "</li>") ||
			strings.Contains(trimmedLine, "<span") || strings.Contains(trimmedLine, "</a>") ||
			strings.Contains(trimmedLine, "<a href") {
			continue
		}
	}

	return md.String()
}

func extractText(html string) string {
	// Remove all HTML tags
	text := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(html, "")
	// Decode HTML entities
	text = decodeHTMLEntities(text)
	// Trim whitespace
	text = strings.TrimSpace(text)
	return text
}

func cleanCode(code string) string {
	// Remove HTML tags
	code = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(code, "")
	// Decode HTML entities
	code = decodeHTMLEntities(code)

	// Split into lines
	lines := strings.Split(code, "\n")

	// Remove leading/trailing empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) == 0 {
		return ""
	}

	// Find minimum indentation (ignoring empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := countLeadingSpaces(line)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent == -1 {
		minIndent = 0
	}

	// Remove common indentation and trailing whitespace
	var cleaned []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			// Preserve empty lines but remove all whitespace
			cleaned = append(cleaned, "")
		} else {
			// Remove common indentation and trailing whitespace
			if len(line) > minIndent {
				line = line[minIndent:]
			}
			line = strings.TrimRight(line, " \t")
			cleaned = append(cleaned, line)
		}
	}

	return strings.Join(cleaned, "\n")
}

func countLeadingSpaces(s string) int {
	count := 0
	for _, ch := range s {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4 // Treat tab as 4 spaces
		} else {
			break
		}
	}
	return count
}

func decodeHTMLEntities(text string) string {
	replacements := map[string]string{
		"&nbsp;":   " ",
		"&amp;":    "&",
		"&lt;":     "<",
		"&gt;":     ">",
		"&quot;":   "\"",
		"&#160;":   " ",
		"&lsaquo;": "<",
		"&rsaquo;": ">",
		"&laquo;":  "«",
		"&raquo;":  "»",
		"&apos;":   "'",
		"&#39;":    "'",
		"&#x27;":   "'",
		"&mdash;":  "—",
		"&ndash;":  "–",
		"&hellip;": "...",
		"&times;":  "×",
		"&divide;": "÷",
		"&copy;":   "©",
		"&reg;":    "®",
		"&trade;":  "™",
		"&euro;":   "€",
		"&pound;":  "£",
		"&yen;":    "¥",
	}

	for entity, replacement := range replacements {
		text = strings.ReplaceAll(text, entity, replacement)
	}

	return text
}

func updateREADME() error {
	readmePath := filepath.Join(docsDir, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	// Replace .html with .md in all links
	updated := regexp.MustCompile(`\.html\)`).ReplaceAllString(string(content), ".md)")

	return os.WriteFile(readmePath, []byte(updated), 0644)
}
