package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"backend/config"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

type Group struct {
	ID          int    `json:"group_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    *int   `json:"parent_id"` // nullable
}

type GroupTreeNode struct {
	ID          int             `json:"group_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	ParentID    *int            `json:"parent_id,omitempty"`
	Depth       int             `json:"depth"`
	Children    []GroupTreeNode `json:"children,omitempty"`
}

type flatGroupNode struct {
	ID          int
	Name        string
	Description string
	ParentID    sql.NullInt32
	Depth       int
	Path        pq.Int64Array
}

/*
name: string
description: string
*/
func CreateGroup(w http.ResponseWriter, r *http.Request) {
	var group Group
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO Groups (name, description, parent_id) VALUES ($1, $2, $3) RETURNING group_id`
	err := config.PostgresDB.QueryRow(query, group.Name, group.Description, group.ParentID).Scan(&group.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

func GetGroups(w http.ResponseWriter, r *http.Request) {
	rows, err := config.PostgresDB.Query("SELECT group_id, name, description, parent_id FROM Groups")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var group Group
		if err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.ParentID); err != nil {
			http.Error(w, "Row scan error", http.StatusInternalServerError)
			return
		}
		groups = append(groups, group)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

/*
name: string
description: string
*/
func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	var group Group
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err := config.PostgresDB.Exec(
		"UPDATE Groups SET name = $1, description = $2, parent_id = $3 WHERE group_id = $4",
		group.Name, group.Description, group.ParentID, groupID,
	)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Check child groups
	var childCount int
	err := config.PostgresDB.QueryRow("SELECT COUNT(*) FROM Groups WHERE parent_id = $1", groupID).Scan(&childCount)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if childCount > 0 {
		http.Error(w, "Cannot delete group with child groups", http.StatusBadRequest)
		return
	}

	// Check users
	var userCount int
	err = config.PostgresDB.QueryRow("SELECT COUNT(*) FROM Users WHERE group_id = $1", groupID).Scan(&userCount)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if userCount > 0 {
		http.Error(w, "Cannot delete group with assigned users", http.StatusBadRequest)
		return
	}

	// Deelte group
	_, err = config.PostgresDB.Exec("DELETE FROM Groups WHERE group_id = $1", groupID)
	if err != nil {
		http.Error(w, "Failed to delete group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/groups/tree or /api/groups/tree?id=3
func GetGroupTree(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	idParam := query.Get("id")

	var rows *sql.Rows
	var err error

	if idParam == "" {
		rows, err = config.PostgresDB.Query("SELECT * FROM get_group_tree(NULL)")
	} else {
		groupID, errParse := strconv.Atoi(idParam)
		if errParse != nil {
			http.Error(w, "Invalid group ID", http.StatusBadRequest)
			return
		}
		rows, err = config.PostgresDB.Query("SELECT * FROM get_group_tree($1)", groupID)
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	nodes := []flatGroupNode{}
	for rows.Next() {
		var node flatGroupNode
		var parentID sql.NullInt32
		err := rows.Scan(&node.ID, &node.Name, &node.Description, &parentID, &node.Depth, &node.Path)
		if err != nil {
			http.Error(w, "Row scan error", http.StatusInternalServerError)
			return
		}
		if parentID.Valid {
			id := int(parentID.Int32)
			node.ParentID = sql.NullInt32{Int32: int32(id), Valid: true}
		}
		nodes = append(nodes, node)
	}

	tree := buildGroupTree(nodes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
}

func buildGroupTree(flatNodes []flatGroupNode) []GroupTreeNode {
	nodeMap := make(map[int]*GroupTreeNode)
	var roots []*GroupTreeNode

	// Create nodes
	for _, flat := range flatNodes {
		node := &GroupTreeNode{
			ID:          flat.ID,
			Name:        flat.Name,
			Description: flat.Description,
			Depth:       flat.Depth,
		}
		if flat.ParentID.Valid {
			id := int(flat.ParentID.Int32)
			node.ParentID = &id
		}
		nodeMap[node.ID] = node
	}

	// Connect parents with children
	for _, node := range nodeMap {
		if node.ParentID != nil {
			if parent, ok := nodeMap[*node.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
			} else {
				roots = append(roots, node)
			}
		} else {
			roots = append(roots, node)
		}
	}

	// []*GroupTreeNode -> []GroupTreeNode for JSON
	var result []GroupTreeNode
	for _, root := range roots {
		result = append(result, *root)
	}
	return result
}
