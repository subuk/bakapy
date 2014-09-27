package bakapy

// Waiting for client authentication
const STORAGE_AUTH_TIMEOUT = 30 // seconds
const STORAGE_TASK_ID_LEN = 36

// Length of filename length header
const STORAGE_FILENAME_LEN_LEN = 4

// Storage connection states
const (
	STATE_WAIT_TASK_ID  = iota
	STATE_WAIT_FILENAME = iota
	STATE_WAIT_DATA     = iota
	STATE_RECEIVING     = iota
	STATE_END           = iota
)
