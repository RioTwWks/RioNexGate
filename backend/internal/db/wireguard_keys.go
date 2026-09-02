package db

import ("crypto/ecdh"; "crypto/rand"; "encoding/base64")

func generateWireGuardKeypair() (string, string, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil { return "", "", err }
	return base64.StdEncoding.EncodeToString(priv.Bytes()), base64.StdEncoding.EncodeToString(priv.PublicKey().Bytes()), nil
}
func generateWireGuardPSK() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil { return "", err }
	return base64.StdEncoding.EncodeToString(buf), nil
}
