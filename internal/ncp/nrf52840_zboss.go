package ncp

// Real ZBOSS NCP serial protocol: LL/HL frame codec, CRC8/CRC16, command IDs.
// Reference: Wireshark ZBOSS NCP dissector (packet-zbncp.c/h).

import (
	"encoding/binary"
	"fmt"
)

// --- LL (Low-Level) header constants ---

const (
	zbossSig0         = 0xDE
	zbossSig1         = 0xAD
	zbossLLHeaderSize = 7 // sig(2) + len(2) + type(1) + flags(1) + crc8(1)
	zbossBodyCRCSize  = 2 // CRC16 at start of body
)

// LL packet type (always 0x06 for ZBOSS NCP API HL; ACK vs DATA is in flags).
const zbossLLType uint8 = 0x06

// LL flags bitmask.
const (
	zbossFlagACK          = 0x01
	zbossFlagRetrans      = 0x02
	zbossFlagPktSeqMask   = 0x0C
	zbossFlagPktSeqShift  = 2
	zbossFlagAckSeqMask   = 0x30
	zbossFlagAckSeqShift  = 4
	zbossFlagFirstFrag    = 0x40
	zbossFlagLastFrag     = 0x80
)

// --- HL (High-Level) header constants ---

const (
	zbossHLVersion    uint8 = 0x00
	zbossHLRequest    uint8 = 0x00
	zbossHLResponse   uint8 = 0x01
	zbossHLIndication uint8 = 0x02
)

// --- Command IDs (call_id, from Wireshark dissector) ---

const (
	// NCP management
	zbossCmdGetModuleVersion    uint16 = 0x0001
	zbossCmdNCPReset            uint16 = 0x0002
	zbossCmdSetZigbeeRole       uint16 = 0x0005
	zbossCmdSetChannelMask      uint16 = 0x0007
	zbossCmdGetChannel          uint16 = 0x0008
	zbossCmdGetPanID            uint16 = 0x0009
	zbossCmdSetPanID            uint16 = 0x000A
	zbossCmdGetLocalIEEE        uint16 = 0x000B
	zbossCmdSetRxOnWhenIdle     uint16 = 0x0013
	zbossCmdSetEDTimeout        uint16 = 0x0017
	zbossCmdSetNwkKey           uint16 = 0x001B
	zbossCmdGetExtPanID         uint16 = 0x0023
	zbossCmdNCPResetInd         uint16 = 0x002B
	zbossCmdSetTCPolicy         uint16 = 0x0032
	zbossCmdSetExtPanID         uint16 = 0x0033
	zbossCmdSetMaxChildren      uint16 = 0x0034

	// AF
	zbossCmdAFSetSimpleDesc uint16 = 0x0101

	// ZDO
	zbossCmdZDOSimpleDescReq    uint16 = 0x0205
	zbossCmdZDOActiveEPReq      uint16 = 0x0206
	zbossCmdZDOBindReq          uint16 = 0x0208
	zbossCmdZDOUnbindReq        uint16 = 0x0209
	zbossCmdZDOMgmtLeaveReq     uint16 = 0x020A
	zbossCmdZDOPermitJoiningReq uint16 = 0x020B
	zbossCmdZDODevAnnceInd      uint16 = 0x020C
	zbossCmdZDODevUpdateInd     uint16 = 0x0215

	// APS
	zbossCmdAPSDEDataReq uint16 = 0x0301
	zbossCmdAPSDEDataInd uint16 = 0x0306

	// NWK
	zbossCmdNwkFormation        uint16 = 0x0401
	zbossCmdNwkDiscovery        uint16 = 0x0402
	zbossCmdNwkGetIEEEByShort   uint16 = 0x0405
	zbossCmdNwkGetShortByIEEE   uint16 = 0x0406
	zbossCmdNwkStartedInd       uint16 = 0x0408
	zbossCmdNwkLeaveInd         uint16 = 0x040B
	zbossCmdNwkAddrUpdateInd    uint16 = 0x041C
	zbossCmdNwkStartWithoutForm uint16 = 0x041D

	// Security indications (diagnostic)
	zbossCmdSecurTCLKInd              uint16 = 0x050E
	zbossCmdSecurTCLKExchangeFailInd  uint16 = 0x050F
	zbossCmdZDODevAuthorizedInd       uint16 = 0x0214
)

// zbossCmdName returns a human-readable name for a ZBOSS command ID.
func zbossCmdName(id uint16) string {
	switch id {
	case zbossCmdGetModuleVersion:
		return "GetModuleVersion"
	case zbossCmdNCPReset:
		return "NCPReset"
	case zbossCmdSetZigbeeRole:
		return "SetZigbeeRole"
	case zbossCmdSetChannelMask:
		return "SetChannelMask"
	case zbossCmdGetChannel:
		return "GetChannel"
	case zbossCmdGetPanID:
		return "GetPanID"
	case zbossCmdSetPanID:
		return "SetPanID"
	case zbossCmdGetLocalIEEE:
		return "GetLocalIEEE"
	case zbossCmdSetRxOnWhenIdle:
		return "SetRxOnWhenIdle"
	case zbossCmdSetEDTimeout:
		return "SetEDTimeout"
	case zbossCmdSetNwkKey:
		return "SetNwkKey"
	case zbossCmdGetExtPanID:
		return "GetExtPanID"
	case zbossCmdNCPResetInd:
		return "NCPResetInd"
	case zbossCmdSetTCPolicy:
		return "SetTCPolicy"
	case zbossCmdSetExtPanID:
		return "SetExtPanID"
	case zbossCmdSetMaxChildren:
		return "SetMaxChildren"
	case zbossCmdAFSetSimpleDesc:
		return "AFSetSimpleDesc"
	case zbossCmdZDOSimpleDescReq:
		return "ZDO_SimpleDesc"
	case zbossCmdZDOActiveEPReq:
		return "ZDO_ActiveEP"
	case zbossCmdZDOBindReq:
		return "ZDO_Bind"
	case zbossCmdZDOUnbindReq:
		return "ZDO_Unbind"
	case zbossCmdZDOMgmtLeaveReq:
		return "ZDO_MgmtLeave"
	case zbossCmdZDOPermitJoiningReq:
		return "ZDO_PermitJoin"
	case zbossCmdZDODevAnnceInd:
		return "ZDO_DevAnnce"
	case zbossCmdZDODevUpdateInd:
		return "ZDO_DevUpdate"
	case zbossCmdAPSDEDataReq:
		return "APSDE_DataReq"
	case zbossCmdAPSDEDataInd:
		return "APSDE_DataInd"
	case zbossCmdNwkFormation:
		return "NwkFormation"
	case zbossCmdNwkDiscovery:
		return "NwkDiscovery"
	case zbossCmdNwkGetIEEEByShort:
		return "NwkGetIEEEByShort"
	case zbossCmdNwkGetShortByIEEE:
		return "NwkGetShortByIEEE"
	case zbossCmdNwkStartedInd:
		return "NwkStartedInd"
	case zbossCmdNwkLeaveInd:
		return "NwkLeaveInd"
	case zbossCmdNwkAddrUpdateInd:
		return "NwkAddrUpdateInd"
	case zbossCmdNwkStartWithoutForm:
		return "NwkStartWithoutForm"
	case zbossCmdSecurTCLKInd:
		return "SECUR_TCLK_IND"
	case zbossCmdSecurTCLKExchangeFailInd:
		return "SECUR_TCLK_EXCHANGE_FAILED_IND"
	case zbossCmdZDODevAuthorizedInd:
		return "ZDO_DevAuthorized"
	default:
		return fmt.Sprintf("0x%04X", id)
	}
}

// zbossStatusName returns a human-readable status description.
func zbossStatusName(cat, code uint8) string {
	if cat == 0 && code == 0 {
		return "OK"
	}
	catName := "Generic"
	switch cat {
	case 2:
		catName = "MAC"
	case 3:
		catName = "NWK"
	case 4:
		catName = "APS"
	case 5:
		catName = "ZDO"
	case 6:
		catName = "CBKE"
	}
	return fmt.Sprintf("%s/%d(0x%02X)", catName, code, code)
}

// Response status categories.
const (
	zbossStatusGeneric uint8 = 0x00
	zbossStatusMAC     uint8 = 0x02
	zbossStatusNWK     uint8 = 0x03
	zbossStatusAPS     uint8 = 0x04
)

// Zigbee roles (ZBOSS DeviceRole enum: ZC=0, ZR=1, ZED=2).
const zbossRoleCoordinator uint8 = 0x00

// ZDO device update status values.
const (
	zbossDevUpdateSecureRejoin uint8 = 0x00
	zbossDevUpdateUnsecureJoin uint8 = 0x01
	zbossDevUpdateLeft         uint8 = 0x02
	zbossDevUpdateTCRejoin     uint8 = 0x03
)

// TC policy types for SET_TC_POLICY (0x0032).
const (
	zbossTCPolicyLinkKeysRequired        uint16 = 0x0000
	zbossTCPolicyICRequired              uint16 = 0x0001
	zbossTCPolicyTCRejoinEnabled         uint16 = 0x0002
	zbossTCPolicyIgnoreTCRejoin          uint16 = 0x0003
	zbossTCPolicyAPSInsecureJoin         uint16 = 0x0004
	zbossTCPolicyDisableNwkMgmtChanUpd   uint16 = 0x0005
)

// APSDE address modes.
const (
	zbossAddrModeShort uint8 = 0x02
	zbossAddrModeIEEE  uint8 = 0x03
)

// ZCL frame control bits.
const (
	zclFrameTypeGlobal     = 0x00
	zclFrameTypeCluster    = 0x01
	zclFlagMfrSpecific     = 0x04
	zclDirServerToClient   = 0x08
	zclDisableDefaultResp  = 0x10
)

// ZCL global command IDs.
const (
	zclCmdReadAttributes     = 0x00
	zclCmdReadAttributesRsp  = 0x01
	zclCmdWriteAttributes    = 0x02
	zclCmdConfigReporting    = 0x06
	zclCmdReportAttributes   = 0x0A
)

// HA profile ID.
const zclProfileHA uint16 = 0x0104

// --- Frame types ---

// zbossLLHeader is the low-level header.
type zbossLLHeader struct {
	Length uint16
	Type   uint8
	Flags  uint8
}

// zbossHLHeader is the high-level header.
type zbossHLHeader struct {
	Version    uint8
	PacketType uint8
	CallID     uint16
	TSN        uint8 // only for Request/Response
	StatusCat  uint8 // only for Response
	StatusCode uint8 // only for Response
}

// zbossFrame is a complete parsed ZBOSS NCP frame (LL + HL + payload).
type zbossFrame struct {
	LL      zbossLLHeader
	HL      zbossHLHeader
	Payload []byte
}

// --- Flag helpers ---

func zbossLLPktSeq(flags uint8) uint8 {
	return (flags >> zbossFlagPktSeqShift) & 0x03
}

func zbossLLAckSeq(flags uint8) uint8 {
	return (flags >> zbossFlagAckSeqShift) & 0x03
}

func zbossLLIsACK(flags uint8) bool {
	return flags&zbossFlagACK != 0
}

// --- CRC-8/KOOP (reflected poly=0xB2 i.e. normal 0x4D, init=0xFF, xorout=0xFF) ---

var crc8Table [256]uint8

func init() {
	const poly = 0xB2 // reflected form of 0x4D
	for i := 0; i < 256; i++ {
		crc := uint8(i)
		for bit := 0; bit < 8; bit++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ poly
			} else {
				crc >>= 1
			}
		}
		crc8Table[i] = crc
	}
}

func zbossCRC8(data []byte) uint8 {
	crc := uint8(0xFF)
	for _, b := range data {
		crc = crc8Table[crc^b]
	}
	return crc ^ 0xFF
}

// --- CRC-16 reflected (poly=0x8408, init=0x0000, xorout=0x0000) ---

func zbossCRC16(data []byte) uint16 {
	crc := uint16(0x0000)
	for _, b := range data {
		crc = (crc >> 8) ^ fcsTable[(crc^uint16(b))&0xFF]
	}
	return crc
}

// --- Encode ---

// zbossEncodeRequest builds a complete ZBOSS frame for an HL request.
// pktSeq is the 2-bit LL packet sequence number.
func zbossEncodeRequest(callID uint16, tsn uint8, pktSeq uint8, payload []byte) []byte {
	// HL header: version(1) + type(1) + callID(2) + tsn(1) = 5 bytes
	hlData := make([]byte, 5+len(payload))
	hlData[0] = zbossHLVersion
	hlData[1] = zbossHLRequest
	binary.LittleEndian.PutUint16(hlData[2:4], callID)
	hlData[4] = tsn
	copy(hlData[5:], payload)

	return zbossEncodeDataFrame(pktSeq, hlData)
}

// zbossEncodeDataFrame wraps HL data in an LL data frame.
func zbossEncodeDataFrame(pktSeq uint8, hlData []byte) []byte {
	bodyCRC := zbossCRC16(hlData)

	// body = CRC16(2) + hlData
	bodyLen := zbossBodyCRCSize + len(hlData)
	// Size includes: size_field(2) + type(1) + flags(1) + crc8(1) + body
	llSize := uint16(5 + bodyLen)

	flags := uint8(zbossFlagFirstFrag | zbossFlagLastFrag)
	flags |= (pktSeq << zbossFlagPktSeqShift) & zbossFlagPktSeqMask

	// Total frame = sig(2) + Size = 2 + llSize
	frame := make([]byte, 2+int(llSize))
	frame[0] = zbossSig0
	frame[1] = zbossSig1
	binary.LittleEndian.PutUint16(frame[2:4], llSize)
	frame[4] = zbossLLType
	frame[5] = flags
	frame[6] = zbossCRC8(frame[2:6]) // CRC8 over size+type+flags

	// Body
	binary.LittleEndian.PutUint16(frame[7:9], bodyCRC)
	copy(frame[9:], hlData)

	return frame
}

// zbossEncodeACK builds an LL ACK frame (7 bytes, no body).
func zbossEncodeACK(ackSeq uint8) []byte {
	frame := make([]byte, zbossLLHeaderSize)
	frame[0] = zbossSig0
	frame[1] = zbossSig1
	binary.LittleEndian.PutUint16(frame[2:4], 5) // size includes itself: size(2)+type(1)+flags(1)+crc8(1)
	frame[4] = zbossLLType
	frame[5] = zbossFlagACK | ((ackSeq << zbossFlagAckSeqShift) & zbossFlagAckSeqMask)
	frame[6] = zbossCRC8(frame[2:6])
	return frame
}

// --- Decode ---

// zbossDecodeFrame parses a complete ZBOSS frame from raw bytes (after HDLC deframing).
func zbossDecodeFrame(data []byte) (*zbossFrame, error) {
	if len(data) < zbossLLHeaderSize {
		return nil, fmt.Errorf("zboss: frame too short: %d bytes", len(data))
	}
	if data[0] != zbossSig0 || data[1] != zbossSig1 {
		return nil, fmt.Errorf("zboss: bad signature: 0x%02X%02X", data[0], data[1])
	}

	llSize := binary.LittleEndian.Uint16(data[2:4])
	llType := data[4]
	llFlags := data[5]
	llCRC := data[6]

	if got := zbossCRC8(data[2:6]); llCRC != got {
		return nil, fmt.Errorf("zboss: LL CRC8 mismatch: got 0x%02X, want 0x%02X", llCRC, got)
	}

	if llType != zbossLLType {
		return nil, fmt.Errorf("zboss: unexpected LL type: 0x%02X", llType)
	}

	// Total frame = sig(2) + Size; verify we have enough data.
	if int(llSize)+2 > len(data) {
		return nil, fmt.Errorf("zboss: frame truncated: need %d, have %d", llSize+2, len(data))
	}

	f := &zbossFrame{
		LL: zbossLLHeader{
			Length: llSize,
			Type:   llType,
			Flags:  llFlags,
		},
	}

	// ACK frames have no body.
	if zbossLLIsACK(llFlags) {
		return f, nil
	}

	body := data[zbossLLHeaderSize : 2+llSize]
	if len(body) < zbossBodyCRCSize {
		return nil, fmt.Errorf("zboss: body too short for CRC16: %d bytes", len(body))
	}

	bodyCRC := binary.LittleEndian.Uint16(body[0:2])
	hlData := body[2:]
	if got := zbossCRC16(hlData); bodyCRC != got {
		return nil, fmt.Errorf("zboss: body CRC16 mismatch: got 0x%04X, want 0x%04X", bodyCRC, got)
	}

	if len(hlData) < 4 {
		return nil, fmt.Errorf("zboss: HL data too short: %d bytes", len(hlData))
	}

	f.HL.Version = hlData[0]
	f.HL.PacketType = hlData[1]
	f.HL.CallID = binary.LittleEndian.Uint16(hlData[2:4])

	pos := 4
	switch f.HL.PacketType {
	case zbossHLRequest:
		if len(hlData) < 5 {
			return nil, fmt.Errorf("zboss: request HL too short for TSN")
		}
		f.HL.TSN = hlData[4]
		pos = 5
	case zbossHLResponse:
		if len(hlData) < 7 {
			return nil, fmt.Errorf("zboss: response HL too short")
		}
		f.HL.TSN = hlData[4]
		f.HL.StatusCat = hlData[5]
		f.HL.StatusCode = hlData[6]
		pos = 7
	case zbossHLIndication:
		pos = 4
	default:
		return nil, fmt.Errorf("zboss: unknown HL packet type: 0x%02X", f.HL.PacketType)
	}

	if pos < len(hlData) {
		f.Payload = make([]byte, len(hlData)-pos)
		copy(f.Payload, hlData[pos:])
	}

	return f, nil
}

// --- ZCL frame helpers ---

// zclBuildReadAttributes builds a ZCL Read Attributes frame.
func zclBuildReadAttributes(seqNum uint8, attrIDs []uint16) []byte {
	buf := make([]byte, 3+len(attrIDs)*2)
	buf[0] = zclFrameTypeGlobal | zclDisableDefaultResp // frame control
	buf[1] = seqNum
	buf[2] = zclCmdReadAttributes
	for i, id := range attrIDs {
		binary.LittleEndian.PutUint16(buf[3+i*2:], id)
	}
	return buf
}

// zclBuildWriteAttributes builds a ZCL Write Attributes frame.
func zclBuildWriteAttributes(seqNum uint8, records []WriteRecord) []byte {
	buf := []byte{
		zclFrameTypeGlobal | zclDisableDefaultResp,
		seqNum,
		zclCmdWriteAttributes,
	}
	for _, rec := range records {
		var rbuf [2]byte
		binary.LittleEndian.PutUint16(rbuf[:], rec.AttrID)
		buf = append(buf, rbuf[0], rbuf[1], rec.DataType)
		buf = append(buf, rec.Value...)
	}
	return buf
}

// zclBuildClusterCommand builds a ZCL cluster-specific command frame.
func zclBuildClusterCommand(seqNum uint8, cmdID uint8, payload []byte) []byte {
	buf := make([]byte, 3+len(payload))
	buf[0] = zclFrameTypeCluster | zclDisableDefaultResp
	buf[1] = seqNum
	buf[2] = cmdID
	copy(buf[3:], payload)
	return buf
}

// zclBuildConfigureReporting builds a ZCL Configure Reporting frame.
func zclBuildConfigureReporting(seqNum uint8, attrID uint16, dataType uint8, minInterval, maxInterval uint16, reportChange []byte) []byte {
	// direction(1) + attrID(2) + dataType(1) + minInterval(2) + maxInterval(2) + reportableChange(N)
	recLen := 1 + 2 + 1 + 2 + 2 + len(reportChange)
	buf := make([]byte, 3+recLen)
	buf[0] = zclFrameTypeGlobal | zclDisableDefaultResp
	buf[1] = seqNum
	buf[2] = zclCmdConfigReporting
	buf[3] = 0x00 // direction: send reports
	binary.LittleEndian.PutUint16(buf[4:6], attrID)
	buf[6] = dataType
	binary.LittleEndian.PutUint16(buf[7:9], minInterval)
	binary.LittleEndian.PutUint16(buf[9:11], maxInterval)
	copy(buf[11:], reportChange)
	return buf
}

// zclParseAttributeReports parses ZCL Report Attributes records from payload.
// Format: [attrID(2) + dataType(1) + value(N)]...
func zclParseAttributeReports(data []byte) []AttributeReportEvent {
	var reports []AttributeReportEvent
	pos := 0
	for pos+3 <= len(data) {
		attrID := binary.LittleEndian.Uint16(data[pos : pos+2])
		dataType := data[pos+2]
		pos += 3

		size := typeSize(dataType)
		var value []byte

		switch {
		case size > 0:
			// Fixed-length type.
			if pos+size > len(data) {
				return reports
			}
			value = make([]byte, size)
			copy(value, data[pos:pos+size])
			pos += size

		case size == typeSizeVariable:
			// Variable-length with 1-byte length prefix (octstr, string).
			if pos >= len(data) {
				return reports
			}
			vlen := int(data[pos])
			if pos+1+vlen > len(data) {
				return reports
			}
			value = make([]byte, 1+vlen)
			copy(value, data[pos:pos+1+vlen])
			pos += 1 + vlen

		case size == typeSizeVariable16:
			// Variable-length with 2-byte length prefix (octstr16, string16).
			if pos+2 > len(data) {
				return reports
			}
			vlen := int(binary.LittleEndian.Uint16(data[pos : pos+2]))
			if pos+2+vlen > len(data) {
				return reports
			}
			value = make([]byte, 2+vlen)
			copy(value, data[pos:pos+2+vlen])
			pos += 2 + vlen

		default:
			// Unknown type — can't determine boundaries, stop parsing.
			return reports
		}

		reports = append(reports, AttributeReportEvent{
			AttrID:   attrID,
			DataType: dataType,
			Value:    value,
		})
	}
	return reports
}

// buildAPSDEDataReq builds the APSDE_DATA_REQ payload.
func buildAPSDEDataReq(dstAddr uint16, dstEP, srcEP uint8, clusterID, profileID uint16, radius uint8, apsData []byte) []byte {
	// param_len(1) + data_len(2) + dst_addr(8) + profile_id(2) + cluster_id(2) +
	// dst_endpoint(1) + src_endpoint(1) + radius(1) + dst_addr_mode(1) +
	// tx_options(1) + use_alias(1) + alias_src_addr(2) + alias_seq_num(1) + data
	const fixedLen = 24
	buf := make([]byte, fixedLen+len(apsData))
	buf[0] = fixedLen - 3 // param_len: fixed params excluding param_len+data_len fields
	binary.LittleEndian.PutUint16(buf[1:3], uint16(len(apsData)))
	// dst_addr: 8-byte union, short addr in first 2 bytes
	binary.LittleEndian.PutUint16(buf[3:5], dstAddr)
	// bytes 5-10 are zero padding for short addr mode
	binary.LittleEndian.PutUint16(buf[11:13], profileID)
	binary.LittleEndian.PutUint16(buf[13:15], clusterID)
	buf[15] = dstEP
	buf[16] = srcEP
	buf[17] = radius
	buf[18] = zbossAddrModeShort // dst_addr_mode
	buf[19] = 0x04              // tx_options: APS ACK (bit2)
	buf[20] = 0x00              // use_alias
	// alias_src_addr(2) + alias_seq_num(1) at 21-23 = 0
	copy(buf[24:], apsData)
	return buf
}
