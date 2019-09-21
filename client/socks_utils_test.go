package main

/*
type myReadWriter struct {
	buf []byte
}

func (mrw *myReadWriter) Read(p []byte) (int, error) {
	if len(p) >= cap(mrw.buf) {
		copy(mrw.buf, p)
		return cap(mrw.buf), nil
	}
	copy(mrw.buf, p)
	return len()
}
func (mrw *myReadWriter) Write(p []byte) (int, error) {
	if len(mrw.buf) >= cap(p) {
		copy(p, mrw.buf[:cap(p)])
		return cap(p), nil
	}
	copy(p[:len(buf)], mrw.buf)
	return len(buf), nil
}

func TestSocks5_handshake(t *testing.T) {
	socks4request := []byte{4, 1, 0}
	socks5_handshake()
}
*/
