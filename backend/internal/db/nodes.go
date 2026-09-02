package db

import (
	"errors"

	"rionexgate/internal/models"

	"gorm.io/gorm"
)

type CreateNodeInput struct {
	Name        string
	Address     string
	Port        int
	Active      bool
	Role        string
	Protocol    string
	Credentials string
	Region      string
	Priority    int
}

type UpdateNodeInput struct {
	Name        *string
	Address     *string
	Port        *int
	Active      *bool
	Role        *string
	Protocol    *string
	Credentials *string
	Region      *string
	Priority    *int
}

func validateNodeRole(role string) error {
	if role == "" || role == models.NodeRoleEntry || role == models.NodeRoleExit {
		return nil
	}
	return errors.New("role must be entry or exit")
}

func (d *DB) ListNodes() ([]models.Node, error) {
	var nodes []models.Node
	err := d.Order("priority asc, id asc").Find(&nodes).Error
	return nodes, err
}

func (d *DB) ListActiveNodesByRole(role string) ([]models.Node, error) {
	var nodes []models.Node
	err := d.Where("active = ? AND role = ?", true, role).
		Order("priority asc, id asc").
		Find(&nodes).Error
	return nodes, err
}

func (d *DB) GetNode(id uint) (*models.Node, error) {
	var node models.Node
	err := d.First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (d *DB) GetBestEntryNode() (*models.Node, error) {
	nodes, err := d.ListActiveNodesByRole(models.NodeRoleEntry)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &nodes[0], nil
}

func (d *DB) GetBestExitNode() (*models.Node, error) {
	nodes, err := d.ListActiveNodesByRole(models.NodeRoleExit)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &nodes[0], nil
}

func (d *DB) CreateNode(in CreateNodeInput) (*models.Node, error) {
	if in.Name == "" {
		return nil, errors.New("name is required")
	}
	if in.Address == "" {
		return nil, errors.New("address is required")
	}
	if in.Port <= 0 {
		return nil, errors.New("port must be positive")
	}
	role := in.Role
	if role == "" {
		role = models.NodeRoleEntry
	}
	if err := validateNodeRole(role); err != nil {
		return nil, err
	}
	proto := in.Protocol
	if proto == "" {
		proto = "vless"
	}
	priority := in.Priority
	if priority <= 0 {
		priority = 100
	}
	node := &models.Node{
		Name:        in.Name,
		Address:     in.Address,
		Port:        in.Port,
		Active:      in.Active,
		Role:        role,
		Protocol:    proto,
		Credentials: in.Credentials,
		Region:      in.Region,
		Priority:    priority,
	}
	if err := d.Create(node).Error; err != nil {
		return nil, err
	}
	return node, nil
}

func (d *DB) UpdateNode(id uint, in UpdateNodeInput) (*models.Node, error) {
	node, err := d.GetNode(id)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Address != nil {
		updates["address"] = *in.Address
	}
	if in.Port != nil {
		if *in.Port <= 0 {
			return nil, errors.New("port must be positive")
		}
		updates["port"] = *in.Port
	}
	if in.Active != nil {
		updates["active"] = *in.Active
	}
	if in.Role != nil {
		if err := validateNodeRole(*in.Role); err != nil {
			return nil, err
		}
		updates["role"] = *in.Role
	}
	if in.Protocol != nil {
		updates["protocol"] = *in.Protocol
	}
	if in.Credentials != nil {
		updates["credentials"] = *in.Credentials
	}
	if in.Region != nil {
		updates["region"] = *in.Region
	}
	if in.Priority != nil {
		updates["priority"] = *in.Priority
	}
	if len(updates) > 0 {
		if err := d.Model(node).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return d.GetNode(id)
}

func (d *DB) DeleteNode(id uint) error {
	return d.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("entry_node_id = ?", id).Update("entry_node_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.User{}).Where("exit_node_id = ?", id).Update("exit_node_id", nil).Error; err != nil {
			return err
		}
		res := tx.Delete(&models.Node{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (d *DB) ResolveUserEntryNode(user *models.User) (*models.Node, error) {
	if user.EntryNodeID != nil {
		node, err := d.GetNode(*user.EntryNodeID)
		if err == nil && node.Active && node.Role == models.NodeRoleEntry {
			return node, nil
		}
	}
	return d.GetBestEntryNode()
}

func (d *DB) ResolveUserExitNode(user *models.User) (*models.Node, error) {
	if user.ExitNodeID != nil {
		node, err := d.GetNode(*user.ExitNodeID)
		if err == nil && node.Active && node.Role == models.NodeRoleExit {
			return node, nil
		}
	}
	return d.GetBestExitNode()
}
