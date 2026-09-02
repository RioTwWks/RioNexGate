package api
import ("context"; "net"; "testing")
func TestCheckNodeTCPReachable(t *testing.T) {
  ln, err := net.Listen("tcp", "127.0.0.1:0"); if err != nil { t.Fatal(err) }
  defer ln.Close(); _, p, _ := net.SplitHostPort(ln.Addr().String())
  var port int; for _, c := range p { port = port*10 + int(c-'0') }
  if !checkNodeTCP(context.Background(), "127.0.0.1", port).Reachable { t.Fatal("expected reachable") }
}
