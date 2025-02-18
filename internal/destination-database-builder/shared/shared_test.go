package destinationdatabasebuilder_shared

import (
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/assert"
)

func Test_getFilteredForeignToPrimaryTableMap(t *testing.T) {
	t.Parallel()
	tables := map[string]struct{}{
		"public.regions":     {},
		"public.jobs":        {},
		"public.countries":   {},
		"public.locations":   {},
		"public.dependents":  {},
		"public.departments": {},
		"public.employees":   {},
	}
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.countries": {
			{Columns: []string{"region_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.regions", Columns: []string{"region_id"}}},
		},
		"public.departments": {
			{Columns: []string{"location_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.locations", Columns: []string{"location_id"}}},
		},
		"public.dependents": {
			{Columns: []string{"dependent_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employees_id"}}},
		},
		"public.locations": {
			{Columns: []string{"country_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.countries", Columns: []string{"country_id"}}},
		},
		"public.employees": {
			{Columns: []string{"department_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.departments", Columns: []string{"department_id"}}},
			{Columns: []string{"job_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.jobs", Columns: []string{"job_id"}}},
			{Columns: []string{"manager_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employee_id"}}},
		},
	}

	expected := map[string][]string{
		"public.regions":     {},
		"public.jobs":        {},
		"public.countries":   {"public.regions"},
		"public.departments": {"public.locations"},
		"public.dependents":  {"public.employees"},
		"public.employees":   {"public.departments", "public.jobs", "public.employees"},
		"public.locations":   {"public.countries"},
	}
	actual := GetFilteredForeignToPrimaryTableMap(dependencies, tables)
	assert.Len(t, actual, len(expected))
	for table, deps := range actual {
		assert.Len(t, deps, len(expected[table]))
		assert.ElementsMatch(t, expected[table], deps)
	}
}

func Test_getFilteredForeignToPrimaryTableMap_filtered(t *testing.T) {
	t.Parallel()
	tables := map[string]struct{}{
		"public.countries": {},
	}
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.countries": {
			{Columns: []string{"region_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.regions", Columns: []string{"region_id"}}}},

		"public.departments": {
			{Columns: []string{"location_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.locations", Columns: []string{"location_id"}}},
		},
		"public.dependents": {
			{Columns: []string{"dependent_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employees_id"}}},
		},
		"public.locations": {
			{Columns: []string{"country_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.countries", Columns: []string{"country_id"}}},
		},
		"public.employees": {
			{Columns: []string{"department_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.departments", Columns: []string{"department_id"}}},
			{Columns: []string{"job_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.jobs", Columns: []string{"job_id"}}},
			{Columns: []string{"manager_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employee_id"}}},
		},
	}

	expected := map[string][]string{
		"public.countries": {},
	}
	actual := GetFilteredForeignToPrimaryTableMap(dependencies, tables)
	assert.Len(t, actual, len(expected))
	for table, deps := range actual {
		assert.Len(t, deps, len(expected[table]))
		assert.ElementsMatch(t, expected[table], deps)
	}
}
