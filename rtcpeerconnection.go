package webrtc

import (
	"fmt"
	"math/rand"

	"github.com/pions/webrtc/internal/dtls"
	"github.com/pions/webrtc/internal/ice"
	"github.com/pions/webrtc/internal/network"
	"github.com/pions/webrtc/internal/sdp"
	"github.com/pions/webrtc/internal/util"
	"github.com/pions/webrtc/pkg/rtp"

	"github.com/pkg/errors"
)

// TrackType determines the type of media we are sending receiving
type TrackType int

// List of supported TrackTypes
const (
	G711 TrackType = iota
	G722
	ILBC
	ISAC
	H264
	VP8
	Opus
)

// RTCPeerConnection represents a WebRTC connection between itself and a remote peer
type RTCPeerConnection struct {
	Ontrack          func(mediaType TrackType, buffers <-chan *rtp.Packet)
	LocalDescription *sdp.SessionDescription

	tlscfg *dtls.TLSCfg

	iceUsername string
	icePassword string
}

// Public

// SetRemoteDescription sets the SessionDescription of the remote peer
func (r *RTCPeerConnection) SetRemoteDescription(string) error {
	return nil
}

// CreateOffer starts the RTCPeerConnection and generates the localDescription
// The order of CreateOffer/SetRemoteDescription determines if we are the offerer or the answerer
// Once the RemoteDescription has been set network activity will start
func (r *RTCPeerConnection) CreateOffer() error {
	if r.tlscfg != nil {
		return errors.Errorf("tlscfg is already defined, CreateOffer can only be called once")
	}
	r.tlscfg = dtls.NewTLSCfg()
	r.iceUsername = util.RandSeq(16)
	r.icePassword = util.RandSeq(32)

	candidates := []string{}
	basePriority := uint16(rand.Uint32() & (1<<16 - 1))
	for id, c := range ice.HostInterfaces() {
		dstPort, err := network.UDPListener(c, []byte(r.icePassword), r.tlscfg, r.generateChannel)
		if err != nil {
			return err
		}
		candidates = append(candidates, fmt.Sprintf("candidate:udpcandidate %d udp %d %s %d typ host", id, basePriority, c, dstPort))
		basePriority = basePriority + 1
	}

	r.LocalDescription = sdp.VP8OnlyDescription(r.iceUsername, r.icePassword, r.tlscfg.Fingerprint(), candidates)

	return nil
}

// AddTrack adds a new track to the RTCPeerConnection
// This function returns a channel to push buffers on, and an error if the channel can't be added
// Closing the channel ends this stream
func (r *RTCPeerConnection) AddTrack(mediaType TrackType) (buffers chan<- []byte, err error) {
	return nil, nil
}

// Private
func (r *RTCPeerConnection) generateChannel(ssrc uint32) (buffers chan<- *rtp.Packet) {
	if r.Ontrack == nil {
		return nil
	}

	bufferTransport := make(chan *rtp.Packet, 15)
	go r.Ontrack(VP8, bufferTransport) // TODO look up media via SSRC in remote SD
	return bufferTransport
}
