package nats

import (
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func TestJetStreamPublishAckAndRead(t *testing.T) {
	options := &server.Options{
		Host:      "127.0.0.1",
		Port:      -1,
		JetStream: true,
		StoreDir:  t.TempDir(),
		NoLog:     true,
		NoSigs:    true,
	}
	natsServer, err := server.NewServer(options)
	if err != nil {
		t.Fatal(err)
	}
	go natsServer.Start()
	if !natsServer.ReadyForConnections(10 * time.Second) {
		natsServer.Shutdown()
		t.Fatalf("embedded NATS server did not become ready")
	}
	t.Cleanup(func() {
		natsServer.Shutdown()
		natsServer.WaitForShutdown()
	})
	connection, err := nats.Connect("nats://"+natsServer.Addr().String(), nats.Timeout(3*time.Second))
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
