package raknet

import (
  "log"
  "sync"
  "time"
)

type SplitPacketHandler struct {
  splitPackets map[int16]splitPacketComposition
  splitPacketLock sync.Mutex
}

type splitPacketComposition struct {
  packets []*EncapsulatedPacketPart
  expiration time.Time
}

func newSplitPacketComposition(pkts int) splitPacketComposition {
  return splitPacketComposition{
    expiration: time.Now().Add(1 * time.Second),
    packets: make([]*EncapsulatedPacketPart, pkts),
  }
}

func NewSplitPacketHandler() SplitPacketHandler {
  return SplitPacketHandler{
    splitPackets: make(map[int16]splitPacketComposition),
  }
}

func (sph *SplitPacketHandler) GarbageCollect() {
  // This operation requires exclusive access.
  sph.splitPacketLock.Lock()
  defer sph.splitPacketLock.Unlock()

  now := time.Now()

  for k, v := range sph.splitPackets {
    if now.After(v.expiration) {
      delete(sph.splitPackets, k)
    }
  }
}

func (sph *SplitPacketHandler) AcceptSplitPacket(pkt EncapsulatedPacketPart) (*[]*EncapsulatedPacketPart) {
  // Multiple goroutines could be accepting split packets. Lock the mutex.
  sph.splitPacketLock.Lock()
  defer sph.splitPacketLock.Unlock()

  allSplit, ok := sph.splitPackets[pkt.PartId]

  if !ok {
    // Lightly verify that this packet is sane
    if pkt.PartIndex >= pkt.PartCount {
      log.Printf("Got a split datagram with an unacceptably large part index (%d >= %d). Ignoring.",
        pkt.PartIndex, pkt.PartCount)
      return nil
    }
    allSplit = newSplitPacketComposition(int(pkt.PartCount))
    sph.splitPackets[pkt.PartId] = allSplit
  }

  // Lightly verify that this packet is sane
  if int(pkt.PartIndex) >= len(allSplit.packets) {
    log.Printf("Got a split datagram with an unacceptably large part index (%d >= %d). Ignoring.",
      pkt.PartIndex, len(allSplit.packets))
    return nil
  }

  // Save this packet
  allSplit.packets[pkt.PartIndex] = &pkt

  // Do we have all the parts for this packet?
  for _, other := range allSplit.packets {
    if other == nil {
      return nil
    }
  }

  delete(sph.splitPackets, pkt.PartId)
  return &allSplit.packets
}
