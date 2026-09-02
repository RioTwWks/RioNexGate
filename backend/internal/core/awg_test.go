package core

import ("strings"; "testing"; "rionexgate/internal/config"; "rionexgate/internal/models")

func testAWGConfig() *config.StealthAWGConfig {
	return &config.StealthAWGConfig{Enabled: true, Port: 51820, PrivateKey: "SP", PublicKey: "SPUB", Jc: 4, Jmin: 40, Jmax: 70, S1: 84, H1: 1, H2: 2, H3: 3, H4: 4}
}
func testAWGPeer() *models.WireGuardPeer {
	return &models.WireGuardPeer{PrivateKey: "CP", PublicKey: "CPUB", PresharedKey: "PSK", ClientIP: "10.8.0.2/32"}
}
func TestBuildAWGClientConfig(t *testing.T) {
	if !strings.Contains(BuildAWGClientConfig("h.example", testAWGConfig(), testAWGPeer()), "Endpoint = h.example:51820") { t.Fatal() }
}
func TestBuildAWGURILink(t *testing.T) {
	if !strings.HasPrefix(BuildAWGURILink("x"), "awg://") { t.Fatal() }
}
func TestGetClientLinkProfilesWithAWG(t *testing.T) {
	s := testStealthConfig(); s.AWG = *testAWGConfig()
	p := GetClientLinkProfiles("h", 443, models.User{Email: "u"}, s, testAWGPeer())
	if len(p) != 3 || p[2].Transport != "awg" { t.Fatalf("%+v", p) }
}
func TestBuildSubscriptionWithAWG(t *testing.T) {
	s := testStealthConfig(); s.AWG = *testAWGConfig()
	for _, l := range BuildSubscriptionLinks("h", 443, models.User{Email: "u"}, s, nil, testAWGPeer()) {
		if strings.HasPrefix(l, "awg://") { return }
	}
	t.Fatal("no awg line")
}
