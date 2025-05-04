package status

type Status int

const (
	Finished Status = iota
	Error
	Pending
)

func (l Status) String() string {
	switch l {
	case Pending:
		return "Pending"
	case Error:
		return "Error"
	case Finished:
		return "Finished"
	default:
		return "Unknown status"
	}
}
