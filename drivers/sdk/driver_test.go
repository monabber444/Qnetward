package sdk_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"netwatcher/drivers/sdk"
)

func TestProfileLoaderAndDrivers(t *testing.T) {
	registry := sdk.NewProfileRegistry()

	// Locate profiles directory
	profilesDir := filepath.Join("..", "profiles")
	if _, err := os.Stat(profilesDir); err != nil {
		t.Fatalf("profiles directory not found: %v", err)
	}

	err := registry.LoadFS(os.DirFS(".."), "profiles")
	if err != nil {
		t.Fatalf("failed to load profiles from FS: %v", err)
	}

	profiles := registry.List()
	if len(profiles) < 5 {
		t.Fatalf("expected at least 5 profiles, loaded %d", len(profiles))
	}

	// 1. Verify Cisco profile
	ciscoProf, ok := registry.Get("cisco_iosxe")
	if !ok {
		t.Fatalf("cisco_iosxe profile not found")
	}
	if ciscoProf.Vendor != "Cisco Systems" {
		t.Fatalf("expected vendor Cisco Systems, got %s", ciscoProf.Vendor)
	}

	// 2. Verify SysObjectID matching
	matched, ok := registry.FindForSysObjectID("1.3.6.1.4.1.9.1.2500")
	if !ok || matched.Name != "cisco_iosxe" {
		t.Fatalf("failed to match Cisco SysObjectID: %+v", matched)
	}

	matchedMikrotik, ok := registry.FindForSysObjectID("1.3.6.1.4.1.14988.1.5")
	if !ok || matchedMikrotik.Name != "mikrotik_routeros" {
		t.Fatalf("failed to match MikroTik SysObjectID: %+v", matchedMikrotik)
	}

	// 3. Test Cisco Driver Execution
	ciscoDriver := sdk.NewProfileDriver(ciscoProf)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	target := sdk.TargetConfig{
		Host:      "10.0.0.1",
		Port:      161,
		Protocol:  "snmp_v2c",
		Community: "public",
	}

	info, err := ciscoDriver.Discover(ctx, target)
	if err != nil {
		t.Fatalf("discover failed: %v", err)
	}
	if info.Vendor != "Cisco Systems" {
		t.Fatalf("unexpected vendor: %s", info.Vendor)
	}

	// Collect Metrics & Optical DOM
	metrics, err := ciscoDriver.CollectMetrics(ctx, target)
	if err != nil {
		t.Fatalf("collect metrics failed: %v", err)
	}
	if metrics.CPUUsagePercent <= 0 {
		t.Fatalf("expected positive CPU usage")
	}
	if len(metrics.Interfaces) < 2 {
		t.Fatalf("expected at least 2 interfaces, got %d", len(metrics.Interfaces))
	}
	if len(metrics.OpticalDOM) == 0 {
		t.Fatalf("expected Optical DOM metrics for Cisco")
	}
	if metrics.OpticalDOM[0].TxPowerDbm >= 0 {
		t.Fatalf("expected negative TxPowerDbm, got %f", metrics.OpticalDOM[0].TxPowerDbm)
	}

	// Fetch Topology (LLDP)
	neighbors, err := ciscoDriver.FetchTopology(ctx, target)
	if err != nil {
		t.Fatalf("fetch topology failed: %v", err)
	}
	if len(neighbors) == 0 {
		t.Fatalf("expected LLDP neighbors")
	}
	if neighbors[0].RemoteSystemName != "dist-switch-01.internal" {
		t.Fatalf("unexpected remote system name: %s", neighbors[0].RemoteSystemName)
	}

	// Parse Trap
	trap, err := ciscoDriver.ParseTrap([]byte("BGP neighbor 10.0.1.2 Down: peer hold timer expired"))
	if err != nil {
		t.Fatalf("parse trap failed: %v", err)
	}
	if trap.Severity != "critical" {
		t.Fatalf("expected critical trap severity, got %s", trap.Severity)
	}

	// 4. Test APC UPS Driver Execution (Power Status)
	apcProf, ok := registry.Get("apc_ups")
	if !ok {
		t.Fatalf("apc_ups profile not found")
	}
	apcDriver := sdk.NewProfileDriver(apcProf)
	apcMetrics, err := apcDriver.CollectMetrics(ctx, target)
	if err != nil {
		t.Fatalf("collect APC metrics failed: %v", err)
	}
	if apcMetrics.Power == nil {
		t.Fatalf("expected power telemetry for APC UPS")
	}
	if apcMetrics.Power.BatteryChargePercent != 100.0 {
		t.Fatalf("expected 100%% battery charge, got %f", apcMetrics.Power.BatteryChargePercent)
	}

	// 5. Test Dell iDRAC Redfish Driver Execution
	idracProf, ok := registry.Get("dell_idrac_redfish")
	if !ok {
		t.Fatalf("dell_idrac_redfish profile not found")
	}
	idracDriver := sdk.NewProfileDriver(idracProf)
	redfishTarget := sdk.TargetConfig{
		Host:        "192.168.1.100",
		Port:        443,
		Protocol:    "redfish",
		Username:    "root",
		Password:    "calvin",
		InsecureTLS: true,
	}
	redfishInfo, err := idracDriver.Discover(ctx, redfishTarget)
	if err != nil {
		t.Fatalf("discover Redfish failed: %v", err)
	}
	if redfishInfo.Vendor != "Dell Technologies" {
		t.Fatalf("expected Dell Technologies, got %s", redfishInfo.Vendor)
	}
}
