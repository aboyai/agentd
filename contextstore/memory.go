// contextstore/memory.go
package contextstore

var memoryDB = make(map[string]string)

func FetchMemory(sessionID string) (string, bool) {
	mem, ok := memoryDB[sessionID]
	return mem, ok
}

func AppendToMemory(sessionID, content string) {
	memoryDB[sessionID] += "\n" + content
}


