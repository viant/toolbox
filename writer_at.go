package toolbox

//ByteWriterAt  represents a bytes writer at
type ByteWriterAt struct {
	Buffer   []byte
	position int
}

//WriteAt returns number of written bytes or error
func (w *ByteWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	if int(offset) == w.position {
		w.Buffer = append(w.Buffer, p...)
		w.position += len(p)
		return len(p), nil
	} else if w.position < int(offset) {
		var diff = (int(offset) - w.position)
		var fillingBytes = make([]byte, diff)
		w.position += len(fillingBytes)
		w.Buffer = append(w.Buffer, fillingBytes...)
		return w.WriteAt(p, offset)
	} else {
		for i := 0; i < len(p); i++ {
			var index = int(offset) + i
			if index < len(w.Buffer) {
				w.Buffer[int(offset)+i] = p[i]
			} else {
				w.Buffer = append(w.Buffer, p[i:]...)
				break
			}
		}
		return len(p), nil
	}
}

//NewWriterAt returns a new instance of byte writer at
func NewByteWriterAt() *ByteWriterAt {
	return &ByteWriterAt{
		Buffer: make([]byte, 0),
	}
}
