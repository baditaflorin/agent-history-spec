package envelope

// Hard limits of the wire format. A producer cuts to these before sending; the store rejects
// an event that exceeds them rather than cutting it itself.
const (
	MaxContentBytes = 32 << 10 // content and command
	MaxRawBytes     = 1 << 20  // raw
	MaxLineBytes    = 2 << 20  // one NDJSON line, which holds one event with its raw
	MaxEdges        = 64
	MaxIDBytes      = 256
	MaxFieldBytes   = 1 << 10 // host, agent, session_key, tool_name, cwd, git_*, title, resume_cmd, edge dst
)
