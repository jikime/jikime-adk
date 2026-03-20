package team

// Transport is the pluggable interface for delivering messages between agents.
// The default implementation uses the file-based Inbox.
// A ZeroMQ or network implementation can satisfy this interface.
type Transport interface {
	// Send delivers msg to the agent identified by msg.To.
	Send(msg *Message) error

	// Receive consumes up to limit messages from agentID's queue.
	// Pass limit ≤ 0 to receive all pending messages.
	Receive(agentID string, limit int) ([]*Message, error)

	// Peek returns up to limit messages without consuming them.
	Peek(agentID string, limit int) ([]*Message, error)

	// Count returns the number of queued messages for agentID.
	Count(agentID string) (int, error)

	// Broadcast delivers a copy of msg to all agentIDs except msg.From.
	Broadcast(msg *Message, agentIDs []string) error
}

// FileTransport is the default Transport backed by TeamInbox (file system).
type FileTransport struct {
	ti *TeamInbox
}

// NewFileTransport wraps a TeamInbox as a Transport.
func NewFileTransport(teamDir string) *FileTransport {
	return &FileTransport{ti: NewTeamInbox(teamDir)}
}

func (f *FileTransport) Send(msg *Message) error {
	return f.ti.Send(msg)
}

func (f *FileTransport) Receive(agentID string, limit int) ([]*Message, error) {
	ib, err := f.ti.For(agentID)
	if err != nil {
		return nil, err
	}
	return ib.Receive(limit)
}

func (f *FileTransport) Peek(agentID string, limit int) ([]*Message, error) {
	ib, err := f.ti.For(agentID)
	if err != nil {
		return nil, err
	}
	return ib.Peek(limit)
}

func (f *FileTransport) Count(agentID string) (int, error) {
	ib, err := f.ti.For(agentID)
	if err != nil {
		return 0, err
	}
	return ib.Count()
}

func (f *FileTransport) Broadcast(msg *Message, agentIDs []string) error {
	return f.ti.Broadcast(msg, agentIDs)
}
