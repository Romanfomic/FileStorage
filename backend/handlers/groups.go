package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"backend/config"

	"github.com/gorilla/mux"
)

type Group struct {
	ID          int    `json:"group_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    *int   `json:"parent_id"` // nullable
}
type GroupTreeNode struct {
	ID          int              `json:"group_id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	ParentID    *int             `json:"parent_id,omitempty"`
	Depth       int              `json:"depth"`
	Children    []*GroupTreeNode `json:"children,omitempty"`
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

func GetGroupByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupIDStr := vars["id"]

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var group Group
	err = config.PostgresDB.QueryRow(`
		SELECT group_id, name, description, parent_id
		FROM Groups
		WHERE group_id = $1
	`, groupID).Scan(&group.ID, &group.Name, &group.Description, &group.ParentID)

	if err == sql.ErrNoRows {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
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

func GetGroupTree(w http.ResponseWriter, r *http.Request) {
	// Get id from query
	rootIDParam := r.URL.Query().Get("id")
	var rootID sql.NullInt32

	if rootIDParam != "" {
		id, err := strconv.Atoi(rootIDParam)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		rootID.Int32 = int32(id)
		rootID.Valid = true
	}

	rows, err := config.PostgresDB.Query(`
		SELECT group_id, name, description, parent_id, depth
		FROM get_group_tree($1)
	`, rootID)
	if err != nil {
		http.Error(w, "DB query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	nodes := make(map[int]*GroupTreeNode)
	var roots []*GroupTreeNode

	for rows.Next() {
		var g GroupTreeNode
		var parentID sql.NullInt32

		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &parentID, &g.Depth); err != nil {
			http.Error(w, "Row scan error", http.StatusInternalServerError)
			return
		}
		if parentID.Valid {
			pid := int(parentID.Int32)
			g.ParentID = &pid
		}

		nodes[g.ID] = &g

		// create tree
		if g.Depth == 0 {
			roots = append(roots, &g)
		} else if g.ParentID != nil {
			parent, exists := nodes[*g.ParentID]
			if exists {
				parent.Children = append(parent.Children, &g)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if len(roots) > 0 {
		json.NewEncoder(w).Encode(roots)
	} else {
		w.Write([]byte("[]"))
	}
}
