package nats

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestJetStreamPublishAckAndRead(t *testing.T) {
	serverURL := os.Getenv("NOMYR_SPIKE_NATS_URL")
	if serverURL == "" {
		t.Skip("NOMYR_SPIKE_NATS_URL is not set; NATS behavior not exercised")
	}
	connection, err := nats.Connect(serverURL, nats.Timeout(3*time.Second))
	if err != nil {
		t.Fatalf("connect to explicit local NATS service: %v", err)
	}
	defer connection.Close()
	jetstream, err := connection.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	streamName := fmt.Sprintf("NOMYR_SPIKE_%d", time.Now().UnixNano())
	subject := "nomyr.spike." + streamName
	if _, err := jetstream.AddStream(&nats.StreamConfig{Name: streamName, Subjects: []string{subject}, Storage: nats.FileStorage}); err != nil {
		t.Fatal(err)
	}
	defer jetstream.DeleteStream(streamName)
	ack, err := jetstream.Publish(subject, []byte("synthetic-event"))
	if err != nil {
		t.Fatal(err)
	}
	message, err := jetstream.GetMsg(streamName, ack.Sequence)
	if err != nil {
		t.Fatal(err)
	}
	if string(message.Data) != "synthetic-event" {
		t.Fatalf("read payload %q, want synthetic event", message.Data)
	}
}
