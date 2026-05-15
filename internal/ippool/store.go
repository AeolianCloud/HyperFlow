package ippool

import (
	"database/sql"
	"fmt"
	"time"

	"hyperflow/internal/timeutil"
)

type Pool struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Gateway     string   `json:"gateway"`
	Netmask     int      `json:"netmask"`
	DNS1        string   `json:"dns1,omitempty"`
	DNS2        string   `json:"dns2,omitempty"`
	Description string   `json:"description,omitempty"`
	Nodes       []string `json:"nodes"`
	Total       int      `json:"total"`
	Available   int      `json:"available"`
	Used        int      `json:"used"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Address struct {
	ID        string    `json:"id"`
	PoolID    string    `json:"poolId"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`
	VMID      *int      `json:"vmId,omitempty"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Store interface {
	CreatePool(p *Pool) error
	GetPool(id string) (*Pool, error)
	ListPools() ([]*Pool, error)
	UpdatePool(p *Pool) error
	DeletePool(id string) error

	BindNodes(poolID string, nodes []string) error
	GetPoolNodes(poolID string) ([]string, error)

	InsertAddresses(addrs []Address) error
	DeleteAddresses(poolID string, addresses []string) error
	ListAddresses(poolID, status string, page, size int) ([]Address, int, error)
	AllocateAddress(poolID, address string, vmID int) (*Address, error)
	AllocateRandomAddress(poolID string, vmID int) (*Address, error)
	ReleaseAddressByVMID(vmID int) error
	CountByPoolID(poolID string) (total, available, used int, err error)
	HasUsedAddresses(poolID string) (bool, error)
	CheckAddressConflict(address string, excludePoolID string) (bool, error)
	GetPoolByNode(node string) ([]*Pool, error)
}

func NewMySQLStore(db *sql.DB) Store {
	return &mysqlStore{db: db}
}

type mysqlStore struct {
	db *sql.DB
}

func (s *mysqlStore) CreatePool(p *Pool) error {
	now := timeutil.NowShanghai()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO ip_pools (id, name, gateway, netmask, dns1, dns2, description, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Gateway, p.Netmask,
		nullableString(p.DNS1), nullableString(p.DNS2), nullableString(p.Description),
		now, now,
	)
	if err != nil {
		return err
	}

	if len(p.Nodes) > 0 {
		if err := s.bindNodesTx(tx, p.ID, p.Nodes); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *mysqlStore) GetPool(id string) (*Pool, error) {
	row := s.db.QueryRow(
		`SELECT id, name, gateway, netmask, dns1, dns2, description, created_at, updated_at
		 FROM ip_pools WHERE id = ?`, id,
	)
	p := &Pool{}
	var dns1, dns2, desc sql.NullString
	err := row.Scan(&p.ID, &p.Name, &p.Gateway, &p.Netmask, &dns1, &dns2, &desc, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.DNS1 = dns1.String
	p.DNS2 = dns2.String
	p.Description = desc.String
	p.CreatedAt = timeutil.InShanghai(p.CreatedAt)
	p.UpdatedAt = timeutil.InShanghai(p.UpdatedAt)

	nodes, err := s.GetPoolNodes(id)
	if err != nil {
		return nil, err
	}
	p.Nodes = nodes

	total, available, used, err := s.CountByPoolID(id)
	if err != nil {
		return nil, err
	}
	p.Total = total
	p.Available = available
	p.Used = used

	return p, nil
}

func (s *mysqlStore) ListPools() ([]*Pool, error) {
	rows, err := s.db.Query(
		`SELECT id, name, gateway, netmask, dns1, dns2, description, created_at, updated_at
		 FROM ip_pools ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pools []*Pool
	for rows.Next() {
		p := &Pool{}
		var dns1, dns2, desc sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.Gateway, &p.Netmask, &dns1, &dns2, &desc, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.DNS1 = dns1.String
		p.DNS2 = dns2.String
		p.Description = desc.String
		p.CreatedAt = timeutil.InShanghai(p.CreatedAt)
		p.UpdatedAt = timeutil.InShanghai(p.UpdatedAt)
		pools = append(pools, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, p := range pools {
		nodes, err := s.GetPoolNodes(p.ID)
		if err != nil {
			return nil, err
		}
		p.Nodes = nodes
		total, available, used, err := s.CountByPoolID(p.ID)
		if err != nil {
			return nil, err
		}
		p.Total = total
		p.Available = available
		p.Used = used
	}
	return pools, nil
}

func (s *mysqlStore) UpdatePool(p *Pool) error {
	now := timeutil.NowShanghai()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE ip_pools SET name = ?, dns1 = ?, dns2 = ?, description = ?, updated_at = ? WHERE id = ?`,
		p.Name, nullableString(p.DNS1), nullableString(p.DNS2), nullableString(p.Description), now, p.ID,
	)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM ip_pool_nodes WHERE pool_id = ?`, p.ID); err != nil {
		return err
	}
	if len(p.Nodes) > 0 {
		if err := s.bindNodesTx(tx, p.ID, p.Nodes); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *mysqlStore) DeletePool(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM ip_pool_nodes WHERE pool_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM ip_pool_addresses WHERE pool_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM ip_pools WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *mysqlStore) BindNodes(poolID string, nodes []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM ip_pool_nodes WHERE pool_id = ?`, poolID); err != nil {
		return err
	}
	if err := s.bindNodesTx(tx, poolID, nodes); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *mysqlStore) bindNodesTx(tx *sql.Tx, poolID string, nodes []string) error {
	for _, node := range nodes {
		if _, err := tx.Exec(
			`INSERT INTO ip_pool_nodes (pool_id, node) VALUES (?, ?)`,
			poolID, node,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *mysqlStore) GetPoolNodes(poolID string) ([]string, error) {
	rows, err := s.db.Query(`SELECT node FROM ip_pool_nodes WHERE pool_id = ? ORDER BY node`, poolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []string
	for rows.Next() {
		var node string
		if err := rows.Scan(&node); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func (s *mysqlStore) InsertAddresses(addrs []Address) error {
	if len(addrs) == 0 {
		return nil
	}
	now := timeutil.NowShanghai()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO ip_pool_addresses (id, pool_id, address, status, vm_id, created_at, updated_at)
		 VALUES (?, ?, ?, 'available', NULL, ?, ?)`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, a := range addrs {
		if _, err := stmt.Exec(a.ID, a.PoolID, a.Address, now, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *mysqlStore) DeleteAddresses(poolID string, addresses []string) error {
	if len(addresses) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`DELETE FROM ip_pool_addresses WHERE pool_id = ? AND address = ? AND status = 'available'`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, addr := range addresses {
		result, execErr := stmt.Exec(poolID, addr)
		if execErr != nil {
			return execErr
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("address %s cannot be deleted: not found or not available", addr)
		}
	}
	return tx.Commit()
}

func (s *mysqlStore) ListAddresses(poolID, status string, page, size int) ([]Address, int, error) {
	var total int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM ip_pool_addresses WHERE pool_id = ?`, poolID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	offset := (page - 1) * size

	var rows *sql.Rows
	if status != "" {
		rows, err = s.db.Query(
			`SELECT id, pool_id, address, status, vm_id, created_at, updated_at
			 FROM ip_pool_addresses
			 WHERE pool_id = ? AND status = ?
			 ORDER BY address ASC
			 LIMIT ? OFFSET ?`,
			poolID, status, size, offset,
		)
	} else {
		rows, err = s.db.Query(
			`SELECT id, pool_id, address, status, vm_id, created_at, updated_at
			 FROM ip_pool_addresses
			 WHERE pool_id = ?
			 ORDER BY address ASC
			 LIMIT ? OFFSET ?`,
			poolID, size, offset,
		)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var addrs []Address
	for rows.Next() {
		var a Address
		var vmID sql.NullInt32
		if err := rows.Scan(&a.ID, &a.PoolID, &a.Address, &a.Status, &vmID, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if vmID.Valid {
			v := int(vmID.Int32)
			a.VMID = &v
		}
		addrs = append(addrs, a)
	}
	return addrs, total, rows.Err()
}

func (s *mysqlStore) AllocateAddress(poolID, address string, vmID int) (*Address, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := timeutil.NowShanghai()
	var addr Address
	var vmid sql.NullInt32
	err = tx.QueryRow(
		`SELECT id, pool_id, address, status, vm_id
		 FROM ip_pool_addresses
		 WHERE pool_id = ? AND address = ?
		 FOR UPDATE`,
		poolID, address,
	).Scan(&addr.ID, &addr.PoolID, &addr.Address, &addr.Status, &vmid)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if addr.Status != "available" {
		return nil, nil
	}

	_, err = tx.Exec(
		`UPDATE ip_pool_addresses SET status = 'reserved', vm_id = ?, updated_at = ? WHERE id = ?`,
		vmID, now, addr.ID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	addr.Status = "reserved"
	addr.VMID = &vmID
	return &addr, nil
}

func (s *mysqlStore) AllocateRandomAddress(poolID string, vmID int) (*Address, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var addr Address
	var vmid sql.NullInt32
	err = tx.QueryRow(
		`SELECT id, pool_id, address, status, vm_id
		 FROM ip_pool_addresses
		 WHERE pool_id = ? AND status = 'available'
		 ORDER BY RAND()
		 LIMIT 1
		 FOR UPDATE`,
		poolID,
	).Scan(&addr.ID, &addr.PoolID, &addr.Address, &addr.Status, &vmid)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	now := timeutil.NowShanghai()
	_, err = tx.Exec(
		`UPDATE ip_pool_addresses SET status = 'reserved', vm_id = ?, updated_at = ? WHERE id = ?`,
		vmID, now, addr.ID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	addr.Status = "reserved"
	addr.VMID = &vmID
	return &addr, nil
}

func (s *mysqlStore) ReleaseAddressByVMID(vmID int) error {
	now := timeutil.NowShanghai()
	_, err := s.db.Exec(
		`UPDATE ip_pool_addresses SET status = 'available', vm_id = NULL, updated_at = ? WHERE vm_id = ?`,
		now, vmID,
	)
	return err
}

func (s *mysqlStore) CountByPoolID(poolID string) (total, available, used int, err error) {
	row := s.db.QueryRow(`SELECT COUNT(*), SUM(CASE WHEN status='available' THEN 1 ELSE 0 END), SUM(CASE WHEN status='used' THEN 1 ELSE 0 END) FROM ip_pool_addresses WHERE pool_id = ?`, poolID)
	var tot, avail, us sql.NullInt64
	if err := row.Scan(&tot, &avail, &us); err != nil {
		return 0, 0, 0, err
	}
	return int(tot.Int64), int(avail.Int64), int(us.Int64), nil
}

func (s *mysqlStore) HasUsedAddresses(poolID string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM ip_pool_addresses WHERE pool_id = ? AND status IN ('used', 'reserved')`,
		poolID,
	).Scan(&count)
	return count > 0, err
}

func (s *mysqlStore) CheckAddressConflict(address string, excludePoolID string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM ip_pool_addresses WHERE address = ? AND pool_id != ?`,
		address, excludePoolID,
	).Scan(&count)
	return count > 0, err
}

func (s *mysqlStore) GetPoolByNode(node string) ([]*Pool, error) {
	rows, err := s.db.Query(
		`SELECT p.id, p.name, p.gateway, p.netmask, p.dns1, p.dns2, p.description, p.created_at, p.updated_at
		 FROM ip_pools p
		 JOIN ip_pool_nodes n ON n.pool_id = p.id
		 WHERE n.node = ?
		 ORDER BY p.name`, node,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pools []*Pool
	for rows.Next() {
		p := &Pool{}
		var dns1, dns2, desc sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.Gateway, &p.Netmask, &dns1, &dns2, &desc, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.DNS1 = dns1.String
		p.DNS2 = dns2.String
		p.Description = desc.String
		p.CreatedAt = timeutil.InShanghai(p.CreatedAt)
		p.UpdatedAt = timeutil.InShanghai(p.UpdatedAt)
		pools = append(pools, p)
	}
	return pools, rows.Err()
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
