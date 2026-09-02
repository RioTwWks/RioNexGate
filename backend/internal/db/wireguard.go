package db

import ("errors"; "fmt"; "net"; "rionexgate/internal/models")

func (d *DB) GetWireGuardPeerByUserID(userID uint) (*models.WireGuardPeer, error) {
	var peer models.WireGuardPeer
	if err := d.Where("user_id = ?", userID).First(&peer).Error; err != nil { return nil, err }
	return &peer, nil
}
func (d *DB) ListWireGuardPeersForUsers(userIDs []uint) ([]models.WireGuardPeer, error) {
	if len(userIDs) == 0 { return nil, nil }
	var peers []models.WireGuardPeer
	return peers, d.Where("user_id IN ?", userIDs).Find(&peers).Error
}
func (d *DB) EnsureWireGuardPeer(userID uint, subnet string) (*models.WireGuardPeer, error) {
	if peer, err := d.GetWireGuardPeerByUserID(userID); err == nil { return peer, nil }
	priv, pub, err := generateWireGuardKeypair(); if err != nil { return nil, err }
	psk, err := generateWireGuardPSK(); if err != nil { return nil, err }
	ip, err := d.allocateClientIP(subnet, userID); if err != nil { return nil, err }
	peer := &models.WireGuardPeer{UserID: userID, PrivateKey: priv, PublicKey: pub, PresharedKey: psk, ClientIP: ip}
	return peer, d.Create(peer).Error
}
func (d *DB) allocateClientIP(subnet string, userID uint) (string, error) {
	if subnet == "" { subnet = "10.8.0.0/24" }
	_, ipNet, err := net.ParseCIDR(subnet); if err != nil { return "", err }
	base := ipNet.IP.To4(); if base == nil { return "", errors.New("awg subnet must be IPv4") }
	var peers []models.WireGuardPeer; _ = d.Find(&peers).Error
	used := map[string]struct{}{}; for _, p := range peers { used[p.ClientIP] = struct{}{} }
	for host := 2; host < 255; host++ {
		ip := net.IPv4(base[0], base[1], base[2], byte(host))
		if !ipNet.Contains(ip) { continue }
		cidr := fmt.Sprintf("%s/32", ip.String())
		if _, ok := used[cidr]; !ok { return cidr, nil }
	}
	return fmt.Sprintf("%s/32", net.IPv4(base[0], base[1], base[2], byte((userID%250)+2)).String()), nil
}
func (d *DB) DeleteWireGuardPeerByUserID(userID uint) error {
	return d.Where("user_id = ?", userID).Delete(&models.WireGuardPeer{}).Error
}
