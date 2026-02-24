package ncp

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestZCLBuildReadAttributes(t *testing.T) {
	attrs := []uint16{0x0000, 0x0001}
	frame := zclBuildReadAttributes(5, attrs)

	if frame[0]&0x03 != zclFrameTypeGlobal {
		t.Errorf("frame type: got 0x%02X, want global", frame[0]&0x03)
	}
	if frame[1] != 5 {
		t.Errorf("seq: got %d, want 5", frame[1])
	}
	if frame[2] != zclCmdReadAttributes {
		t.Errorf("cmd: got 0x%02X, want 0x%02X", frame[2], zclCmdReadAttributes)
	}
	if binary.LittleEndian.Uint16(frame[3:5]) != 0x0000 {
		t.Error("attr[0] mismatch")
	}
	if binary.LittleEndian.Uint16(frame[5:7]) != 0x0001 {
		t.Error("attr[1] mismatch")
	}
}

func TestZCLBuildWriteAttributes(t *testing.T) {
	records := []WriteRecord{
		{AttrID: 0x0100, DataType: 0x20, Value: []byte{0x42}},
	}
	frame := zclBuildWriteAttributes(10, records)

	if frame[2] != zclCmdWriteAttributes {
		t.Errorf("cmd: got 0x%02X, want 0x%02X", frame[2], zclCmdWriteAttributes)
	}
	// attr_id(2) + data_type(1) + value(1) at offset 3
	attrID := binary.LittleEndian.Uint16(frame[3:5])
	if attrID != 0x0100 {
		t.Errorf("attrID: got 0x%04X, want 0x0100", attrID)
	}
	if frame[5] != 0x20 {
		t.Errorf("dataType: got 0x%02X, want 0x20", frame[5])
	}
	if frame[6] != 0x42 {
		t.Errorf("value: got 0x%02X, want 0x42", frame[6])
	}
}

func TestZCLBuildClusterCommand(t *testing.T) {
	frame := zclBuildClusterCommand(7, 0x01, []byte{0xFF})
	if frame[0]&0x03 != zclFrameTypeCluster {
		t.Errorf("frame type: got 0x%02X, want cluster", frame[0]&0x03)
	}
	if frame[2] != 0x01 {
		t.Errorf("cmdID: got 0x%02X, want 0x01", frame[2])
	}
	if frame[3] != 0xFF {
		t.Errorf("payload: got 0x%02X, want 0xFF", frame[3])
	}
}

func TestZCLBuildConfigureReporting(t *testing.T) {
	change := []byte{0x01, 0x00}
	frame := zclBuildConfigureReporting(3, 0x0000, 0x29, 10, 300, change)

	if frame[2] != zclCmdConfigReporting {
		t.Errorf("cmd: got 0x%02X", frame[2])
	}
	if frame[3] != 0x00 {
		t.Errorf("direction: got 0x%02X, want 0x00", frame[3])
	}
	attrID := binary.LittleEndian.Uint16(frame[4:6])
	if attrID != 0x0000 {
		t.Errorf("attrID: got 0x%04X", attrID)
	}
	minI := binary.LittleEndian.Uint16(frame[7:9])
	if minI != 10 {
		t.Errorf("minInterval: got %d, want 10", minI)
	}
	maxI := binary.LittleEndian.Uint16(frame[9:11])
	if maxI != 300 {
		t.Errorf("maxInterval: got %d, want 300", maxI)
	}
}

func TestZCLParseAttributeReports(t *testing.T) {
	// attrID=0x0000 dataType=0x29(int16) value=0x1234
	data := []byte{0x00, 0x00, 0x29, 0x34, 0x12}
	reports := zclParseAttributeReports(data)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].AttrID != 0x0000 {
		t.Errorf("attrID: got 0x%04X", reports[0].AttrID)
	}
	if reports[0].DataType != 0x29 {
		t.Errorf("dataType: got 0x%02X", reports[0].DataType)
	}
	if !bytes.Equal(reports[0].Value, []byte{0x34, 0x12}) {
		t.Errorf("value: got %X", reports[0].Value)
	}
}

func TestZCLParseMultipleAttributeReports(t *testing.T) {
	// Two uint8 attributes.
	data := []byte{
		0x00, 0x00, 0x20, 0xAA, // attrID=0x0000, uint8, val=0xAA
		0x01, 0x00, 0x20, 0xBB, // attrID=0x0001, uint8, val=0xBB
	}
	reports := zclParseAttributeReports(data)
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
	if reports[0].Value[0] != 0xAA || reports[1].Value[0] != 0xBB {
		t.Errorf("values: %X, %X", reports[0].Value, reports[1].Value)
	}
}

func TestHandleAPSDEDataInd(t *testing.T) {
	// Build a minimal APSDE_DATA_IND payload with a ZCL Report Attributes frame.
	zclReport := []byte{
		zclFrameTypeGlobal | zclDirServerToClient, // frame control
		0x01,                  // zcl seq
		zclCmdReportAttributes, // cmd
		0x00, 0x00,            // attrID=0x0000
		0x20,                  // dataType=uint8
		0x42,                  // value
	}

	payload := make([]byte, 24+len(zclReport))
	payload[0] = 21                                                       // param_len
	binary.LittleEndian.PutUint16(payload[1:3], uint16(len(zclReport)))   // data_len
	payload[3] = 0x00                                                     // aps_fc
	binary.LittleEndian.PutUint16(payload[4:6], 0x1234)                   // src_nwk_addr
	payload[11] = 1                                                       // src_endpoint
	binary.LittleEndian.PutUint16(payload[12:14], 0x0402)                 // cluster_id (temperature)
	binary.LittleEndian.PutUint16(payload[14:16], zclProfileHA)           // profile_id
	copy(payload[24:], zclReport)

	var gotReport AttributeReportEvent
	called := false

	n := &NRF52840NCP{
		zclPending: make(map[uint8]chan []byte),
	}
	n.handleAPSDEDataInd(payload, func(evt AttributeReportEvent) {
		gotReport = evt
		called = true
	})

	if !called {
		t.Fatal("onReport not called")
	}
	if gotReport.SrcAddr != 0x1234 {
		t.Errorf("SrcAddr: got 0x%04X, want 0x1234", gotReport.SrcAddr)
	}
	if gotReport.SrcEP != 1 {
		t.Errorf("SrcEP: got %d, want 1", gotReport.SrcEP)
	}
	if gotReport.ClusterID != 0x0402 {
		t.Errorf("ClusterID: got 0x%04X, want 0x0402", gotReport.ClusterID)
	}
	if gotReport.AttrID != 0x0000 {
		t.Errorf("AttrID: got 0x%04X", gotReport.AttrID)
	}
	if gotReport.DataType != 0x20 {
		t.Errorf("DataType: got 0x%02X", gotReport.DataType)
	}
	if len(gotReport.Value) != 1 || gotReport.Value[0] != 0x42 {
		t.Errorf("Value: got %X, want [42]", gotReport.Value)
	}
}

func TestZCLParseAttributeReportsWithString(t *testing.T) {
	// attrID=0x0004(manufacturer) dataType=0x42(string) len=3 "ABC" + attrID=0x0000 uint8 0xFF
	data := []byte{
		0x04, 0x00, 0x42, 0x03, 'A', 'B', 'C', // string attr
		0x00, 0x00, 0x20, 0xFF, // uint8 attr after the string
	}
	reports := zclParseAttributeReports(data)
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
	// String value includes length prefix.
	if !bytes.Equal(reports[0].Value, []byte{0x03, 'A', 'B', 'C'}) {
		t.Errorf("string value: got %X", reports[0].Value)
	}
	if reports[1].AttrID != 0x0000 || reports[1].Value[0] != 0xFF {
		t.Errorf("uint8 after string: attrID=0x%04X value=%X", reports[1].AttrID, reports[1].Value)
	}
}

func TestZCLParseAttributeReportsOctstr16(t *testing.T) {
	// attrID=0x0001 dataType=0x43(octstr16) len=2 [0xAA, 0xBB]
	data := []byte{
		0x01, 0x00, 0x43, 0x02, 0x00, 0xAA, 0xBB,
	}
	reports := zclParseAttributeReports(data)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	// Value includes 2-byte length prefix.
	if !bytes.Equal(reports[0].Value, []byte{0x02, 0x00, 0xAA, 0xBB}) {
		t.Errorf("octstr16 value: got %X", reports[0].Value)
	}
}

func TestHandleAPSDEDataIndMfrSpecific(t *testing.T) {
	// Manufacturer-specific ZCL Report Attributes frame.
	zclReport := []byte{
		zclFrameTypeGlobal | zclFlagMfrSpecific | zclDirServerToClient, // frame control (mfr-specific)
		0x5E, 0x11, // manufacturer code (0x115E = Xiaomi)
		0x01,                   // zcl seq
		zclCmdReportAttributes, // cmd
		0x00, 0x00,             // attrID=0x0000
		0x20,                   // dataType=uint8
		0x55,                   // value
	}

	payload := make([]byte, 24+len(zclReport))
	payload[0] = 21
	binary.LittleEndian.PutUint16(payload[1:3], uint16(len(zclReport)))
	binary.LittleEndian.PutUint16(payload[4:6], 0x5678) // src_nwk_addr
	payload[11] = 1                                      // src_endpoint
	binary.LittleEndian.PutUint16(payload[12:14], 0x0000) // cluster_id (basic)
	binary.LittleEndian.PutUint16(payload[14:16], zclProfileHA)
	copy(payload[24:], zclReport)

	var gotReport AttributeReportEvent
	called := false

	n := &NRF52840NCP{
		zclPending: make(map[uint8]chan []byte),
	}
	n.handleAPSDEDataInd(payload, func(evt AttributeReportEvent) {
		gotReport = evt
		called = true
	})

	if !called {
		t.Fatal("onReport not called for mfr-specific frame")
	}
	if gotReport.SrcAddr != 0x5678 {
		t.Errorf("SrcAddr: got 0x%04X, want 0x5678", gotReport.SrcAddr)
	}
	if gotReport.AttrID != 0x0000 {
		t.Errorf("AttrID: got 0x%04X, want 0x0000", gotReport.AttrID)
	}
	if len(gotReport.Value) != 1 || gotReport.Value[0] != 0x55 {
		t.Errorf("Value: got %X, want [55]", gotReport.Value)
	}
}

func TestHandleAPSDEDataIndReadResponse(t *testing.T) {
	// ZCL Read Attributes Response for ZCL seq=0x07.
	zclResp := []byte{
		zclFrameTypeGlobal | zclDirServerToClient | zclDisableDefaultResp, // frame control
		0x07,                    // zcl seq
		zclCmdReadAttributesRsp, // cmd
		// Record: attrID=0x0004, status=0x00, type=0x42(string), len=5, "Hello"
		0x04, 0x00, 0x00, 0x42, 0x05, 'H', 'e', 'l', 'l', 'o',
	}

	payload := make([]byte, 24+len(zclResp))
	payload[0] = 21
	binary.LittleEndian.PutUint16(payload[1:3], uint16(len(zclResp)))
	binary.LittleEndian.PutUint16(payload[4:6], 0xAAAA)
	payload[11] = 1
	binary.LittleEndian.PutUint16(payload[12:14], 0x0000)
	binary.LittleEndian.PutUint16(payload[14:16], zclProfileHA)
	copy(payload[24:], zclResp)

	// Register a pending ZCL response channel for seq=0x07.
	ch := make(chan []byte, 1)
	n := &NRF52840NCP{
		zclPending: map[uint8]chan []byte{0x07: ch},
	}
	n.handleAPSDEDataInd(payload, nil)

	select {
	case data := <-ch:
		// Should receive the records portion (after ZCL header).
		results := parseAttributeResponses(data)
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].AttrID != 0x0004 {
			t.Errorf("attrID: got 0x%04X, want 0x0004", results[0].AttrID)
		}
		if results[0].Status != 0x00 {
			t.Errorf("status: got 0x%02X, want 0x00", results[0].Status)
		}
	default:
		t.Fatal("no data received on zclPending channel")
	}
}

func TestBuildSimpleDescPayload(t *testing.T) {
	in := []uint16{0x0000, 0x0006}
	out := []uint16{0x0006}
	buf := buildSimpleDescPayload(1, 0x0104, 0x0005, 0, in, out)

	if buf[0] != 1 {
		t.Errorf("endpoint: got %d", buf[0])
	}
	profileID := binary.LittleEndian.Uint16(buf[1:3])
	if profileID != 0x0104 {
		t.Errorf("profileID: got 0x%04X", profileID)
	}
	if buf[6] != 2 {
		t.Errorf("in_count: got %d, want 2", buf[6])
	}
	if buf[7] != 1 {
		t.Errorf("out_count: got %d, want 1", buf[7])
	}
	if len(buf) != 8+2*2+1*2 {
		t.Errorf("total length: got %d, want %d", len(buf), 8+4+2)
	}
}
