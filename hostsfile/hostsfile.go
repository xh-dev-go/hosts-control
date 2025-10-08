package interfaces

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xh-dev-go/hosts-control/iputils"
)

// Represents one line in /etc/hosts
type HostLine struct {
	Raw     string   `json:"raw"`     // Original line (unmodified)
	IsEntry bool     `json:"isEntry"` // Whether this is an IP entry line
	IP      string   `json:"ip"`      // IP address, if applicable
	Domains []string `json:"domains"` // Hostnames or aliases
	Comment string   `json:"comment"` // Inline or full-line comment
	IsBlank bool     `json:"isBlank"` // True if the line is empty
}

// Represents the entire /etc/hosts file
type HostsFile struct {
	Lines    []HostLine           `json:"lines"`    // Ordered lines (preserve file structure)
	ByIP     map[string]*HostLine `json:"byIp"`     // IP → entry
	ByDomain map[string]*HostLine `json:"byDomain"` // Domain → entry
}

func (hf *HostsFile) ToJSON(pretty bool) (string, error) {
	var (
		data []byte
		err  error
	)
	if pretty {
		data, err = json.MarshalIndent(hf, "", "  ")
	} else {
		data, err = json.Marshal(hf)
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToHostsFileString reconstructs the /etc/hosts file content from the HostsFile struct.
func (hf *HostsFile) ToHostsFileString() string {
	var sb strings.Builder

	for i, line := range hf.Lines {
		switch {
		case line.IsBlank:
			sb.WriteString("\n")

		case !line.IsEntry:
			// Comment or malformed line
			if strings.TrimSpace(line.Comment) != "" {
				sb.WriteString("# " + strings.TrimSpace(line.Comment) + "\n")
			} else if line.Raw != "" {
				sb.WriteString(line.Raw + "\n")
			} else {
				sb.WriteString("\n")
			}

		default:
			// Regular entry
			sb.WriteString(line.IP)
			if len(line.Domains) > 0 {
				sb.WriteString(" " + strings.Join(line.Domains, " "))
			}
			if strings.TrimSpace(line.Comment) != "" {
				sb.WriteString(" # " + strings.TrimSpace(line.Comment))
			}
			sb.WriteString("\n")
		}

		// Ensure no extra newline at the end
		if i == len(hf.Lines)-1 {
			break
		}
	}

	return sb.String()
}

// --- Helper to rebuild lookup maps from hf.Lines ---
func (hf *HostsFile) rebuildIndex() {
	hf.ByIP = make(map[string]*HostLine)
	hf.ByDomain = make(map[string]*HostLine)

	for i := range hf.Lines {
		l := &hf.Lines[i]
		if l.IsEntry {
			hf.ByIP[l.IP] = l
			for _, d := range l.Domains {
				hf.ByDomain[strings.TrimSpace(d)] = l
			}
		}
	}
}

// --- RemoveDomain only modifies hf.Lines ---
func (hf *HostsFile) RemoveDomain(domain string) bool {
	domain = strings.TrimSpace(domain)
	var changed bool

	for i := 0; i < len(hf.Lines); i++ {
		l := &hf.Lines[i]
		if !l.IsEntry {
			continue
		}

		newDomains := make([]string, 0, len(l.Domains))
		for _, d := range l.Domains {
			if strings.TrimSpace(d) != domain {
				newDomains = append(newDomains, d)
			} else {
				changed = true
			}
		}
		l.Domains = newDomains

		// Remove line if no domains left
		if len(l.Domains) == 0 && l.IP != "" {
			hf.Lines = append(hf.Lines[:i], hf.Lines[i+1:]...)
			i-- // adjust index after deletion
		}
	}

	if changed {
		hf.rebuildIndex()
	}
	return changed
}

// --- AddDomain only modifies hf.Lines ---
func (hf *HostsFile) AddDomain(ip, domain, comment string) (bool, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return false, fmt.Errorf("domain cannot be empty")
	}
	if !iputils.IsIPAddress(ip) {
		return false, fmt.Errorf("invalid IP address: %s", ip)
	}

	var changed bool
	var foundEntry *HostLine

	// 1. Check if domain exists
	for i := range hf.Lines {
		l := &hf.Lines[i]
		if !l.IsEntry {
			continue
		}

		for _, d := range l.Domains {
			if strings.TrimSpace(d) == domain {
				foundEntry = l
				break
			}
		}
		if foundEntry != nil {
			break
		}
	}

	// Case 1: Domain exists under same IP
	if foundEntry != nil && foundEntry.IP == ip {
		if comment != "" && foundEntry.Comment != comment {
			foundEntry.Comment = comment
			changed = true
		}
		if !changed {
			return false, nil
		}
		hf.rebuildIndex()
		return true, nil
	}

	// Case 2 & 3: Domain exists under different IP → remove first
	if foundEntry != nil && foundEntry.IP != ip {
		hf.RemoveDomain(domain)
		changed = true
	}

	// Case 4: IP exists → append domain
	var ipEntry *HostLine
	for i := range hf.Lines {
		l := &hf.Lines[i]
		if l.IsEntry && l.IP == ip {
			ipEntry = l
			break
		}
	}

	if ipEntry != nil {
		ipEntry.Domains = append(ipEntry.Domains, domain)
		if comment != "" && ipEntry.Comment != comment {
			ipEntry.Comment = comment
		}
		changed = true
		hf.rebuildIndex()
		return changed, nil
	}

	// Case 5: IP does not exist → create new entry
	newEntry := HostLine{
		Raw:     ip + " " + domain,
		IP:      ip,
		Domains: []string{domain},
		IsEntry: true,
		Comment: comment,
	}
	hf.Lines = append(hf.Lines, newEntry)
	changed = true
	hf.rebuildIndex()
	return changed, nil
}

// --- Public wrapper without comment ---
func (hf *HostsFile) AddDomainSimple(ip, domain string) (bool, error) {
	return hf.AddDomain(ip, domain, "")
}

// --- Public wrapper with comment ---
func (hf *HostsFile) AddDomainWithComment(ip, domain, comment string) (bool, error) {
	return hf.AddDomain(ip, domain, comment)
}

// MergeByIP merges host file lines that share the same IP address.
// The first occurrence of an IP keeps its line, and subsequent lines
// with the same IP have their domains and comments (if the first is empty)
// appended to the first line.
func (hf *HostsFile) MergeByIP() bool {
	newLines := make([]HostLine, 0, len(hf.Lines))
	ipToLine := make(map[string]*HostLine)
	var changed bool

	for _, line := range hf.Lines {
		if !line.IsEntry || line.IP == "" {
			newLines = append(newLines, line)
			continue
		}

		if existingLine, found := ipToLine[line.IP]; found {
			existingLine.Domains = append(existingLine.Domains, line.Domains...)
			if existingLine.Comment == "" && line.Comment != "" {
				existingLine.Comment = line.Comment
			}
			changed = true
		} else {
			newLine := line // Make a copy to avoid pointer issues
			newLines = append(newLines, newLine)
			ipToLine[line.IP] = &newLines[len(newLines)-1]
		}
	}

	if changed {
		hf.Lines = newLines
		hf.rebuildIndex()
	}
	return changed
}
