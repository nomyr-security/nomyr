package securityharness

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type SourceKind string

const (
	SourceFile         SourceKind = "file"
	SourceArtifact     SourceKind = "artifact"
	SourceLog          SourceKind = "log"
	SourceTrace        SourceKind = "trace"
	SourceDatabaseDump SourceKind = "database_dump"
	canaryReadChunk               = 32 * 1024
)

type Source struct {
	Kind SourceKind
	Open func() (io.ReadCloser, error)
}

type CanaryFinding struct {
	Kind  SourceKind `json:"kind"`
	Code  string     `json:"code"`
	Count int        `json:"count"`
}

func NewCanary() ([]byte, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("generate runtime canary")
	}
	return []byte(hex.EncodeToString(secret)), nil
}

func ScanSources(canary []byte, sources []Source) ([]CanaryFinding, error) {
	if len(canary) == 0 {
		return nil, fmt.Errorf("canary must not be empty")
	}
	var findings []CanaryFinding
	for _, source := range sources {
		if !validSourceKind(source.Kind) {
			return nil, fmt.Errorf("unsupported source kind")
		}
		if source.Open == nil {
			return nil, fmt.Errorf("source cannot be opened")
		}
		reader, err := source.Open()
		if err != nil {
			return nil, fmt.Errorf("source cannot be opened")
		}
		count, readErr := countCanary(reader, canary)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			return nil, fmt.Errorf("source cannot be scanned")
		}
		if count > 0 {
			findings = append(findings, CanaryFinding{Kind: source.Kind, Code: "CANARY_LEAK", Count: count})
		}
	}
	return findings, nil
}

func validSourceKind(kind SourceKind) bool {
	switch kind {
	case SourceFile, SourceArtifact, SourceLog, SourceTrace, SourceDatabaseDump:
		return true
	default:
		return false
	}
}

func countCanary(reader io.Reader, canary []byte) (int, error) {
	buffer := make([]byte, canaryReadChunk+len(canary)-1)
	carry := 0
	count := 0
	for {
		read, err := reader.Read(buffer[carry:])
		if read > 0 {
			length := carry + read
			count += bytes.Count(buffer[:length], canary)
			carry = min(len(canary)-1, length)
			copy(buffer[:carry], buffer[length-carry:length])
		}
		if err == io.EOF {
			return count, nil
		}
		if err != nil {
			return 0, err
		}
		if read == 0 {
			return 0, io.ErrNoProgress
		}
	}
}

func CheckSources(canary []byte, sources []Source) error {
	findings, err := ScanSources(canary, sources)
	if err != nil {
		return err
	}
	if len(findings) > 0 {
		return fmt.Errorf("canary leak detected")
	}
	return nil
}
