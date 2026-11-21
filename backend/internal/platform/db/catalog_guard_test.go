package db

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tenantTable struct {
	name             string
	rls              bool
	forcedRLS        bool
	emailAppPolicies int
}

type foreignKey struct {
	name          string
	childTable    string
	parentTable   string
	childColumns  []string
	parentColumns []string
}

func TestAC_SEC_01_CatalogGuard(t *testing.T) {
	raw := openTestDB(t)

	tables := loadTenantTables(t, raw)
	problems := checkTenantTables(tables)
	require.Empty(t, problems, strings.Join(problems, "\n"))

	tenantNames := make(map[string]bool, len(tables))
	for _, table := range tables {
		tenantNames[table.name] = true
	}
	problems = checkCompositeFKs(tenantNames, loadForeignKeys(t, raw))
	require.Empty(t, problems, strings.Join(problems, "\n"))
}

func TestAC_SEC_01_CatalogGuardDetectsViolation(t *testing.T) {
	t.Run("missing forced RLS and policy", func(t *testing.T) {
		problems := checkTenantTables([]tenantTable{{name: "widgets", rls: true}})
		require.Len(t, problems, 2)
		assert.Contains(t, strings.Join(problems, "\n"), "widgets")
	})

	t.Run("empty catalog", func(t *testing.T) {
		assert.NotEmpty(t, checkTenantTables(nil))
	})

	t.Run("non-composite foreign key", func(t *testing.T) {
		problems := checkCompositeFKs(
			map[string]bool{"widgets": true, "things": true},
			[]foreignKey{{
				name:          "widgets_thing_id_fkey",
				childTable:    "widgets",
				parentTable:   "things",
				childColumns:  []string{"thing_id"},
				parentColumns: []string{"id"},
			}},
		)
		require.Len(t, problems, 1)
		assert.Contains(t, problems[0], "widgets_thing_id_fkey")
	})

	t.Run("composite foreign key passes", func(t *testing.T) {
		problems := checkCompositeFKs(
			map[string]bool{"widgets": true, "things": true},
			[]foreignKey{{
				name:          "widgets_thing_fkey",
				childTable:    "widgets",
				parentTable:   "things",
				childColumns:  []string{"organization_id", "thing_id"},
				parentColumns: []string{"organization_id", "id"},
			}},
		)
		assert.Empty(t, problems)
	})
}

func checkTenantTables(tables []tenantTable) []string {
	if len(tables) == 0 {
		return []string{"no tenant tables found in the catalog"}
	}

	var problems []string
	for _, table := range tables {
		if !table.rls {
			problems = append(problems, table.name+": row-level security is not enabled")
		}
		if !table.forcedRLS {
			problems = append(problems, table.name+": row-level security is not forced")
		}
		if table.emailAppPolicies == 0 {
			problems = append(problems, table.name+": no policy applies to email_app")
		}
	}
	return problems
}

func checkCompositeFKs(tenantNames map[string]bool, fks []foreignKey) []string {
	var problems []string
	for _, fk := range fks {
		if !tenantNames[fk.childTable] || !tenantNames[fk.parentTable] {
			continue
		}
		if fk.parentTable == "organizations" {
			continue
		}
		if !hasColumn(fk.childColumns, "organization_id") || !hasColumn(fk.parentColumns, "organization_id") {
			problems = append(problems, fmt.Sprintf(
				"%s: foreign key %s -> %s must be composite on organization_id (child %v, parent %v)",
				fk.childTable, fk.name, fk.parentTable, fk.childColumns, fk.parentColumns,
			))
		}
	}
	return problems
}

func hasColumn(columns []string, name string) bool {
	for _, column := range columns {
		if column == name {
			return true
		}
	}
	return false
}

func loadTenantTables(t *testing.T, db *sql.DB) []tenantTable {
	t.Helper()

	rows, err := db.Query(`
		SELECT c.relname,
		       c.relrowsecurity,
		       c.relforcerowsecurity,
		       (SELECT count(*)
		          FROM pg_policies p
		         WHERE p.schemaname = 'public'
		           AND p.tablename = c.relname
		           AND ('email_app'::name = ANY(p.roles) OR 'public'::name = ANY(p.roles)))
		  FROM pg_class c
		  JOIN pg_namespace n ON n.oid = c.relnamespace
		 WHERE n.nspname = 'public'
		   AND c.relkind IN ('r', 'p')
		   AND NOT c.relispartition
		   AND (c.relname = 'organizations'
		        OR EXISTS (
		            SELECT 1 FROM pg_attribute a
		             WHERE a.attrelid = c.oid
		               AND a.attname = 'organization_id'
		               AND a.attnum > 0
		               AND NOT a.attisdropped
		        ))
		 ORDER BY c.relname`)
	require.NoError(t, err)
	defer rows.Close()

	var tables []tenantTable
	for rows.Next() {
		var table tenantTable
		require.NoError(t, rows.Scan(&table.name, &table.rls, &table.forcedRLS, &table.emailAppPolicies))
		tables = append(tables, table)
	}
	require.NoError(t, rows.Err())

	return tables
}

func loadForeignKeys(t *testing.T, db *sql.DB) []foreignKey {
	t.Helper()

	rows, err := db.Query(`
		SELECT con.conname,
		       child.relname,
		       parent.relname,
		       array_to_string(ARRAY(
		           SELECT a.attname
		             FROM unnest(con.conkey) WITH ORDINALITY AS k(attnum, ord)
		             JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = k.attnum
		            ORDER BY k.ord), ','),
		       array_to_string(ARRAY(
		           SELECT a.attname
		             FROM unnest(con.confkey) WITH ORDINALITY AS k(attnum, ord)
		             JOIN pg_attribute a ON a.attrelid = con.confrelid AND a.attnum = k.attnum
		            ORDER BY k.ord), ',')
		  FROM pg_constraint con
		  JOIN pg_class child ON child.oid = con.conrelid
		  JOIN pg_class parent ON parent.oid = con.confrelid
		  JOIN pg_namespace n ON n.oid = child.relnamespace
		 WHERE con.contype = 'f'
		   AND n.nspname = 'public'
		   AND NOT child.relispartition
		 ORDER BY child.relname, con.conname`)
	require.NoError(t, err)
	defer rows.Close()

	var fks []foreignKey
	for rows.Next() {
		var fk foreignKey
		var childColumns, parentColumns string
		require.NoError(t, rows.Scan(&fk.name, &fk.childTable, &fk.parentTable, &childColumns, &parentColumns))
		fk.childColumns = splitColumns(childColumns)
		fk.parentColumns = splitColumns(parentColumns)
		fks = append(fks, fk)
	}
	require.NoError(t, rows.Err())

	return fks
}

func splitColumns(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}
