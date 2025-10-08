package transforms

import (
	"os"
	"strings"

	"github.com/xh-dev-go/hosts-control/interfaces"
	"github.com/xh-dev-go/hosts-control/iputils"
)

// LoadFileToString reads the entire content of the file at the given path into a string.
func LoadFileToString(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func LoadStringToHostsFile(content string) (*interfaces.HostsFile, error) {
	hf := &interfaces.HostsFile{
		Lines:    []interfaces.HostLine{},
		ByIP:     make(map[string]*interfaces.HostLine),
		ByDomain: make(map[string]*interfaces.HostLine),
	}

	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		entry := interfaces.HostLine{Raw: rawLine}

		// Blank line
		if line == "" {
			entry.IsBlank = true
			hf.Lines = append(hf.Lines, entry)
			continue
		}

		// Full-line comment
		if strings.HasPrefix(line, "#") {
			entry.Comment = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			hf.Lines = append(hf.Lines, entry)
			continue
		}

		// Inline comment
		comment := ""
		if hashIndex := strings.Index(line, "#"); hashIndex != -1 {
			comment = strings.TrimSpace(line[hashIndex+1:])
			line = strings.TrimSpace(line[:hashIndex])
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			// Not a valid entry
			entry.Comment = comment
			hf.Lines = append(hf.Lines, entry)
			continue
		}

		ipStr := fields[0]
		if !iputils.IsIPAddress(ipStr) {
			entry.Comment = comment
			hf.Lines = append(hf.Lines, entry)
			continue
		}

		entry.IP = ipStr
		entry.Domains = fields[1:]
		entry.IsEntry = true
		entry.Comment = comment

		hf.Lines = append(hf.Lines, entry)
		entryPtr := &hf.Lines[len(hf.Lines)-1]

		// Fill lookup maps
		hf.ByIP[ipStr] = entryPtr
		for _, domain := range entry.Domains {
			hf.ByDomain[domain] = entryPtr
		}
	}

	return hf, nil
}

// SaveHostsFile writes the string representation of a HostsFile struct to a file.
func SaveHostsFile(filePath string, hf *interfaces.HostsFile) error {
	content := hf.ToHostsFileString()
	return os.WriteFile(filePath, []byte(content), 0644)
}
