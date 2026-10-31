package cloudconnect

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"netwatcher/pkg/eventbus"
)

func TestAWSConnector_ParseFlowLogs(t *testing.T) {
	conn := NewAWSConnector(CloudAccountConfig{
		AccountID: "123456789012",
		Regions:   []string{"us-east-1"},
	})

	sampleRaw := []byte(`version account-id interface-id srcaddr dstaddr srcport dstport protocol packets bytes start end action log-status
2 123456789012 eni-1a2b3c4d 10.0.1.50 172.217.16.206 49152 443 6 15 1250 1730376000 1730376060 ACCEPT OK
2 123456789012 eni-1a2b3c4d 198.51.100.4 10.0.1.50 54321 22 6 3 180 1730376000 1730376060 REJECT OK`)

	records, err := conn.ParseFlowLogs(sampleRaw)
	if err != nil {
		t.Fatalf("unexpected error parsing AWS flow logs: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	r1 := records[0]
	if r1.Provider != ProviderAWS || r1.SrcIP != "10.0.1.50" || r1.DstPort != 443 || r1.Action != ActionAccept {
		t.Errorf("unexpected record 1: %+v", r1)
	}

	r2 := records[1]
	if r2.Action != ActionReject || r2.DstPort != 22 {
		t.Errorf("unexpected record 2: %+v", r2)
	}
}

func TestGCPConnector_ParseFlowLogs(t *testing.T) {
	conn := NewGCPConnector(CloudAccountConfig{
		AccountID: "my-gcp-project",
		Regions:   []string{"us-central1"},
	})

	sampleJSON := []byte(`{
		"timestamp": "2026-10-31T12:00:00Z",
		"jsonPayload": {
			"connection": {
				"src_ip": "10.128.0.2",
				"dest_ip": "10.128.0.3",
				"src_port": 45120,
				"dest_port": 80,
				"protocol": 6
			},
			"bytes_sent": "2412",
			"packets_sent": "18",
			"src_instance": {
				"project_id": "prod-infra",
				"zone": "us-central1-a"
			}
		}
	}`)

	records, err := conn.ParseFlowLogs(sampleJSON)
	if err != nil {
		t.Fatalf("unexpected error parsing GCP flow logs: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	r := records[0]
	if r.Provider != ProviderGCP || r.SrcIP != "10.128.0.2" || r.DstIP != "10.128.0.3" || r.Bytes != 2412 {
		t.Errorf("unexpected GCP record: %+v", r)
	}
}

func TestAzureConnector_ParseFlowLogs(t *testing.T) {
	conn := NewAzureConnector(CloudAccountConfig{
		AccountID: "sub-12345",
		Regions:   []string{"eastus"},
	})

	sampleJSON := []byte(`{
		"records": [
			{
				"time": "2026-10-31T12:00:00.000Z",
				"resourceId": "/subscriptions/sub-12345/resourceGroups/rg/providers/Microsoft.Network/networkSecurityGroups/nsg",
				"properties": {
					"Version": 2,
					"flows": [
						{
							"rule": "AllowHTTPS",
							"flows": [
								{
									"mac": "000D3A123456",
									"flowTuples": [
										"1730376000,10.0.0.4,142.250.190.46,49152,443,T,O,A,B,10,4096"
									]
								}
							]
						}
					]
				}
			}
		]
	}`)

	records, err := conn.ParseFlowLogs(sampleJSON)
	if err != nil {
		t.Fatalf("unexpected error parsing Azure flow logs: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	r := records[0]
	if r.Provider != ProviderAzure || r.DstPort != 443 || r.Action != ActionAccept || r.Bytes != 4096 {
		t.Errorf("unexpected Azure record: %+v", r)
	}
}

func TestCloudManager_IngestAndEventBus(t *testing.T) {
	bus := eventbus.NewMemoryBus()
	defer bus.Close()
	mgr := NewCloudManager(bus)

	var mu sync.Mutex
	var receivedRecords []VPCFlowRecord
	var wg sync.WaitGroup
	wg.Add(2)

	sub, err := bus.Subscribe(context.Background(), "telemetry.cloud.aws.flow", func(ctx context.Context, msg eventbus.Message) error {
		defer wg.Done()
		var rec VPCFlowRecord
		if err := json.Unmarshal(msg.Data, &rec); err == nil {
			mu.Lock()
			receivedRecords = append(receivedRecords, rec)
			mu.Unlock()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	records := []VPCFlowRecord{
		{
			Provider:  ProviderAWS,
			AccountID: "12345",
			SrcIP:     "10.0.0.1",
			DstIP:     "10.0.0.2",
			Packets:   5,
			Bytes:     500,
			Action:    ActionAccept,
		},
		{
			Provider:  ProviderAWS,
			AccountID: "12345",
			SrcIP:     "10.0.0.3",
			DstIP:     "10.0.0.4",
			Packets:   2,
			Bytes:     200,
			Action:    ActionReject,
		},
	}

	mgr.IngestFlowRecords(records)

	// Wait for eventbus notification
	wg.Wait()
	mu.Lock()
	count := len(receivedRecords)
	mu.Unlock()

	if count != 2 {
		t.Fatalf("expected 2 received flow records via eventbus, got %d", count)
	}

	acc, rej, pkts := mgr.Stats()
	if acc != 500 || rej != 200 || pkts != 7 {
		t.Errorf("unexpected stats: acc=%d rej=%d pkts=%d", acc, rej, pkts)
	}

	// Test PollAllMetrics
	mgr.RegisterConnector("aws-prod", NewAWSConnector(CloudAccountConfig{}))
	metrics, err := mgr.PollAllMetrics(context.Background())
	if err != nil {
		t.Fatalf("failed to poll all metrics: %v", err)
	}
	if len(metrics) == 0 {
		t.Errorf("expected cloud metrics, got 0")
	}
}
