package envelope

// SessionKey builds "agent:host:native_session_id". Agent and host never contain ':' (Validate
// enforces it); the native id may, because the key is only ever split on its first two colons.
func SessionKey(agent, host, nativeID string) string {
	return agent + ":" + host + ":" + nativeID
}
