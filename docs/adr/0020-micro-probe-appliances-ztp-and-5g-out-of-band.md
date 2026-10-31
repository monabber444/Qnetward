# 20. Micro-Probe Appliances, Zero-Touch Provisioning, and 5G Out-of-Band

## Status

Accepted

## Context

Enterprise networks often extend into edge branch offices, remote retail stores, manufacturing plants, and distribution centers without dedicated onsite network engineers:
1. **Deployment Friction**: Shipping complex x86 servers requiring manual Linux configuration, SSH key installation, and static IP setup is slow, error-prone, and expensive.
2. **WAN Blindness During Outages**: When a branch office's primary fiber or cable WAN connection is severed, conventional network monitoring loses connectivity to the remote agent, leaving engineers blind to whether the failure is a local hardware freeze, an unplugged Ethernet cord, or an upstream ISP outage.
3. **Emergency Out-of-Band Recovery**: Operators need out-of-band management (OOBM) with cellular failover and remote serial console access to recover frozen branch routers without costly "truck rolls" (dispatching physical technicians).

## Decision

1. **Micro-Probe Appliance Engine & ZTP (`pkg/appliance`)**:
   - **Supported Hardware Profiles**: Minimal embedded Alpine/Debian Linux firmware for Raspberry Pi 4/5 (Compute Module 4), Intel NUC Edge, Odroid M1, and fanless x86 mini-PCs.
   - **Zero-Touch Provisioning (ZTP)**:
     - Administrators generate a one-time cryptographic enrollment token with configurable expiration.
     - Upon boot, the appliance obtains an IP via DHCP, contacts the Qnetward Ingestion Gateway, verifies mTLS cryptographic identity, and registers its serial number and location automatically.
   - **`ApplianceManager` (`pkg/appliance/manager.go`)**: Coordinates enrollment tokens, appliance health state, and uplink interfaces.

2. **Cellular LTE/5G Out-of-Band Management (OOBM)**:
   - **Automated Uplink Failover Engine**:
     - Constantly monitors primary WAN health (`eth0`) and cellular 5G modem state (`wwan0`, RSRP signal quality).
     - Upon primary WAN packet loss, seamlessly shifts telemetry transmission to 5G cellular (`UplinkCellular5G`), marking status as `FAILOVER`.
     - Automatically restores primary WAN when link returns to baseline (`RECOVERED`).
   - Dispatches structured events to `eventbus.Bus` over topics `appliance.failover` and `appliance.recovered`.

3. **Web Cockpit Appliance Workspace (`/appliances`)**:
   - **Fleet Overview**: Live status cards tracking appliance hardware model, primary WAN link, 5G signal strength (dBm), and active uplink.
   - **ZTP Provisioning Hub**: One-click token generator and enrollment workflow instructions.
   - **Emergency 5G Serial Console Simulator**: Interactive terminal providing remote diagnostic command execution and root-cause discrimination (e.g. local Ethernet disconnected vs ISP carrier cut).

## Consequences

- Cuts branch office rollout time from weeks to minutes via plug-and-play ZTP.
- Eliminates monitoring blind spots during branch outages through continuous 5G cellular failover.
- Prevents expensive technician dispatches through remote serial console diagnostics over cellular OOBM.
