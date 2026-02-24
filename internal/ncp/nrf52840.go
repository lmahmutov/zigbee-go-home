package ncp

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.bug.st/serial"
)

// NRF52840NCP implements NCP using nRF52840 with real ZBOSS NCP protocol (HDLC framing).
type NRF52840NCP struct {
	port     serial.Port
	portName string
	portMode *serial.Mode
	reader   *bufio.Reader
	logger   *slog.Logger

	// HL-level request/response tracking (keyed by TSN).
	hlTSN     atomic.Uint32
	hlPending map[uint8]chan *zbossFrame
	hlMu      sync.Mutex

	// LL-level packet sequencing and ACK.
	llPktSeq uint8 // our 2-bit send sequence
	llSeqMu  sync.Mutex
	llAckCh  chan uint8 // received ACK sequence numbers
	writeMu  sync.Mutex

	// ZCL sequence number for ZCL frames.
	zclSeq atomic.Uint32

	// ZCL-level read response tracking (keyed by ZCL sequence number).
	zclPending map[uint8]chan []byte
	zclMu      sync.Mutex

	// Indication callbacks.
	handlerMu    sync.RWMutex
	onJoined     func(DeviceJoinedEvent)
	onLeft       func(DeviceLeftEvent)
	onAnnounce   func(DeviceAnnounceEvent)
	onReport     func(AttributeReportEvent)
	onClusterCmd    func(ClusterCommandEvent)
	onNwkAddrUpdate func(uint16)
	onReset         func()

	// Signaled when NCPResetInd is received (used by resetAndReconnect).
	resetIndCh chan struct{}

	ncpInfo NCPInfo

	// lifecycleMu protects concurrent resetState/Close access to port, done,
	// llAckCh, closeOnce. Must be held when transitioning between states.
	lifecycleMu sync.Mutex
	done        chan struct{}
	closeOnce   sync.Once
	closed      bool // set after final Close, prevents resetState on closed NCP
	wg          sync.WaitGroup
}

// NewNRF52840NCP creates a new nRF52840 NCP backend.
func NewNRF52840NCP(portName string, baudRate int, logger *slog.Logger) (*NRF52840NCP, error) {
	mode := &serial.Mode{
		BaudRate: baudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, fmt.Errorf("nrf52840 ncp: open %s: %w", portName, err)
	}

	// USB CDC ACM: assert DTR/RTS for NCP firmware.
	_ = port.SetDTR(true)
	_ = port.SetRTS(true)

	n := &NRF52840NCP{
		port:       port,
		portName:   portName,
		portMode:   mode,
		reader:     bufio.NewReader(port),
		logger:     logger,
		hlPending:  make(map[uint8]chan *zbossFrame),
		zclPending: make(map[uint8]chan []byte),
		llAckCh:    make(chan uint8, 4),
		resetIndCh: make(chan struct{}, 1),
		done:       make(chan struct{}),
	}
	n.wg.Add(1)
	go n.readLoop()
	return n, nil
}

// nextTSN allocates the next HL transaction sequence number.
func (n *NRF52840NCP) nextTSN() uint8 {
	return uint8(n.hlTSN.Add(1))
}

// nextZCLSeq allocates the next ZCL sequence number.
func (n *NRF52840NCP) nextZCLSeq() uint8 {
	return uint8(n.zclSeq.Add(1))
}

// nextPktSeq advances the LL packet sequence (cycles 1→2→3→1).
func (n *NRF52840NCP) nextPktSeq() uint8 {
	n.llSeqMu.Lock()
	n.llPktSeq = n.llPktSeq%3 + 1
	seq := n.llPktSeq
	n.llSeqMu.Unlock()
	return seq
}

// --- Transport: write with LL ACK ---

const (
	llACKTimeout  = 500 * time.Millisecond
	llMaxRetries  = 3
	hlRespTimeout = 5 * time.Second
)

// request sends an HL request and waits for the HL response.
func (n *NRF52840NCP) request(ctx context.Context, callID uint16, payload []byte) (*zbossFrame, error) {
	tsn := n.nextTSN()

	ch := make(chan *zbossFrame, 1)
	n.hlMu.Lock()
	n.hlPending[tsn] = ch
	n.hlMu.Unlock()
	defer func() {
		n.hlMu.Lock()
		delete(n.hlPending, tsn)
		n.hlMu.Unlock()
	}()

	pktSeq := n.nextPktSeq()
	raw := zbossEncodeRequest(callID, tsn, pktSeq, payload)

	// Write with LL ACK retry.
	if err := n.writeWithACK(ctx, raw, pktSeq); err != nil {
		return nil, fmt.Errorf("nrf write cmd 0x%04X: %w", callID, err)
	}

	cmdName := zbossCmdName(callID)
	n.logger.Info("zboss TX", "cmd", cmdName, "tsn", tsn, "payload", fmt.Sprintf("%X", payload))

	// Wait for HL response.
	select {
	case resp := <-ch:
		if resp == nil {
			return nil, fmt.Errorf("ncp reset: request cancelled")
		}
		status := zbossStatusName(resp.HL.StatusCat, resp.HL.StatusCode)
		if resp.HL.StatusCat != 0 || resp.HL.StatusCode != 0 {
			n.logger.Warn("zboss RX", "cmd", cmdName, "tsn", tsn, "status", status, "payload", fmt.Sprintf("%X", resp.Payload))
			return resp, fmt.Errorf("zboss %s: %s", cmdName, status)
		}
		n.logger.Info("zboss RX", "cmd", cmdName, "tsn", tsn, "status", status, "payload", fmt.Sprintf("%X", resp.Payload))
		return resp, nil
	case <-ctx.Done():
		n.logger.Warn("zboss timeout", "cmd", cmdName, "tsn", tsn, "err", ctx.Err())
		return nil, ctx.Err()
	case <-n.done:
		return nil, fmt.Errorf("ncp closed")
	}
}

// writeWithACK writes a raw ZBOSS frame and waits for LL ACK with retries.
func (n *NRF52840NCP) writeWithACK(ctx context.Context, frame []byte, pktSeq uint8) error {
	for attempt := 0; attempt <= llMaxRetries; attempt++ {
		n.writeMu.Lock()
		_, err := n.port.Write(frame)
		n.writeMu.Unlock()
		if err != nil {
			return fmt.Errorf("serial write: %w", err)
		}
		n.logger.Debug("nrf52840 frame sent", "len", len(frame))

		// Wait for matching ACK, draining stale ACKs within the timeout window.
		deadline := time.NewTimer(llACKTimeout)
	waitACK:
		for {
			select {
			case ackSeq := <-n.llAckCh:
				if ackSeq == pktSeq {
					deadline.Stop()
					return nil
				}
				// Wrong ACK seq (stale from previous frame), drain and keep waiting.
				n.logger.Debug("zboss LL stale ACK drained", "got", ackSeq, "want", pktSeq)
			case <-deadline.C:
				n.logger.Warn("zboss LL ACK timeout", "attempt", attempt+1, "pktSeq", pktSeq)
				break waitACK
			case <-ctx.Done():
				deadline.Stop()
				return ctx.Err()
			case <-n.done:
				deadline.Stop()
				return fmt.Errorf("ncp closed")
			}
		}
	}
	return fmt.Errorf("zboss LL ACK timeout after %d retries", llMaxRetries+1)
}

// sendACK sends an LL ACK for the given packet sequence.
func (n *NRF52840NCP) sendACK(pktSeq uint8) {
	raw := zbossEncodeACK(pktSeq)
	n.writeMu.Lock()
	_, err := n.port.Write(raw)
	n.writeMu.Unlock()
	if err != nil {
		n.logger.Error("zboss send ACK failed", "err", err)
	}
}

// --- Transport: read loop ---

func (n *NRF52840NCP) readLoop() {
	defer n.wg.Done()

	backoff := 10 * time.Millisecond
	const maxBackoff = 5 * time.Second

	for {
		select {
		case <-n.done:
			return
		default:
		}

		raw, err := readRawZBOSSFrame(n.reader)
		if err != nil {
			select {
			case <-n.done:
				return
			default:
				if err != io.EOF && !strings.Contains(err.Error(), "closed") {
					n.logger.Error("nrf52840 read error", "err", err)
				}
				select {
				case <-time.After(backoff):
				case <-n.done:
					return
				}
				if backoff < maxBackoff {
					backoff *= 2
					if backoff > maxBackoff {
						backoff = maxBackoff
					}
				}
				continue
			}
		}
		backoff = 10 * time.Millisecond

		n.logger.Debug("nrf52840 frame received", "len", len(raw))

		frame, err := zbossDecodeFrame(raw)
		if err != nil {
			n.logger.Warn("nrf52840 zboss decode error", "err", err)
			continue
		}

		// Handle LL ACK frames.
		if zbossLLIsACK(frame.LL.Flags) {
			ackSeq := zbossLLAckSeq(frame.LL.Flags)
			n.logger.Debug("zboss LL ACK received", "ackSeq", ackSeq)
			select {
			case n.llAckCh <- ackSeq:
			default:
			}
			continue
		}

		// Data frame — send LL ACK.
		pktSeq := zbossLLPktSeq(frame.LL.Flags)
		n.sendACK(pktSeq)

		n.logger.Debug("zboss frame received",
			"cmd", fmt.Sprintf("0x%04X", frame.HL.CallID),
			"type", frame.HL.PacketType,
			"tsn", frame.HL.TSN)

		switch frame.HL.PacketType {
		case zbossHLResponse:
			n.hlMu.Lock()
			ch, ok := n.hlPending[frame.HL.TSN]
			n.hlMu.Unlock()
			if ok {
				select {
				case ch <- frame:
				default:
				}
			} else {
				status := zbossStatusName(frame.HL.StatusCat, frame.HL.StatusCode)
				n.logger.Warn("zboss orphaned response (too late)",
					"cmd", zbossCmdName(frame.HL.CallID),
					"tsn", frame.HL.TSN,
					"status", status,
					"payload", fmt.Sprintf("%X", frame.Payload))
			}

		case zbossHLIndication:
			n.handleIndication(frame)
		}
	}
}

// --- Indication handlers ---

func (n *NRF52840NCP) handleIndication(f *zbossFrame) {
	n.handlerMu.RLock()
	onJoined := n.onJoined
	onLeft := n.onLeft
	onAnnounce := n.onAnnounce
	onReport := n.onReport
	onClusterCmd := n.onClusterCmd
	onNwkAddrUpdate := n.onNwkAddrUpdate
	onReset := n.onReset
	n.handlerMu.RUnlock()

	switch f.HL.CallID {
	case zbossCmdZDODevAnnceInd:
		// Payload: nwk_addr(2) + ieee(8) + capability(1)
		if onAnnounce != nil && len(f.Payload) >= 11 {
			evt := DeviceAnnounceEvent{
				ShortAddr:  binary.LittleEndian.Uint16(f.Payload[0:2]),
				Capability: f.Payload[10],
			}
			copy(evt.IEEEAddr[:], f.Payload[2:10])
			onAnnounce(evt)
		}

	case zbossCmdZDODevUpdateInd:
		// Payload: ieee(8) + nwk_addr(2) + status(1)
		if len(f.Payload) >= 11 {
			var ieee [8]byte
			copy(ieee[:], f.Payload[0:8])
			shortAddr := binary.LittleEndian.Uint16(f.Payload[8:10])
			status := f.Payload[10]

			statusName := "unknown"
			switch status {
			case zbossDevUpdateSecureRejoin:
				statusName = "secure_rejoin"
			case zbossDevUpdateUnsecureJoin:
				statusName = "unsecure_join"
			case zbossDevUpdateLeft:
				statusName = "left"
			case zbossDevUpdateTCRejoin:
				statusName = "tc_rejoin"
			}
			n.logger.Info("DevUpdateInd", "ieee", fmt.Sprintf("%016X", ieee),
				"short", fmt.Sprintf("0x%04X", shortAddr), "status", statusName)

			switch status {
			case zbossDevUpdateSecureRejoin, zbossDevUpdateUnsecureJoin, zbossDevUpdateTCRejoin:
				if onJoined != nil {
					onJoined(DeviceJoinedEvent{ShortAddr: shortAddr, IEEEAddr: ieee})
				}
			case zbossDevUpdateLeft:
				if onLeft != nil {
					onLeft(DeviceLeftEvent{ShortAddr: shortAddr, IEEEAddr: ieee})
				}
			default:
				n.logger.Warn("DevUpdateInd unknown status", "status", status)
			}
		}

	case zbossCmdNwkLeaveInd:
		// Payload: ieee(8) + rejoin(1)
		if len(f.Payload) >= 8 {
			var ieee [8]byte
			copy(ieee[:], f.Payload[0:8])
			rejoin := false
			if len(f.Payload) >= 9 {
				rejoin = f.Payload[8] != 0
			}
			n.logger.Info("NwkLeaveInd", "ieee", fmt.Sprintf("%016X", ieee), "rejoin", rejoin)
			if onLeft != nil && !rejoin {
				onLeft(DeviceLeftEvent{IEEEAddr: ieee})
			}
		}

	case zbossCmdAPSDEDataInd:
		n.handleAPSDEDataInd(f.Payload, onReport, onClusterCmd)

	case zbossCmdNCPResetInd:
		n.logger.Warn("NCPResetInd received")
		select {
		case n.resetIndCh <- struct{}{}:
		default:
		}
		if onReset != nil {
			onReset()
		}

	case zbossCmdSecurTCLKInd:
		// TCLK was successfully exchanged with a device.
		if len(f.Payload) >= 8 {
			var ieee [8]byte
			copy(ieee[:], f.Payload[0:8])
			n.logger.Info("SECUR_TCLK_IND: TC link key exchanged", "ieee", fmt.Sprintf("%016X", ieee))
		} else {
			n.logger.Info("SECUR_TCLK_IND", "payload", fmt.Sprintf("%X", f.Payload))
		}

	case zbossCmdSecurTCLKExchangeFailInd:
		// TCLK exchange failed. Payload: status_category(1) + status_code(1).
		if len(f.Payload) >= 2 {
			status := zbossStatusName(f.Payload[0], f.Payload[1])
			n.logger.Error("SECUR_TCLK_EXCHANGE_FAILED", "status", status)
		} else {
			n.logger.Error("SECUR_TCLK_EXCHANGE_FAILED", "payload", fmt.Sprintf("%X", f.Payload))
		}

	case zbossCmdZDODevAuthorizedInd:
		if len(f.Payload) >= 10 {
			var ieee [8]byte
			copy(ieee[:], f.Payload[0:8])
			n.logger.Info("ZDO_DevAuthorized", "ieee", fmt.Sprintf("%016X", ieee))
		} else {
			n.logger.Info("ZDO_DevAuthorized", "payload", fmt.Sprintf("%X", f.Payload))
		}

	case zbossCmdNwkAddrUpdateInd:
		// Payload: nwk_addr(2) — a device changed its short address.
		if len(f.Payload) >= 2 {
			newAddr := binary.LittleEndian.Uint16(f.Payload[0:2])
			n.logger.Warn("NwkAddrUpdateInd: device changed short address", "new_short", fmt.Sprintf("0x%04X", newAddr))
			if onNwkAddrUpdate != nil {
				onNwkAddrUpdate(newAddr)
			}
		}

	default:
		n.logger.Warn("zboss unhandled indication",
			"cmd", zbossCmdName(f.HL.CallID),
			"payload", fmt.Sprintf("%X", f.Payload))
	}
}

// handleAPSDEDataInd parses APSDE_DATA_IND and dispatches ZCL responses, reports, and cluster commands.
func (n *NRF52840NCP) handleAPSDEDataInd(payload []byte, onReport func(AttributeReportEvent), onClusterCmd func(ClusterCommandEvent)) {
	if len(payload) < 25 {
		return
	}
	// param_len(1) + data_len(2) + aps_fc(1) + src_nwk_addr(2) + dst_nwk_addr(2) +
	// group_addr(2) + dst_endpoint(1) + src_endpoint(1) + cluster_id(2) + profile_id(2) +
	// aps_counter(1) + src_mac_addr(2) + dst_mac_addr(2) + lqi(1) + rssi(1) + aps_key_attr(1) + data[]
	dataLen := binary.LittleEndian.Uint16(payload[1:3])
	srcAddr := binary.LittleEndian.Uint16(payload[4:6])
	srcEP := payload[11]
	clusterID := binary.LittleEndian.Uint16(payload[12:14])
	lqi := payload[21]
	rssi := int8(payload[22])

	const apsHdrSize = 24
	if int(dataLen) == 0 || len(payload) < apsHdrSize+int(dataLen) {
		return
	}
	zclData := payload[apsHdrSize : apsHdrSize+int(dataLen)]

	// Parse ZCL frame header, accounting for manufacturer-specific frames.
	// Format: frame_control(1) + [mfr_code(2)] + seq(1) + cmd_id(1)
	if len(zclData) < 3 {
		return
	}
	frameCtrl := zclData[0]
	hdrLen := 3 // frame_control + seq + cmd_id
	if frameCtrl&zclFlagMfrSpecific != 0 {
		hdrLen += 2 // manufacturer code
	}
	if len(zclData) < hdrLen {
		return
	}
	zclSeq := zclData[hdrLen-2]
	cmdID := zclData[hdrLen-1]

	frameType := frameCtrl & 0x03

	// Handle cluster-specific commands (e.g., OTA QueryNextImageRequest, Tuya DP).
	if frameType == zclFrameTypeCluster {
		if clusterID == 0x0019 && cmdID == 0x01 {
			// OTA QueryNextImageRequest — respond with NO_IMAGE_AVAILABLE (0x98).
			// Must run in a goroutine: request() waits for ACK/response that
			// can only be delivered by this readLoop, so calling it inline deadlocks.
			n.logger.Info("OTA query from device, responding NO_IMAGE_AVAILABLE",
				"short", fmt.Sprintf("0x%04X", srcAddr), "ep", srcEP)
			go n.sendOTANoImageAvailable(srcAddr, srcEP, zclSeq)
		} else if onClusterCmd != nil {
			onClusterCmd(ClusterCommandEvent{
				SrcAddr:   srcAddr,
				SrcEP:     srcEP,
				ClusterID: clusterID,
				CommandID: cmdID,
				Payload:   zclData[hdrLen:],
				LQI:       lqi,
				RSSI:      rssi,
			})
		}
		return
	}

	if frameType != zclFrameTypeGlobal {
		return
	}

	records := zclData[hdrLen:]

	switch cmdID {
	case zclCmdReadAttributesRsp:
		// Dispatch to pending ReadAttributes caller by ZCL sequence number.
		n.zclMu.Lock()
		ch, ok := n.zclPending[zclSeq]
		n.zclMu.Unlock()
		if ok {
			select {
			case ch <- records:
			default:
			}
		}

	case zclCmdReportAttributes:
		if onReport == nil {
			return
		}
		reports := zclParseAttributeReports(records)
		for _, rpt := range reports {
			rpt.SrcAddr = srcAddr
			rpt.SrcEP = srcEP
			rpt.ClusterID = clusterID
			rpt.LQI = lqi
			rpt.RSSI = rssi
			onReport(rpt)
		}
	}
}

// sendOTANoImageAvailable responds to an OTA QueryNextImageRequest with NO_IMAGE_AVAILABLE.
func (n *NRF52840NCP) sendOTANoImageAvailable(dstAddr uint16, dstEP uint8, zclSeq uint8) {
	// OTA QueryNextImageResponse (cmd 0x02): status(1) = 0x98 (NO_IMAGE_AVAILABLE)
	// Direction: server-to-client (coordinator is the OTA server).
	zclFrame := make([]byte, 4)
	zclFrame[0] = zclFrameTypeCluster | zclDirServerToClient | zclDisableDefaultResp
	zclFrame[1] = zclSeq
	zclFrame[2] = 0x02 // QueryNextImageResponse
	zclFrame[3] = 0x98 // NO_IMAGE_AVAILABLE
	apsPayload := buildAPSDEDataReq(dstAddr, dstEP, 1, 0x0019, zclProfileHA, 30, zclFrame)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := n.request(ctx, zbossCmdAPSDEDataReq, apsPayload); err != nil {
		n.logger.Warn("OTA no-image response failed", "err", err)
	}
}

// --- NCP interface: Network management ---

// ZBOSS NCP reset options.
const (
	zbossResetNoOption    uint8 = 0x00
	zbossResetEraseNVRAM  uint8 = 0x01
	zbossResetFactory     uint8 = 0x02
)

// resetAndReconnect sends an NCP reset command and waits for USB to re-enumerate.
// After reset the nRF52840 USB device disconnects and reconnects, so we must
// close the old serial port and reopen it.
func (n *NRF52840NCP) resetAndReconnect(ctx context.Context, option uint8) error {
	optName := "reset"
	if option == zbossResetFactory {
		optName = "factory reset"
	}

	// Send reset with all 3 possible LL packet sequences. After a process
	// restart the NCP's expected sequence is unknown (stale from the previous
	// session), so only the matching one will be accepted. Fire-and-forget:
	// the NCP reboots immediately on a valid reset, ACK may never arrive.
	tsn := n.nextTSN()
	for _, seq := range []uint8{1, 2, 3} {
		raw := zbossEncodeRequest(zbossCmdNCPReset, tsn, seq, []byte{option})
		n.writeMu.Lock()
		_, _ = n.port.Write(raw)
		n.writeMu.Unlock()
	}
	time.Sleep(100 * time.Millisecond) // let the NCP process before we close the port
	n.logger.Info("NCP "+optName+" sent, waiting for USB reconnect...")

	// Stop the read loop and close the port — NCP will disconnect from USB.
	// Close port first to unblock readLoop's blocking serial read, then wait.
	n.lifecycleMu.Lock()
	n.closeOnce.Do(func() { close(n.done) })
	n.port.Close()
	n.lifecycleMu.Unlock()
	n.wg.Wait()

	// nRF52840 USB re-enumerates after reset (may do 2 cycles on factory reset).
	// Retry opening the port and verifying NCP responds.
	for attempt := 1; attempt <= 30; attempt++ {
		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}

		port, err := serial.Open(n.portName, n.portMode)
		if err != nil {
			n.logger.Debug("waiting for NCP USB", "attempt", attempt, "err", err)
			continue
		}
		_ = port.SetDTR(true)
		_ = port.SetRTS(true)

		// Reset internal state and restart read loop with new port.
		n.resetState(port)

		// Verify NCP is alive with a quick GetModuleVersion.
		probeCtx, probeCancel := context.WithTimeout(ctx, 3*time.Second)
		_, err = n.request(probeCtx, zbossCmdGetModuleVersion, nil)
		probeCancel()
		if err == nil {
			n.logger.Info("NCP reconnected after "+optName, "attempts", attempt)
			// Wait for NCPResetInd — signals ZBOSS stack is fully initialized.
			// Without this, NwkFormation may fail with NO_MATCH.
			select {
			case <-n.resetIndCh:
				n.logger.Info("NCPResetInd confirmed, NCP fully ready")
			case <-time.After(3 * time.Second):
				n.logger.Warn("NCPResetInd not received, proceeding anyway")
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		}

		// NCP opened but not responding (probably mid reboot cycle).
		n.logger.Debug("NCP not ready yet, retrying", "attempt", attempt, "err", err)
		n.lifecycleMu.Lock()
		n.closeOnce.Do(func() { close(n.done) })
		port.Close()
		n.lifecycleMu.Unlock()
		n.wg.Wait()
	}

	return fmt.Errorf("NCP did not recover after %s", optName)
}

func (n *NRF52840NCP) Reset(ctx context.Context) error {
	return n.resetAndReconnect(ctx, zbossResetNoOption)
}

func (n *NRF52840NCP) FactoryReset(ctx context.Context) error {
	return n.resetAndReconnect(ctx, zbossResetFactory)
}

// resetState reinitializes internal state with a new serial port.
// Caller must ensure the previous readLoop has exited (wg.Wait) before calling.
func (n *NRF52840NCP) resetState(port serial.Port) {
	n.lifecycleMu.Lock()
	n.port = port
	n.reader = bufio.NewReader(port)
	n.done = make(chan struct{})
	n.llAckCh = make(chan uint8, 4)
	n.resetIndCh = make(chan struct{}, 1)
	n.closeOnce = sync.Once{}
	n.lifecycleMu.Unlock()

	// Drain and close old pending channels to unblock any waiting goroutines.
	n.hlMu.Lock()
	for tsn, ch := range n.hlPending {
		close(ch)
		delete(n.hlPending, tsn)
	}
	n.hlPending = make(map[uint8]chan *zbossFrame)
	n.hlMu.Unlock()

	n.zclMu.Lock()
	for seq, ch := range n.zclPending {
		close(ch)
		delete(n.zclPending, seq)
	}
	n.zclPending = make(map[uint8]chan []byte)
	n.zclMu.Unlock()

	n.llSeqMu.Lock()
	n.llPktSeq = 0
	n.llSeqMu.Unlock()
	n.hlTSN.Store(0)
	n.zclSeq.Store(0)

	n.wg.Add(1)
	go n.readLoop()
}

func (n *NRF52840NCP) Init(ctx context.Context) error {
	resp, err := n.request(ctx, zbossCmdGetModuleVersion, nil)
	if err != nil {
		return err
	}
	if len(resp.Payload) >= 12 {
		fw := binary.LittleEndian.Uint32(resp.Payload[0:4])
		stack := binary.LittleEndian.Uint32(resp.Payload[4:8])
		proto := binary.LittleEndian.Uint32(resp.Payload[8:12])
		stackStr := fmt.Sprintf("%d.%d.%d.%d", (stack>>24)&0xFF, (stack>>16)&0xFF, (stack>>8)&0xFF, stack&0xFF)
		n.ncpInfo = NCPInfo{
			FWVersion:       fw,
			StackVersion:    stackStr,
			ProtocolVersion: proto,
		}
		n.logger.Info("NCP module version", "fw", fw, "stack", stackStr, "protocol", proto)
	}

	// Set Trust Center policies for legacy support (before form or resume).
	// Legacy security: well-known ZigBeeAlliance09 key, no install codes.
	// APS Insecure Join=false means TC will distribute the network key
	// encrypted with the well-known link key (standard TC key exchange).
	tcPolicies := []struct {
		typ  uint16
		val  uint8
		name string
	}{
		{zbossTCPolicyLinkKeysRequired, 0, "TC link keys required=false"},
		{zbossTCPolicyICRequired, 0, "IC required=false"},
		{zbossTCPolicyTCRejoinEnabled, 1, "TC rejoin enabled=true"},
		{zbossTCPolicyIgnoreTCRejoin, 0, "ignore TC rejoin=false"},
		{zbossTCPolicyAPSInsecureJoin, 0, "APS insecure join=false"},
		{zbossTCPolicyDisableNwkMgmtChanUpd, 0, "disable mgmt chan update=false"},
	}
	for _, p := range tcPolicies {
		if err := n.setTCPolicy(ctx, p.typ, p.val); err != nil {
			return fmt.Errorf("set TC policy %s: %w", p.name, err)
		}
	}

	return nil
}

// setTCPolicy sets a single Trust Center policy value.
// Payload: policy_type(2 LE) + value(1).
func (n *NRF52840NCP) setTCPolicy(ctx context.Context, policyType uint16, value uint8) error {
	buf := make([]byte, 3)
	binary.LittleEndian.PutUint16(buf[0:2], policyType)
	buf[2] = value
	_, err := n.request(ctx, zbossCmdSetTCPolicy, buf)
	return err
}

func (n *NRF52840NCP) FormNetwork(ctx context.Context, cfg NetworkConfig) error {
	// Sequence matches zigpy-zboss write_network_info() exactly.

	// 1. Set coordinator role.
	if _, err := n.request(ctx, zbossCmdSetZigbeeRole, []byte{zbossRoleCoordinator}); err != nil {
		return fmt.Errorf("set role: %w", err)
	}

	// 2. Set extended PAN ID (before channel mask, matching zigpy-zboss order).
	if _, err := n.request(ctx, zbossCmdSetExtPanID, cfg.ExtPanID[:]); err != nil {
		return fmt.Errorf("set ext pan id: %w", err)
	}

	// 3. Set channel mask: page(1) + mask(4).
	chanBuf := make([]byte, 5)
	chanBuf[0] = 0x00 // channel page 0 (2.4 GHz)
	binary.LittleEndian.PutUint32(chanBuf[1:], 1<<uint(cfg.Channel))
	if _, err := n.request(ctx, zbossCmdSetChannelMask, chanBuf); err != nil {
		return fmt.Errorf("set channel mask: %w", err)
	}

	// 4. Generate and set a random network key.
	nwkKey := make([]byte, 17) // key(16) + key_seq_num(1)
	if _, err := rand.Read(nwkKey[:16]); err != nil {
		return fmt.Errorf("generate nwk key: %w", err)
	}
	nwkKey[16] = 0x00 // key sequence number
	if _, err := n.request(ctx, zbossCmdSetNwkKey, nwkKey); err != nil {
		return fmt.Errorf("set nwk key: %w", err)
	}
	n.ncpInfo.NetworkKey = make([]byte, 16)
	copy(n.ncpInfo.NetworkKey, nwkKey[:16])
	n.logger.Info("network key set")

	// 5. Form network: channelList(1+5) + scanDuration(1) + distNetFlag(1) + distNetAddr(2) + extPanId(8)
	// Note: extPanId is not in the DSR spec PDF but IS in zigpy-zboss reference and required by NCP firmware.
	formBuf := make([]byte, 18)
	formBuf[0] = 0x01                                                     // 1 channel entry
	formBuf[1] = 0x00                                                     // page 0
	binary.LittleEndian.PutUint32(formBuf[2:6], 1<<uint(cfg.Channel))     // channel mask
	formBuf[6] = 0x05                                                     // scan duration
	formBuf[7] = 0x00                                                     // centralized network (ZC)
	binary.LittleEndian.PutUint16(formBuf[8:10], 0x0000)                  // distributed net addr (unused)
	copy(formBuf[10:18], cfg.ExtPanID[:])
	// NwkFormation may fail transiently after factory reset (NO_MATCH) while
	// the NCP MAC layer finishes initialization. Retry with delay.
	var formErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if _, formErr = n.request(ctx, zbossCmdNwkFormation, formBuf); formErr == nil {
			break
		}
		n.logger.Warn("NwkFormation failed, retrying", "attempt", attempt, "err", formErr)
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if formErr != nil {
		return fmt.Errorf("form network: %w", formErr)
	}

	// 6. Set PAN ID AFTER formation (required by ZBOSS NCP — see zigpy-zboss comment).
	panBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(panBuf, cfg.PanID)
	if _, err := n.request(ctx, zbossCmdSetPanID, panBuf); err != nil {
		return fmt.Errorf("set pan id: %w", err)
	}

	// 7. Set RxOnWhenIdle=true so coordinator listens for incoming frames.
	if _, err := n.request(ctx, zbossCmdSetRxOnWhenIdle, []byte{0x01}); err != nil {
		return fmt.Errorf("set rx on when idle: %w", err)
	}

	// 8. Set end device timeout (256 minutes, matching zigpy-zboss default).
	if _, err := n.request(ctx, zbossCmdSetEDTimeout, []byte{0x08}); err != nil {
		n.logger.Warn("set ED timeout", "err", err)
	}

	// 9. Set max children.
	if _, err := n.request(ctx, zbossCmdSetMaxChildren, []byte{100}); err != nil {
		n.logger.Warn("set max children", "err", err)
	}

	// 10. Wait for PAN ID to persist (zigpy-zboss does sleep(1) here).
	time.Sleep(1 * time.Second)

	return nil
}

func (n *NRF52840NCP) StartNetwork(ctx context.Context) error {
	if _, err := n.request(ctx, zbossCmdNwkStartWithoutForm, nil); err != nil {
		return err
	}

	// Register endpoint 1 with HA profile (after start, matching zigpy-zboss order).
	epDesc := buildSimpleDescPayload(1, zclProfileHA, 0x0005, 0, nil, nil)
	if _, err := n.request(ctx, zbossCmdAFSetSimpleDesc, epDesc); err != nil {
		return fmt.Errorf("register EP1: %w", err)
	}

	return nil
}

func (n *NRF52840NCP) PermitJoin(ctx context.Context, duration uint8) error {
	// ZDO_PERMIT_JOINING_REQ: dest_short(2) + duration(1) + tc_significance(1)
	buf := []byte{0x00, 0x00, duration, 0x01}
	_, err := n.request(ctx, zbossCmdZDOPermitJoiningReq, buf)
	return err
}

func (n *NRF52840NCP) MgmtLeave(ctx context.Context, shortAddr uint16, ieeeAddr [8]byte) error {
	// ZDO_MGMT_LEAVE_REQ: dest_short(2) + ieee(8) + flags(1)
	// flags=0x00: leave permanently, no rejoin
	buf := make([]byte, 11)
	binary.LittleEndian.PutUint16(buf[0:2], shortAddr)
	copy(buf[2:10], ieeeAddr[:])
	buf[10] = 0x00
	_, err := n.request(ctx, zbossCmdZDOMgmtLeaveReq, buf)
	return err
}

func (n *NRF52840NCP) NetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	info := &NetworkInfo{}
	var lastErr error

	resp, err := n.request(ctx, zbossCmdGetChannel, nil)
	if err == nil && len(resp.Payload) >= 2 {
		// Response: channel_page(1) + channel(1)
		info.Channel = resp.Payload[1]
	} else if err != nil {
		lastErr = err
	}

	resp, err = n.request(ctx, zbossCmdGetPanID, nil)
	if err == nil && len(resp.Payload) >= 2 {
		info.PanID = binary.LittleEndian.Uint16(resp.Payload)
	} else if err != nil {
		lastErr = err
	}

	resp, err = n.request(ctx, zbossCmdGetExtPanID, nil)
	if err == nil && len(resp.Payload) >= 8 {
		copy(info.ExtPanID[:], resp.Payload[:8])
	} else if err != nil {
		lastErr = err
	}

	if info.Channel == 0 && info.PanID == 0 && lastErr != nil {
		return nil, fmt.Errorf("network info: all queries failed: %w", lastErr)
	}
	return info, nil
}

func (n *NRF52840NCP) NetworkScan(ctx context.Context) ([]NetworkScanResult, error) {
	// NWK_DISCOVERY (0x0402): blocking call, passive beacon scan.
	// Request: channel_list_len(1) + [page(1) + mask(4)] + scan_duration(1)
	// Scan all 2.4GHz channels (11-26): mask = 0x07FFF800
	buf := make([]byte, 7)
	buf[0] = 0x01                                                 // 1 channel list entry
	buf[1] = 0x00                                                 // page 0 (2.4 GHz)
	binary.LittleEndian.PutUint32(buf[2:6], 0x07FFF800)           // channels 11-26
	buf[6] = 0x05                                                 // scan duration (2^5 + 1 superframes = ~500ms per channel, ~8s total)

	// NWK_DISCOVERY is a blocking call with ~5s timeout on the NCP side.
	// Use a longer context timeout to account for scanning all channels.
	scanCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	resp, err := n.request(scanCtx, zbossCmdNwkDiscovery, buf)
	if err != nil {
		// MAC/NO_BEACON (cat=2, code=0xEA) means scan completed but no networks found.
		if resp != nil && resp.HL.StatusCat == zbossStatusMAC && resp.HL.StatusCode == 0xEA {
			n.logger.Info("network scan complete", "networks_found", 0)
			return nil, nil
		}
		return nil, fmt.Errorf("network scan: %w", err)
	}

	// Response: network_count(1) + network_descriptors[count * 16]
	// Each descriptor: ext_pan_id(8) + pan_id(2) + nwk_update_id(1) + channel_page(1) + channel(1) + flags(1) + lqi(1) + rssi(1)
	// Note: DSR spec response table says "count * 14" but field sizes sum to 16. Using 16.
	if len(resp.Payload) < 1 {
		return nil, nil
	}
	count := int(resp.Payload[0])
	const descSize = 16
	results := make([]NetworkScanResult, 0, count)
	for i := 0; i < count; i++ {
		off := 1 + i*descSize
		if off+descSize > len(resp.Payload) {
			break
		}
		d := resp.Payload[off : off+descSize]
		r := NetworkScanResult{
			PanID:    binary.LittleEndian.Uint16(d[8:10]),
			UpdateID: d[10],
			Channel:  d[12],
			LQI:      d[14],
			RSSI:     int8(d[15]),
		}
		copy(r.ExtPanID[:], d[0:8])
		flags := d[13]
		r.PermitJoin = flags&0x01 != 0
		r.RouterCap = flags&0x02 != 0
		r.EDCap = flags&0x04 != 0
		r.StackProfile = (flags >> 4) & 0x0F
		results = append(results, r)
	}

	n.logger.Info("network scan complete", "networks_found", len(results))
	return results, nil
}

func (n *NRF52840NCP) GetLocalIEEE(ctx context.Context) ([8]byte, error) {
	var ieee [8]byte
	// Request: mac_interface_num(1) = 0
	resp, err := n.request(ctx, zbossCmdGetLocalIEEE, []byte{0x00})
	if err != nil {
		return ieee, fmt.Errorf("get local ieee: %w", err)
	}
	// Response: mac_interface_num(1) + ieee(8)
	if len(resp.Payload) >= 9 {
		copy(ieee[:], resp.Payload[1:9])
	}
	return ieee, nil
}

// --- NCP interface: ZDO ---

func (n *NRF52840NCP) ActiveEndpoints(ctx context.Context, shortAddr uint16) ([]uint8, error) {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, shortAddr)
	resp, err := n.request(ctx, zbossCmdZDOActiveEPReq, buf)
	if err != nil {
		return nil, err
	}
	// ZBOSS response payload: ep_count(1) + ep_list[count] + nwk_addr(2)
	if len(resp.Payload) < 1 {
		return nil, fmt.Errorf("zboss: active EP response empty")
	}
	count := int(resp.Payload[0])
	if len(resp.Payload) < 1+count {
		return nil, fmt.Errorf("zboss: active EP payload truncated: need %d, have %d", 1+count, len(resp.Payload))
	}
	eps := make([]uint8, count)
	copy(eps, resp.Payload[1:1+count])
	n.logger.Info("active endpoints", "short", fmt.Sprintf("0x%04X", shortAddr), "endpoints", eps)
	return eps, nil
}

func (n *NRF52840NCP) SimpleDescriptor(ctx context.Context, shortAddr uint16, endpoint uint8) (*SimpleDescriptor, error) {
	buf := make([]byte, 3)
	binary.LittleEndian.PutUint16(buf, shortAddr)
	buf[2] = endpoint
	resp, err := n.request(ctx, zbossCmdZDOSimpleDescReq, buf)
	if err != nil {
		return nil, err
	}
	// ZBOSS response payload: ep(1) + profile(2) + device_type(2) + device_version(1) +
	//   in_count(1) + out_count(1) + in_clusters[in_count*2] + out_clusters[out_count*2] + nwk_addr(2)
	if len(resp.Payload) < 8 {
		return nil, fmt.Errorf("zboss: simple desc response too short: %d bytes", len(resp.Payload))
	}
	sd := &SimpleDescriptor{
		Endpoint:  resp.Payload[0],
		ProfileID: binary.LittleEndian.Uint16(resp.Payload[1:3]),
		DeviceID:  binary.LittleEndian.Uint16(resp.Payload[3:5]),
	}
	// Payload[5] = device_version (skip)
	inCount := int(resp.Payload[6])
	outCount := int(resp.Payload[7])
	pos := 8
	for i := 0; i < inCount && pos+2 <= len(resp.Payload); i++ {
		sd.InClusters = append(sd.InClusters, binary.LittleEndian.Uint16(resp.Payload[pos:pos+2]))
		pos += 2
	}
	for i := 0; i < outCount && pos+2 <= len(resp.Payload); i++ {
		sd.OutClusters = append(sd.OutClusters, binary.LittleEndian.Uint16(resp.Payload[pos:pos+2]))
		pos += 2
	}
	n.logger.Info("simple descriptor",
		"short", fmt.Sprintf("0x%04X", shortAddr),
		"ep", sd.Endpoint,
		"profile", fmt.Sprintf("0x%04X", sd.ProfileID),
		"device", fmt.Sprintf("0x%04X", sd.DeviceID),
		"in", fmt.Sprintf("%v", sd.InClusters),
		"out", fmt.Sprintf("%v", sd.OutClusters))
	return sd, nil
}

func (n *NRF52840NCP) Bind(ctx context.Context, req BindRequest) error {
	// ZDO_BIND_REQ: nwk_addr(2) + src_ieee(8) + src_ep(1) + cluster(2) + dst_addr_mode(1) + dst_addr(8) + dst_ep(1)
	buf := make([]byte, 23)
	binary.LittleEndian.PutUint16(buf[0:2], req.TargetShortAddr)
	copy(buf[2:10], req.SrcIEEE[:])
	buf[10] = req.SrcEP
	binary.LittleEndian.PutUint16(buf[11:13], req.ClusterID)
	buf[13] = zbossAddrModeIEEE
	copy(buf[14:22], req.DstIEEE[:])
	buf[22] = req.DstEP
	_, err := n.request(ctx, zbossCmdZDOBindReq, buf)
	return err
}

func (n *NRF52840NCP) Unbind(ctx context.Context, req BindRequest) error {
	buf := make([]byte, 23)
	binary.LittleEndian.PutUint16(buf[0:2], req.TargetShortAddr)
	copy(buf[2:10], req.SrcIEEE[:])
	buf[10] = req.SrcEP
	binary.LittleEndian.PutUint16(buf[11:13], req.ClusterID)
	buf[13] = zbossAddrModeIEEE
	copy(buf[14:22], req.DstIEEE[:])
	buf[22] = req.DstEP
	_, err := n.request(ctx, zbossCmdZDOUnbindReq, buf)
	return err
}

// --- NCP interface: ZCL (all via APSDE_DATA_REQ) ---

func (n *NRF52840NCP) ReadAttributes(ctx context.Context, req ReadAttributesRequest) ([]AttributeResponse, error) {
	n.logger.Info("ZCL read attrs TX",
		"short", fmt.Sprintf("0x%04X", req.DstAddr),
		"ep", req.DstEP,
		"cluster", fmt.Sprintf("0x%04X", req.ClusterID),
		"attrs", fmt.Sprintf("%v", req.AttrIDs))

	seq := n.nextZCLSeq()
	zclFrame := zclBuildReadAttributes(seq, req.AttrIDs)
	apsPayload := buildAPSDEDataReq(req.DstAddr, req.DstEP, 1, req.ClusterID, zclProfileHA, 30, zclFrame)

	// Register a channel to receive the ZCL Read Attributes Response.
	ch := make(chan []byte, 1)
	n.zclMu.Lock()
	n.zclPending[seq] = ch
	n.zclMu.Unlock()
	defer func() {
		n.zclMu.Lock()
		delete(n.zclPending, seq)
		n.zclMu.Unlock()
	}()

	// Send the APSDE_DATA_REQ (this confirms transmission, not the ZCL response).
	if _, err := n.request(ctx, zbossCmdAPSDEDataReq, apsPayload); err != nil {
		return nil, err
	}

	// Wait for the ZCL Read Attributes Response arriving via APSDE_DATA_IND.
	select {
	case data := <-ch:
		results := parseAttributeResponses(data)
		for _, r := range results {
			n.logger.Info("ZCL read attrs RX",
				"short", fmt.Sprintf("0x%04X", req.DstAddr),
				"cluster", fmt.Sprintf("0x%04X", req.ClusterID),
				"attr", fmt.Sprintf("0x%04X", r.AttrID),
				"status", r.Status,
				"type", fmt.Sprintf("0x%02X", r.DataType),
				"value", fmt.Sprintf("%X", r.Value))
		}
		return results, nil
	case <-ctx.Done():
		n.logger.Warn("ZCL read attrs timeout",
			"short", fmt.Sprintf("0x%04X", req.DstAddr),
			"cluster", fmt.Sprintf("0x%04X", req.ClusterID))
		return nil, ctx.Err()
	case <-n.done:
		return nil, fmt.Errorf("ncp closed")
	}
}

func (n *NRF52840NCP) WriteAttributes(ctx context.Context, req WriteAttributesRequest) error {
	zclFrame := zclBuildWriteAttributes(n.nextZCLSeq(), req.Records)
	apsPayload := buildAPSDEDataReq(req.DstAddr, req.DstEP, 1, req.ClusterID, zclProfileHA, 30, zclFrame)
	_, err := n.request(ctx, zbossCmdAPSDEDataReq, apsPayload)
	return err
}

func (n *NRF52840NCP) SendCommand(ctx context.Context, req ClusterCommandRequest) error {
	zclFrame := zclBuildClusterCommand(n.nextZCLSeq(), req.CommandID, req.Payload)
	apsPayload := buildAPSDEDataReq(req.DstAddr, req.DstEP, 1, req.ClusterID, zclProfileHA, 30, zclFrame)
	_, err := n.request(ctx, zbossCmdAPSDEDataReq, apsPayload)
	return err
}

func (n *NRF52840NCP) ConfigureReporting(ctx context.Context, req ConfigureReportingRequest) error {
	zclFrame := zclBuildConfigureReporting(n.nextZCLSeq(), req.AttrID, req.DataType, req.MinInterval, req.MaxInterval, req.ReportChange)
	apsPayload := buildAPSDEDataReq(req.DstAddr, req.DstEP, 1, req.ClusterID, zclProfileHA, 30, zclFrame)
	_, err := n.request(ctx, zbossCmdAPSDEDataReq, apsPayload)
	return err
}

// --- Indication callback setters ---

func (n *NRF52840NCP) OnDeviceJoined(handler func(DeviceJoinedEvent)) {
	n.handlerMu.Lock()
	defer n.handlerMu.Unlock()
	n.onJoined = handler
}
func (n *NRF52840NCP) OnDeviceLeft(handler func(DeviceLeftEvent)) {
	n.handlerMu.Lock()
	defer n.handlerMu.Unlock()
	n.onLeft = handler
}
func (n *NRF52840NCP) OnDeviceAnnounce(handler func(DeviceAnnounceEvent)) {
	n.handlerMu.Lock()
	defer n.handlerMu.Unlock()
	n.onAnnounce = handler
}
func (n *NRF52840NCP) OnAttributeReport(handler func(AttributeReportEvent)) {
	n.handlerMu.Lock()
	defer n.handlerMu.Unlock()
	n.onReport = handler
}
func (n *NRF52840NCP) OnClusterCommand(handler func(ClusterCommandEvent)) {
	n.handlerMu.Lock()
	defer n.handlerMu.Unlock()
	n.onClusterCmd = handler
}

func (n *NRF52840NCP) OnNwkAddrUpdate(handler func(uint16)) {
	n.handlerMu.Lock()
	defer n.handlerMu.Unlock()
	n.onNwkAddrUpdate = handler
}

// OnNCPReset registers a callback for spontaneous NCP reset events.
func (n *NRF52840NCP) OnNCPReset(handler func()) {
	n.handlerMu.Lock()
	defer n.handlerMu.Unlock()
	n.onReset = handler
}

// GetNCPInfo returns a copy of cached firmware/stack/protocol version information.
func (n *NRF52840NCP) GetNCPInfo() *NCPInfo {
	info := n.ncpInfo
	if n.ncpInfo.NetworkKey != nil {
		info.NetworkKey = make([]byte, len(n.ncpInfo.NetworkKey))
		copy(info.NetworkKey, n.ncpInfo.NetworkKey)
	}
	return &info
}

// Close stops the NCP and waits for readLoop to exit.
func (n *NRF52840NCP) Close() error {
	n.lifecycleMu.Lock()
	if n.closed {
		n.lifecycleMu.Unlock()
		return nil
	}
	n.closed = true
	n.closeOnce.Do(func() { close(n.done) })
	err := n.port.Close()
	n.lifecycleMu.Unlock()

	n.wg.Wait()

	n.hlMu.Lock()
	for tsn, ch := range n.hlPending {
		close(ch)
		delete(n.hlPending, tsn)
	}
	n.hlMu.Unlock()

	n.zclMu.Lock()
	for seq, ch := range n.zclPending {
		close(ch)
		delete(n.zclPending, seq)
	}
	n.zclMu.Unlock()

	return err
}

// --- Helpers ---

// buildSimpleDescPayload builds AF_SET_SIMPLE_DESC payload.
func buildSimpleDescPayload(ep uint8, profileID, deviceID uint16, devVersion uint8, inClusters, outClusters []uint16) []byte {
	buf := make([]byte, 8+len(inClusters)*2+len(outClusters)*2)
	buf[0] = ep
	binary.LittleEndian.PutUint16(buf[1:3], profileID)
	binary.LittleEndian.PutUint16(buf[3:5], deviceID)
	buf[5] = devVersion
	buf[6] = uint8(len(inClusters))
	buf[7] = uint8(len(outClusters))
	pos := 8
	for _, c := range inClusters {
		binary.LittleEndian.PutUint16(buf[pos:pos+2], c)
		pos += 2
	}
	for _, c := range outClusters {
		binary.LittleEndian.PutUint16(buf[pos:pos+2], c)
		pos += 2
	}
	return buf
}
