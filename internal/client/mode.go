package client

type mode byte

const (
	None         mode = 0
	InContainer  mode = 1
	LinuxNsenter mode = 2
)

func (m mode) String() (rv string) {
	switch m {
	case InContainer:
		return "in-container"
	case LinuxNsenter:
		return "linux-nsenter"
	case None:
	}

	return
}
