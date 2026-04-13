package client

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/ao-data/albiondata-client/client/photon"
	"github.com/ao-data/albiondata-client/log"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type listener struct {
	handle        *pcap.Handle
	sourcePackets chan gopacket.Packet
	rawCommands   chan photon.RawPacket
	displayName   string
	parser        *photon.PhotonParser
	quit          chan bool
	router        *Router
}

func newListener(router *Router) *listener {
	l := &listener{
		rawCommands: make(chan photon.RawPacket, 1),
		quit:        make(chan bool, 1),
		router:      router,
	}
	l.parser = photon.NewPhotonParser(l.onRequest, l.onResponse, l.onEvent)
	l.parser.OnEncrypted = l.onEncrypted
	return l
}

func (l *listener) startOnline(device string, port int) {
	handle, err := pcap.OpenLive(device, 2048, false, pcap.BlockForever)
	if err != nil {
		log.Panic(err)
	}
	l.handle = handle

	err = l.handle.SetBPFFilter(fmt.Sprintf("tcp port %d || udp port %d", port, port))
	if err != nil {
		log.Panic(err)
	}

	source := gopacket.NewPacketSource(l.handle, l.handle.LinkType())
	l.sourcePackets = source.Packets()

	l.displayName = fmt.Sprintf("online: %s:%d", device, port)
	l.run()
}

func (l *listener) startOfflinePcap(path string) {
	handle, err := pcap.OpenOffline(path)
	if err != nil {
		log.Panicf("Problem creating offline source. Error: %v", err)
	}
	l.handle = handle

	source := gopacket.NewPacketSource(handle, handle.LinkType())
	l.sourcePackets = source.Packets()

	l.displayName = fmt.Sprintf("Offline Pcap: %s", path)
	l.run()
}

func (l *listener) startOfflineCommandGob(path string) {
	// Set up packets with an empty channel
	l.sourcePackets = make(chan gopacket.Packet, 1)

	var decoder *gob.Decoder
	file, err := os.Open(path)
	if err != nil {
		log.Panic("Could not open commands input file ", err)
	} else {
		decoder = gob.NewDecoder(file)
	}

	go func() {
		for {
			raw := &photon.RawPacket{}
			if decoder == nil {
				break
			}
			err = decoder.Decode(raw)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Error("Could not decode raw packet ", err)
				continue
			}
			l.rawCommands <- *raw
		}

		err = file.Close()
		if err != nil {
			log.Error("Could not close commands input file ", err)
		}
		log.Info("All offline commands should processed now.")
	}()

	l.displayName = fmt.Sprintf("Offline Commands: %s", path)
	l.run()
}

func (l *listener) run() {
	log.Debugf("Starting listener (%s)...", l.displayName)

	for {
		select {
		case <-l.quit:
			log.Debugf("Listener shutting down (%s)...", l.displayName)
			l.handle.Close()
			return
		case packet := <-l.sourcePackets:
			if packet != nil {
				l.processPacket(packet)
			} else {
				// MUST only happen with the offline processor.
				l.handle.Close()
				return
			}
		case raw := <-l.rawCommands:
			l.parser.ReceivePacket(raw.Payload)
		}
	}
}

func (l *listener) stop() {
	l.quit <- true
	l.handle.Close()
}

func (l *listener) processPacket(packet gopacket.Packet) {
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return
	}

	ipv4 := ipLayer.(*layers.IPv4)
	log.Tracef("Packet came from: %s", ipv4.SrcIP)

	if ipv4.SrcIP == nil {
		log.Trace("No IPv4 detected")
		return
	}

	l.router.albionstate.GameServerIP = ipv4.SrcIP.String()
	l.router.albionstate.AODataServerID, l.router.albionstate.AODataIngestBaseURL = l.router.albionstate.GetServer()
	log.Tracef("Server ID: %d", l.router.albionstate.AODataServerID)
	log.Tracef("Using AODataIngestBaseURL: %s", l.router.albionstate.AODataIngestBaseURL)

	// Extract the raw Photon payload from the UDP or TCP layer.
	var payload []byte
	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		payload = udpLayer.(*layers.UDP).Payload
	} else if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		payload = tcpLayer.(*layers.TCP).Payload
	}

	if len(payload) == 0 {
		return
	}

	if ConfigGlobal.RecordPath != "" {
		l.router.recordRawPacket <- photon.RawPacket{Payload: payload}
	}

	l.parser.ReceivePacket(payload)
}

func (l *listener) onEncrypted() {
	if l.router.albionstate.WaitingForMarketData {
		l.router.albionstate.WaitingForMarketData = false
		log.Info("Market data is encrypted. Please see https://www.albion-online-data.com/client/encryption.html for more information.")
	}
}

func (l *listener) onRequest(opCode byte, params map[byte]interface{}) {
	if _, ok := params[253]; !ok {
		params[253] = uint16(opCode)
	}

	operation, err := decodeRequest(params)
	if params[253] != nil {
		if number, ok := toUint16(params[253]); ok {
			shouldDebug, exists := ConfigGlobal.DebugOperations[int(number)]
			if (exists && shouldDebug) || (!exists && ConfigGlobal.DebugOperationsString == "") {
				log.Debugf("OperationRequest: [%v]%v - %s", number, OperationType(number), formatDebugPhotonParams(params))
			}
		} else {
			log.Debugf("OperationRequest: unexpected type for params[253]: %T = %s", params[253], formatDebugValue(params[253], 0))
		}
	} else if !ConfigGlobal.DebugIgnoreDecodingErrors {
		log.Debugf("OperationRequest: ERROR - %s", formatDebugPhotonParams(params))
	}

	l.dispatchOperation(operation, err, params)
}

func (l *listener) onResponse(opCode byte, returnCode int16, _ string, params map[byte]interface{}) {
	if _, ok := toUint16(params[253]); !ok {
		params[253] = uint16(opCode)
	}

	// Market order data arrives as params[0]=[]string (set by dispatchResponse in parser).
	if _, ok := params[0].([]string); ok {
		params[253] = uint16(opAuctionGetOffers)
	}

	if returnCode != 0 {
		rc := uint16(returnCode)
		name := returnCodeName(rc)
		if name == "" {
			name = "Unknown"
		}
		if number, ok := toUint16(params[253]); ok {
			log.Debugf("OperationResponse: rc=%d(%s) op=[%v]%v", returnCode, name, number, OperationType(number))
		} else {
			log.Debugf("OperationResponse: rc=%d(%s)", returnCode, name)
		}
	}

	operation, err := decodeResponse(params)
	if params[253] != nil {
		if number, ok := toUint16(params[253]); ok {
			shouldDebug, exists := ConfigGlobal.DebugOperations[int(number)]
			if (exists && shouldDebug) || (!exists && ConfigGlobal.DebugOperationsString == "") {
				log.Debugf("OperationResponse: [%v]%v - %s", number, OperationType(number), formatDebugPhotonParams(params))
			}
		} else {
			log.Debugf("OperationResponse: unexpected type for params[253]: %T = %s", params[253], formatDebugValue(params[253], 0))
		}
	} else if !ConfigGlobal.DebugIgnoreDecodingErrors {
		log.Debugf("OperationResponse: ERROR - %s", formatDebugPhotonParams(params))
	}

	l.dispatchOperation(operation, err, params)
}

func (l *listener) onEvent(code byte, params map[byte]interface{}) {
	if _, ok := toUint16(params[252]); !ok {
		params[252] = uint16(code)
	}

	operation, err := decodeEvent(params)
	if params[252] != nil {
		if number, ok := toUint16(params[252]); ok {
			shouldDebug, exists := ConfigGlobal.DebugEvents[int(number)]
			if (exists && shouldDebug) || (!exists && ConfigGlobal.DebugEventsString == "") {
				log.Debugf("EventDataType: [%v]%v - %s", number, EventType(number), formatDebugPhotonParams(params))
			}
		} else {
			log.Debugf("EventDataType: unexpected type for params[252]: %T = %s", params[252], formatDebugValue(params[252], 0))
		}
	} else if !ConfigGlobal.DebugIgnoreDecodingErrors {
		log.Debugf("EventDataType: ERROR - %s", formatDebugPhotonParams(params))
	}

	l.dispatchOperation(operation, err, params)
}

func (l *listener) dispatchOperation(op operation, err error, params map[byte]interface{}) {
	if err != nil && !ConfigGlobal.DebugIgnoreDecodingErrors {
		log.Debugf("Error while decoding an event or operation: %v - params: %s", err, formatDebugPhotonParams(params))
		return
	}
	if op != nil {
		l.router.newOperation <- op
	}
}

func normalizeLocationID(v string) string {
	s := strings.TrimSpace(strings.Trim(v, ",."))
	if s == "" {
		return ""
	}
	reIsland := regexp.MustCompile(`(?i)@island@[0-9a-f-]{36}`)
	if m := reIsland.FindString(s); m != "" {
		return "@ISLAND@" + m[len("@island@"):]
	}
	reNumeric := regexp.MustCompile(`^[0-9]{3,6}$`)
	if reNumeric.MatchString(s) {
		return s
	}
	ls := strings.ToLower(s)
	if strings.HasPrefix(ls, "island-player-") ||
		strings.HasPrefix(ls, "@player-island") ||
		strings.HasPrefix(ls, "@island-") ||
		strings.HasPrefix(s, "BLACKBANK-") ||
		strings.HasSuffix(s, "-HellDen") ||
		strings.HasSuffix(s, "-Auction2") {
		return s
	}
	return ""
}

