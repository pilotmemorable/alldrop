package engine

import "os"

func osWrite(p string, b []byte) error { return os.WriteFile(p, b, 0o644) }
func osRead(p string) ([]byte, error)  { return os.ReadFile(p) }
