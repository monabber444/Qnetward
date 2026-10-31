package collector

import (
	"encoding/binary"
	"errors"
	"net"
	"time"
)

var (
	ErrPacketTooShort = errors.New("flow packet too short")
	ErrUnsupportedVer = errors.New("unsupported flow protocol version")
)

// FlowRecord captures standard IP flow telemetry (NetFlow v5 / IPFIX).
type FlowRecord struct {
	SrcIP       string        `json:"src_ip"`
	DstIP       string        `json:"dst_ip"`
	NextHopIP   string        `json:"next_hop_ip"`
	SrcPort     uint16        `json:"src_port"`
	DstPort     uint16        `json:"dst_port"`
	Protocol    uint8         `json:"protocol"` // 6=TCP, 17=UDP, 1=ICMP
	TCPFlags    uint8         `json:"tcp_flags"`
	Packets     uint32        `json:"packets"`
	Bytes       uint32        `json:"bytes"`
	InputIf     uint16        `json:"input_if"`
	OutputIf    uint16        `json:"output_if"`
	Duration    time.Duration `json:"duration"`
}

// FlowPacket represents an ingested NetFlow / IPFIX / sFlow packet.
type FlowPacket struct {
	Version      uint16       `json:"version"`
	Count        uint16       `json:"count"`
	SysUptimeMs  uint32       `json:"sys_uptime_ms"`
	UnixSecs     uint32       `json:"unix_secs"`
	FlowSequence uint32       `json:"flow_sequence"`
	Records      []FlowRecord `json:"records"`
}

// ParseNetFlowV5 unpacks a standard 24-byte NetFlow v5 header and 48-byte records.
func ParseNetFlowV5(data []byte) (*FlowPacket, error) {
	if len(data) < 24 {
		return nil, ErrPacketTooShort
	}

	version := binary.BigEndian.Uint16(data[0:2])
	if version != 5 {
		return nil, ErrUnsupportedVer
	}

	count := binary.BigEndian.Uint16(data[2:4])
	uptime := binary.BigEndian.Uint32(data[4:8])
	unixSecs := binary.BigEndian.Uint32(data[8:12])
	seq := binary.BigEndian.Uint32(data[16:20])

	packet := &FlowPacket{
		Version:      version,
		Count:        count,
		SysUptimeMs:  uptime,
		UnixSecs:     unixSecs,
		FlowSequence: seq,
		Records:      make([]FlowRecord, 0, count),
	}

	offset := 24
	const recordLen = 48
	for i := 0; i < int(count) && offset+recordLen <= len(data); i++ {
		recData := data[offset : offset+recordLen]
		srcIP := net.IP(recData[0:4]).String()
		dstIP := net.IP(recData[4:8]).String()
		nextHop := net.IP(recData[8:12]).String()
		inIf := binary.BigEndian.Uint16(recData[12:14])
		outIf := binary.BigEndian.Uint16(recData[14:16])
		pkts := binary.BigEndian.Uint32(recData[16:20])
		bytes := binary.BigEndian.Uint32(recData[20:24])
		first := binary.BigEndian.Uint32(recData[24:28])
		last := binary.BigEndian.Uint32(recData[28:32])
		srcPort := binary.BigEndian.Uint16(recData[32:34])
		dstPort := binary.BigEndian.Uint16(recData[34:36])
		tcpFlags := recData[37]
		proto := recData[38]

		var dur time.Duration
		if last >= first {
			dur = time.Duration(last-first) * time.Millisecond
		}

		packet.Records = append(packet.Records, FlowRecord{
			SrcIP:     srcIP,
			DstIP:     dstIP,
			NextHopIP: nextHop,
			SrcPort:   srcPort,
			DstPort:   dstPort,
			Protocol:  proto,
			TCPFlags:  tcpFlags,
			Packets:   pkts,
			Bytes:     bytes,
			InputIf:   inIf,
			OutputIf:  outIf,
			Duration:  dur,
		})
		offset += recordLen
	}

	return packet, nil
}
