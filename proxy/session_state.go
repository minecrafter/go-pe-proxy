package proxy

type sessionState int

const (
  // This is the initial state that all new known connections are placed into.
  // We will await for the client to identify the username, UUID, etc. before
  // we move to connecting the player.
  STATE_IDENTIFY sessionState = iota
  // Players are moved to this state when we are connecting to our initial server.
  STATE_CONNECTING
  // Players are moved to this state when we have successfully initiated a connection.
  STATE_CONNECTED
)

type sessionConnectorState int

const (
  // This is the initial state that all new known connectors are moved to.
  // They are not yet connected to a server.
  C_STATE_UNCONNECTED sessionConnectorState = iota
  // The connector is connected and is sending the RakNet handshake and MCPE login
  // packets.
  C_STATE_IDENTIFY
  // The connector has identified this player with the server. From here on out,
  // most packets will be simply be tampered with by the proxy, instead of being
  // handled by the proxy.
  C_STATE_CONNECTED
)
