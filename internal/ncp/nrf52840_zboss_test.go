package ncp

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestCRC8KnownValues(t *testing.T) {
	// Verify CRC8 produces consistent results.
	data := []byte{0x03, 0x00, 0x00, 0xC0}
	crc := zbossCRC8(data)
	// Recompute to verify determinism.
	if crc != zbossCRC8(data) {
		t.Fatal("CRC8 not deterministic")
	}
	// Zero-length input.
	if zbossCRC8(nil) != 0x00 {
		// init=0xFF, no data, xorout=0xFF → 0xFF^0xFF=0x00
		t.Errorf("CRC8(nil) = 0x%02X, want 0x00", zbossCRC8(nil))
	}
}

func TestCRC16Deterministic(t *testing.T) {
	data := []byte{0x00, 0x00, 0x01, 0x00, 0x42}
	a := zbossCRC16(data)
	b := zbossCRC16(data)
	if a != b {
		t.Fatalf("CRC16 not deterministic: 0x%04X vs 0x%04X", a, b)
	}
}

func TestEncodeDecodeRequestRoundTrip(t *testing.T) {
	payload := []byte{0xAA, 0xBB, 0xCC}
	callID := uint16(0x0301)
	tsn := uint8(42)
	pktSeq := uint8(1)

	encoded := zbossEncodeRequest(callID, tsn, pktSeq, payload)

	// Verify signature.
	if encoded[0] != zbossSig0 || encoded[1] != zbossSig1 {
		t.Fatalf("bad signature: 0x%02X%02X", encoded[0], encoded[1])
	}

	decoded, err := zbossDecodeFrame(encoded)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if decoded.HL.PacketType != zbossHLRequest {
		t.Errorf("PacketType: got %d, want %d", decoded.HL.PacketType, zbossHLRequest)
	}
	if decoded.HL.CallID != callID {
		t.Errorf("CallID: got 0x%04X, want 0x%04X", decoded.HL.CallID, callID)
	}
	if decoded.HL.TSN != tsn {
		t.Errorf("TSN: got %d, want %d", decoded.HL.TSN, tsn)
	}
	if !bytes.Equal(decoded.Payload, payload) {
		t.Errorf("Payload: got %X, want %X", decoded.Payload, payload)
	}
	if zbossLLPktSeq(decoded.LL.Flags) != pktSeq {
		t.Errorf("PktSeq: got %d, want %d", zbossLLPktSeq(decoded.LL.Flags), pktSeq)
	}
}

func TestEncodeDecodeACKRoundTrip(t *testing.T) {
	for seq := uint8(0); seq < 4; seq++ {
		encoded := zbossEncodeACK(seq)
		decoded, err := zbossDecodeFrame(encoded)
		if err != nil {
			t.Fatalf("seq=%d decode error: %v", seq, err)
		}
		if !zbossLLIsACK(decoded.LL.Flags) {
			t.Errorf("seq=%d: not an ACK frame", seq)
		}
		if got := zbossLLAckSeq(decoded.LL.Flags); got != seq {
			t.Errorf("seq=%d: AckSeq got %d", seq, got)
		}
	}
}

func TestDecodeFrameTooShort(t *testing.T) {
	_, err := zbossDecodeFrame([]byte{0xDE, 0xAD})
	if err == nil {
		t.Error("expected error for short frame")
	}
}

func TestDecodeFrameBadSignature(t *testing.T) {
	data := make([]byte, 10)
	data[0] = 0xFF
	data[1] = 0xFF
	_, err := zbossDecodeFrame(data)
	if err == nil {
		t.Error("expected error for bad signature")
	}
}

func TestDecodeFrameBadCRC8(t *testing.T) {
	encoded := zbossEncodeACK(0)
	encoded[6] ^= 0xFF // corrupt CRC8
	_, err := zbossDecodeFrame(encoded)
	if err == nil {
		t.Error("expected CRC8 error")
	}
}

func TestDecodeFrameBadCRC16(t *testing.T) {
	encoded := zbossEncodeRequest(0x0001, 1, 0, nil)
	// Corrupt body CRC16 (at offset 7).
	encoded[7] ^= 0xFF
	_, err := zbossDecodeFrame(encoded)
	if err == nil {
		t.Error("expected CRC16 error")
	}
}

func TestLLFlagHelpers(t *testing.T) {
	flags := uint8(zbossFlagFirstFrag | zbossFlagLastFrag | (2 << zbossFlagPktSeqShift))
	if zbossLLPktSeq(flags) != 2 {
		t.Errorf("PktSeq: got %d, want 2", zbossLLPktSeq(flags))
	}
	if zbossLLIsACK(flags) {
		t.Error("should not be ACK")
	}

	ackFlags := uint8(zbossFlagACK | (3 << zbossFlagAckSeqShift))
	if !zbossLLIsACK(ackFlags) {
		t.Error("should be ACK")
	}
	if zbossLLAckSeq(ackFlags) != 3 {
		t.Errorf("AckSeq: got %d, want 3", zbossLLAckSeq(ackFlags))
	}
}

func TestEncodeEmptyPayloadRequest(t *testing.T) {
	encoded := zbossEncodeRequest(zbossCmdNCPReset, 0, 0, nil)
	decoded, err := zbossDecodeFrame(encoded)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if decoded.HL.CallID != zbossCmdNCPReset {
		t.Errorf("CallID: got 0x%04X, want 0x%04X", decoded.HL.CallID, zbossCmdNCPReset)
	}
	if len(decoded.Payload) != 0 {
		t.Errorf("Payload should be empty, got %X", decoded.Payload)
	}
}

func TestBuildAPSDEDataReq(t *testing.T) {
	zclData := []byte{0x10, 0x01, 0x00, 0x00, 0x00}
	buf := buildAPSDEDataReq(0x1234, 1, 1, 0x0006, zclProfileHA, 30, zclData)

	if len(buf) != 24+len(zclData) {
		t.Fatalf("length: got %d, want %d", len(buf), 24+len(zclData))
	}

	dataLen := binary.LittleEndian.Uint16(buf[1:3])
	if dataLen != uint16(len(zclData)) {
		t.Errorf("data_len: got %d, want %d", dataLen, len(zclData))
	}

	dstAddr := binary.LittleEndian.Uint16(buf[3:5])
	if dstAddr != 0x1234 {
		t.Errorf("dst_addr: got 0x%04X, want 0x1234", dstAddr)
	}

	clusterID := binary.LittleEndian.Uint16(buf[13:15])
	if clusterID != 0x0006 {
		t.Errorf("cluster_id: got 0x%04X, want 0x0006", clusterID)
	}

	if buf[18] != zbossAddrModeShort {
		t.Errorf("addr_mode: got 0x%02X, want 0x%02X", buf[18], zbossAddrModeShort)
	}
}
