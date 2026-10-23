package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const catalogPath = "requirements/catalog.json"
const matrixPath = "docs/plan/04-traceability-matrix.md"

var (
	frPattern       = regexp.MustCompile(`\| ([A-Z]+-[0-9]{2}) \|`)
	workflowPattern = regexp.MustCompile(`^\| (6\.[0-9]+) \|`)
	linkPattern     = regexp.MustCompile(`^(.+)#L([1-9][0-9]*)$`)
)

type record struct {
	ID      string   `json:"id"`
	Kind    string   `json:"kind"`
	Status  string   `json:"status"`
	Source  string   `json:"source"`
	History []change `json:"history,omitempty"`
}

type change struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Evidence string `json:"evidence"`
}

type catalog struct {
	Records []record `json:"records"`
}

var expectedCounts = map[string]int{
	"functional_requirement": 132,
	"detection":              48,
	"acceptance_test":        32,
	"workflow":               14,
	"connector":              28,
}

var expectedFRPrefixes = map[string]int{
	"DISC": 6, "INV": 6, "GRAPH": 6, "OWN": 6, "ENT": 7,
	"TELE": 6, "EXP": 6, "POST": 6, "DET": 7, "CASE": 6,
	"REVIEW": 6, "PROV": 6, "VAULT": 5, "ROT": 7, "LIFE": 6,
	"THIRD": 5, "AGENT": 6, "ACCESS": 6, "DEV": 5, "REPORT": 6,
	"ADMIN": 6, "EXT": 6,
}

var expectedConnectorIDs = []string{
	"AWS IAM / Organizations", "AWS workload services", "Microsoft Entra ID",
	"Azure resources", "GCP IAM / workloads", "Kubernetes", "Active Directory",
	"Windows / Linux / macOS hosts", "GitHub / GHE", "GitLab", "Jenkins, Azure DevOps",
	"Okta", "Google Workspace, M365/SharePoint/Teams, Slack, Salesforce",
	"ServiceNow (identity side + W)", "Snowflake, Databricks",
	"PostgreSQL, MySQL, SQL Server, Oracle, MongoDB, Redis", "HashiCorp Vault, OpenBao",
	"AWS SM/SSM/KMS, Azure Key Vault, GCP Secret Manager",
	"CyberArk / Conjur; Infisical / 1Password / Bitwarden SM",
	"Certificate authorities / cert-manager / ACME", "SSH / SSH CAs",
	"AI platform administration", "MCP / agent frameworks",
	"SPIRE / workload federation / access gateways",
	"HR / directory / CMDB / service catalogs (B)", "SIEM / SOAR / ITSM / collaboration (U/W)",
	"Cartography / Prowler / scanners / BloodHound / PMapper (imports)",
	"Custom application SDK / generic HTTP",
}

var sourceLines = map[string]int{
	"functional_requirement": 28,
	"detection":              29,
	"acceptance_test":        30,
	"workflow":               31,
	"connector":              32,
}

var validStatuses = map[string]bool{
	"specified": true, "designed": true, "implemented": true,
	"tested": true, "certified": true, "unavailable": true,
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("requirements", flag.ContinueOnError)
	flags.SetOutput(stderr)
	seed := flags.Bool("seed", false, "regenerate the initial catalog from the public traceability matrix")
	transition := flags.String("transition", "", "transition one record using ID=STATUS")
	evidence := flags.String("evidence", "", "source link required for a status transition")
	root := flags.String("root", ".", "repository root")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *seed {
		cat, err := seedCatalog(filepath.Join(*root, matrixPath), *root)
		if err != nil {
			return err
		}
		if current, err := readCatalog(filepath.Join(*root, catalogPath)); err == nil {
			statuses := make(map[string]record, len(current.Records))
			for _, item := range current.Records {
				if _, exists := statuses[item.ID]; exists {
					return fmt.Errorf("cannot preserve duplicate ID %q while seeding", item.ID)
				}
				statuses[item.ID] = item
			}
			for index := range cat.Records {
				if previous, ok := statuses[cat.Records[index].ID]; ok {
					cat.Records[index].Status = previous.Status
					cat.Records[index].History = previous.History
				}
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := validate(cat, *root); err != nil {
			return err
		}
		if err := validateAgainstMatrix(cat, *root); err != nil {
			return err
		}
		return writeCatalog(filepath.Join(*root, catalogPath), cat)
	}
	cat, err := readCatalog(filepath.Join(*root, catalogPath))
	if err != nil {
		return err
	}
	if *transition != "" {
		id, status, ok := strings.Cut(*transition, "=")
		if !ok || id == "" || status == "" {
			return errors.New("-transition must be ID=STATUS")
		}
		if *evidence == "" {
			return errors.New("-evidence is required for a status transition")
		}
		if err := applyTransition(&cat, id, status, *evidence, *root); err != nil {
			return err
		}
		if err := validate(cat, *root); err != nil {
			return err
		}
		if err := validateAgainstMatrix(cat, *root); err != nil {
			return err
		}
		return writeCatalog(filepath.Join(*root, catalogPath), cat)
	}
	if err := validate(cat, *root); err != nil {
		return err
	}
	if err := validateAgainstMatrix(cat, *root); err != nil {
		return err
	}
	_, err = io.WriteString(stdout, render(cat))
	return err
}

func seedCatalog(matrix, root string) (catalog, error) {
	file, err := os.Open(matrix)
	if err != nil {
		return catalog{}, err
	}
	defer func() { _ = file.Close() }()
	var cat catalog
	scanner := bufio.NewScanner(file)
	section := ""
	for scanner.Scan() {
		content := scanner.Text()
		switch strings.TrimSpace(content) {
		case "## 1. Functional requirements (132)":
			section = "functional"
		case "## 2. Workflows (spec §6)":
			section = "workflows"
		case "## 3. Detection catalog (spec §8.2)":
			section = "detections"
		case "## 4. Acceptance tests (spec §14.2)":
			section = "acceptance"
		case "## 5. Connector families (spec §7.2)":
			section = "connectors"
		case "## 6. UI surfaces (spec §9.1)":
			section = ""
		}
		switch section {
		case "functional":
			for _, match := range frPattern.FindAllStringSubmatch(content, -1) {
				cat.Records = append(cat.Records, seededRecord("FR-"+match[1], "functional_requirement"))
			}
		case "workflows":
			if match := workflowPattern.FindStringSubmatch(content); match != nil {
				cat.Records = append(cat.Records, seededRecord(match[1], "workflow"))
			}
		case "detections":
			if strings.HasPrefix(content, "| 3 | D01–D06") {
				for id := 1; id <= 48; id++ {
					cat.Records = append(cat.Records, seededRecord(fmt.Sprintf("D%02d", id), "detection"))
				}
			}
		case "acceptance":
			if strings.HasPrefix(content, "| 1 | AT01,") {
				for id := 1; id <= 32; id++ {
					cat.Records = append(cat.Records, seededRecord(fmt.Sprintf("AT%02d", id), "acceptance_test"))
				}
			}
		case "connectors":
			if strings.HasPrefix(content, "| ") && !strings.HasPrefix(content, "| Connector family") && !strings.HasPrefix(content, "|---") {
				cells := strings.Split(content, "|")
				if len(cells) >= 3 {
					cat.Records = append(cat.Records, seededRecord(strings.TrimSpace(cells[1]), "connector"))
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return catalog{}, err
	}
	return cat, nil
}

func validateAgainstMatrix(cat catalog, root string) error {
	if _, err := os.Stat(filepath.Join(root, matrixPath)); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	expected, err := seedCatalog(filepath.Join(root, matrixPath), root)
	if err != nil {
		return fmt.Errorf("read expected requirement IDs: %w", err)
	}
	actualIDs := make(map[string]string, len(cat.Records))
	for _, item := range cat.Records {
		actualIDs[item.ID] = item.Kind
	}
	for _, item := range expected.Records {
		kind, ok := actualIDs[item.ID]
		if !ok {
			return fmt.Errorf("missing requirement ID %q from %s", item.ID, item.Source)
		}
		if kind != item.Kind {
			return fmt.Errorf("requirement %q has kind %q; want %q", item.ID, kind, item.Kind)
		}
	}
	expectedIDs := make(map[string]bool, len(expected.Records))
	for _, item := range expected.Records {
		expectedIDs[item.ID] = true
	}
	for _, item := range cat.Records {
		if !expectedIDs[item.ID] {
			return fmt.Errorf("unknown requirement ID %q", item.ID)
		}
	}
	return nil
}

func seededRecord(id, kind string) record {
	return record{ID: id, Kind: kind, Status: "specified", Source: fmt.Sprintf("requirements/README.md#L%d", sourceLines[kind])}
}

func readCatalog(path string) (catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return catalog{}, err
	}
	var cat catalog
	if err := json.Unmarshal(data, &cat); err != nil {
		return catalog{}, fmt.Errorf("decode %s: %w", path, err)
	}
	return cat, nil
}

func writeCatalog(path string, cat catalog) error {
	data, err := json.MarshalIndent(cat, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func validate(cat catalog, root string) error {
	seen := make(map[string]bool, len(cat.Records))
	counts := make(map[string]int)
	ids := make(map[string]bool, len(cat.Records))
	for _, item := range cat.Records {
		if item.ID == "" || item.Kind == "" {
			return errors.New("every requirement must have an ID and kind")
		}
		if seen[item.ID] {
			return fmt.Errorf("duplicate requirement ID %q", item.ID)
		}
		seen[item.ID] = true
		ids[item.ID] = true
		if _, ok := expectedCounts[item.Kind]; !ok {
			return fmt.Errorf("requirement %q has invalid kind %q", item.ID, item.Kind)
		}
		counts[item.Kind]++
		if !validStatuses[item.Status] {
			return fmt.Errorf("requirement %q has invalid status %q", item.ID, item.Status)
		}
		if err := validateLink(root, item.Source); err != nil {
			return fmt.Errorf("requirement %q source: %w", item.ID, err)
		}
		if item.Source != seededRecord(item.ID, item.Kind).Source {
			return fmt.Errorf("requirement %q source does not match its category source register", item.ID)
		}
		current := "specified"
		for _, step := range item.History {
			if step.From != current || !allowedTransition(step.From, step.To) {
				return fmt.Errorf("requirement %q has invalid status transition %q -> %q", item.ID, step.From, step.To)
			}
			if err := validateLink(root, step.Evidence); err != nil {
				return fmt.Errorf("requirement %q evidence: %w", item.ID, err)
			}
			current = step.To
		}
		if item.Status != current {
			return fmt.Errorf("requirement %q status %q does not match evidence history ending at %q", item.ID, item.Status, current)
		}
	}
	kinds := make([]string, 0, len(expectedCounts))
	for kind := range expectedCounts {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		expected := expectedCounts[kind]
		if counts[kind] != expected {
			return fmt.Errorf("kind %q has %d records; want %d", kind, counts[kind], expected)
		}
	}
	for prefix, count := range expectedFRPrefixes {
		for id := 1; id <= count; id++ {
			requirementID := fmt.Sprintf("FR-%s-%02d", prefix, id)
			if !ids[requirementID] {
				return fmt.Errorf("missing requirement ID %q", requirementID)
			}
		}
	}
	for id := 1; id <= 48; id++ {
		if requirementID := fmt.Sprintf("D%02d", id); !ids[requirementID] {
			return fmt.Errorf("missing requirement ID %q", requirementID)
		}
	}
	for id := 1; id <= 32; id++ {
		if requirementID := fmt.Sprintf("AT%02d", id); !ids[requirementID] {
			return fmt.Errorf("missing requirement ID %q", requirementID)
		}
	}
	for id := 1; id <= 14; id++ {
		if requirementID := fmt.Sprintf("6.%d", id); !ids[requirementID] {
			return fmt.Errorf("missing requirement ID %q", requirementID)
		}
	}
	for _, requirementID := range expectedConnectorIDs {
		if !ids[requirementID] {
			return fmt.Errorf("missing connector-family row %q", requirementID)
		}
	}
	return nil
}

func validateLink(root, link string) error {
	match := linkPattern.FindStringSubmatch(link)
	if match == nil {
		return fmt.Errorf("%q must use path#Lline format", link)
	}
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	path := filepath.Join(cleanRoot, filepath.Clean(match[1]))
	if !strings.HasPrefix(path, cleanRoot+string(filepath.Separator)) {
		return fmt.Errorf("%q escapes repository root", link)
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open %q: %w", match[1], err)
	}
	defer func() { _ = file.Close() }()
	wanted := 0
	if _, err := fmt.Sscanf(match[2], "%d", &wanted); err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	for current := 1; scanner.Scan(); current++ {
		if current == wanted && strings.TrimSpace(scanner.Text()) != "" {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return fmt.Errorf("%q points to a missing or blank line", link)
}

func applyTransition(cat *catalog, id, next, evidence, root string) error {
	if !validStatuses[next] {
		return fmt.Errorf("invalid status %q", next)
	}
	if err := validateLink(root, evidence); err != nil {
		return fmt.Errorf("evidence: %w", err)
	}
	for index := range cat.Records {
		item := &cat.Records[index]
		if item.ID != id {
			continue
		}
		if !allowedTransition(item.Status, next) {
			return fmt.Errorf("transition %s -> %s is not allowed for %s", item.Status, next, id)
		}
		item.History = append(item.History, change{From: item.Status, To: next, Evidence: evidence})
		item.Status = next
		return nil
	}
	return fmt.Errorf("unknown requirement ID %q", id)
}

func allowedTransition(current, next string) bool {
	if next == "unavailable" {
		return current != "unavailable" && current != "certified"
	}
	successor := map[string]string{
		"specified": "designed", "designed": "implemented",
		"implemented": "tested", "tested": "certified",
	}
	return successor[current] == next
}

func render(cat catalog) string {
	counts := make(map[string]map[string]int)
	for _, item := range cat.Records {
		if counts[item.Kind] == nil {
			counts[item.Kind] = make(map[string]int)
		}
		counts[item.Kind][item.Status]++
	}
	kinds := make([]string, 0, len(expectedCounts))
	for kind := range expectedCounts {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	var output strings.Builder
	output.WriteString("# Requirement status\n\n")
	output.WriteString("Generated from `requirements/catalog.json`. Statuses describe specification and evidence maturity; they do not imply that planned product behavior is available.\n\n")
	output.WriteString("| Category | Total | Specified | Designed | Implemented | Tested | Certified | Unavailable |\n|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, kind := range kinds {
		row := counts[kind]
		fmt.Fprintf(&output, "| %s | %d | %d | %d | %d | %d | %d | %d |\n", kind, expectedCounts[kind], row["specified"], row["designed"], row["implemented"], row["tested"], row["certified"], row["unavailable"])
	}
	return output.String()
}
