package util

func CloseChannel(ch chan error) {
	if _, ok := <-ch; ok {
		close(ch)
	}
}
