package db

import (
	"errors"
	"fmt"
	"sort"

	"rionexgate/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ListUserExits returns active exit assignments for a user, lowest priority first.
func (d *DB) ListUserExits(userID uint) ([]models.UserExit, error) {
	var rows []models.UserExit
	err := d.Where("user_id = ? AND active = ?", userID, true).
		Order("priority asc, id asc").
		Find(&rows).Error
	return rows, err
}

// ListUserExitEmails returns all routing emails for stats aggregation.
func (d *DB) ListUserExitEmails(userID uint) ([]string, error) {
	user, err := d.GetUser(userID)
	if err != nil {
		return nil, err
	}
	emails := []string{user.Email}
	rows, err := d.ListUserExits(userID)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{user.Email: true}
	for _, r := range rows {
		if r.Email == "" || seen[r.Email] {
			continue
		}
		seen[r.Email] = true
		emails = append(emails, r.Email)
	}
	return emails, nil
}

// ResolveUserExitNodes returns exit nodes bound to the user (via user_exits),
// falling back to ExitNodeID / best exit for backward compatibility.
func (d *DB) ResolveUserExitNodes(user *models.User) ([]models.Node, error) {
	rows, err := d.ListUserExits(user.ID)
	if err != nil {
		return nil, err
	}
	var nodes []models.Node
	seen := map[uint]bool{}
	for _, row := range rows {
		node, err := d.GetNode(row.NodeID)
		if err != nil || !node.Active || node.Role != models.NodeRoleExit {
			continue
		}
		if seen[node.ID] {
			continue
		}
		seen[node.ID] = true
		nodes = append(nodes, *node)
	}
	if len(nodes) > 0 {
		return nodes, nil
	}
	// Legacy single exit_node_id / auto best
	node, err := d.ResolveUserExitNode(user)
	if err != nil {
		return nil, err
	}
	return []models.Node{*node}, nil
}

// SetUserExits replaces the user's exit assignments. The first id in exitNodeIDs
// (or lowest priority among them) becomes the primary identity (keeps user UUID/email).
// Empty exitNodeIDs clears assignments and exit_node_id.
func (d *DB) SetUserExits(userID uint, exitNodeIDs []uint) (*models.User, error) {
	user, err := d.GetUser(userID)
	if err != nil {
		return nil, err
	}

	// Deduplicate preserving order
	ordered := make([]uint, 0, len(exitNodeIDs))
	seen := map[uint]bool{}
	for _, id := range exitNodeIDs {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		ordered = append(ordered, id)
	}

	type prepared struct {
		node     models.Node
		priority int
	}
	var exits []prepared
	for i, id := range ordered {
		node, err := d.GetNode(id)
		if err != nil {
			return nil, fmt.Errorf("exit node %d: %w", id, err)
		}
		if node.Role != models.NodeRoleExit {
			return nil, fmt.Errorf("node %d is not an exit", id)
		}
		if !node.Active {
			return nil, fmt.Errorf("exit node %d is inactive", id)
		}
		exits = append(exits, prepared{node: *node, priority: (i + 1) * 100})
	}

	// Preserve UUID/email for exits that remain assigned so client profiles stay stable.
	existing, _ := d.ListUserExits(userID)
	prevByNode := make(map[uint]models.UserExit, len(existing))
	for _, row := range existing {
		prevByNode[row.NodeID] = row
	}

	err = d.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserExit{}).Error; err != nil {
			return err
		}
		if len(exits) == 0 {
			return tx.Model(&models.User{}).Where("id = ?", userID).
				Updates(map[string]interface{}{"exit_node_id": nil}).Error
		}

		for i, ex := range exits {
			primary := i == 0
			slug := models.ExitSlug(ex.node)
			clientUUID := user.UUID
			clientEmail := user.Email
			if !primary {
				if prev, ok := prevByNode[ex.node.ID]; ok && prev.UUID != "" && prev.UUID != user.UUID {
					clientUUID = prev.UUID
					clientEmail = prev.Email
					if clientEmail == "" {
						clientEmail = models.ExitRouteEmail(user.Email, slug, false)
					}
				} else {
					clientUUID = uuid.New().String()
					clientEmail = models.ExitRouteEmail(user.Email, slug, false)
				}
			}
			row := models.UserExit{
				UserID:   userID,
				NodeID:   ex.node.ID,
				UUID:     clientUUID,
				Email:    clientEmail,
				Priority: ex.priority,
				Active:   true,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		primaryID := exits[0].node.ID
		return tx.Model(&models.User{}).Where("id = ?", userID).
			Updates(map[string]interface{}{"exit_node_id": primaryID}).Error
	})
	if err != nil {
		return nil, err
	}
	return d.GetUser(userID)
}

// BackfillUserExits creates user_exits rows from legacy users.exit_node_id.
func (d *DB) BackfillUserExits() error {
	var users []models.User
	if err := d.Where("exit_node_id IS NOT NULL").Find(&users).Error; err != nil {
		return err
	}
	for _, u := range users {
		var count int64
		if err := d.Model(&models.UserExit{}).Where("user_id = ?", u.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 || u.ExitNodeID == nil {
			continue
		}
		node, err := d.GetNode(*u.ExitNodeID)
		if err != nil || node.Role != models.NodeRoleExit {
			continue
		}
		row := models.UserExit{
			UserID:   u.ID,
			NodeID:   node.ID,
			UUID:     u.UUID,
			Email:    u.Email,
			Priority: 100,
			Active:   true,
		}
		if err := d.Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

// ClearUserExits removes all exit assignments for a user.
func (d *DB) ClearUserExits(userID uint) error {
	return d.Where("user_id = ?", userID).Delete(&models.UserExit{}).Error
}

// ListUserExitAssignments joins user_exits with nodes.
func (d *DB) ListUserExitAssignments(userID uint) ([]models.UserExitAssignment, error) {
	rows, err := d.ListUserExits(userID)
	if err != nil {
		return nil, err
	}
	out := make([]models.UserExitAssignment, 0, len(rows))
	for _, row := range rows {
		node, err := d.GetNode(row.NodeID)
		if err != nil || !node.Active || node.Role != models.NodeRoleExit {
			continue
		}
		out = append(out, models.UserExitAssignment{UserExit: row, Node: *node})
	}
	if len(out) > 0 {
		return out, nil
	}
	// Legacy fallback: single ResolveUserExitNode with primary identity
	user, err := d.GetUser(userID)
	if err != nil {
		return nil, err
	}
	node, err := d.ResolveUserExitNode(user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return []models.UserExitAssignment{{
		UserExit: models.UserExit{
			UserID:   user.ID,
			NodeID:   node.ID,
			UUID:     user.UUID,
			Email:    user.Email,
			Priority: 100,
			Active:   true,
		},
		Node: *node,
	}}, nil
}

// ListAllExitAssignmentsForActiveUsers returns assignments for config generation.
func (d *DB) ListAllExitAssignmentsForActiveUsers(users []models.User) ([]models.UserExitAssignment, error) {
	var all []models.UserExitAssignment
	for _, u := range users {
		as, err := d.ListUserExitAssignments(u.ID)
		if err != nil {
			return nil, err
		}
		all = append(all, as...)
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].UserExit.UserID != all[j].UserExit.UserID {
			return all[i].UserExit.UserID < all[j].UserExit.UserID
		}
		return all[i].UserExit.Priority < all[j].UserExit.Priority
	})
	return all, nil
}
