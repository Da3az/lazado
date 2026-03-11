package api

import (
	"fmt"
	"strings"
)

// WIQL provides a builder for Work Item Query Language queries.
type WIQL struct {
	fields  []string
	from    string
	where   string
	orderBy []string
	top     int
}

// NewWIQL creates a new WIQL query builder.
func NewWIQL() *WIQL {
	return &WIQL{
		from: "WorkItems",
	}
}

// Select sets the fields to return.
func (q *WIQL) Select(fields ...string) *WIQL {
	q.fields = fields
	return q
}

// Where sets the WHERE clause.
func (q *WIQL) Where(clause string) *WIQL {
	q.where = clause
	return q
}

// Top limits the number of results returned.
func (q *WIQL) Top(n int) *WIQL {
	q.top = n
	return q
}

// OrderByDesc adds an ORDER BY ... DESC clause.
func (q *WIQL) OrderByDesc(field string) *WIQL {
	q.orderBy = append(q.orderBy, fmt.Sprintf("[%s] DESC", field))
	return q
}

// OrderByAsc adds an ORDER BY ... ASC clause.
func (q *WIQL) OrderByAsc(field string) *WIQL {
	q.orderBy = append(q.orderBy, fmt.Sprintf("[%s] ASC", field))
	return q
}

// Build constructs the WIQL query string.
func (q *WIQL) Build() string {
	var b strings.Builder
	b.WriteString("SELECT ")
	if q.top > 0 {
		b.WriteString(fmt.Sprintf("TOP %d ", q.top))
	}
	if len(q.fields) > 0 {
		quoted := make([]string, len(q.fields))
		for i, f := range q.fields {
			quoted[i] = "[" + f + "]"
		}
		b.WriteString(strings.Join(quoted, ", "))
	} else {
		b.WriteString("[System.Id]")
	}
	b.WriteString(" FROM ")
	b.WriteString(q.from)

	if q.where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(q.where)
	}

	if len(q.orderBy) > 0 {
		b.WriteString(" ORDER BY ")
		b.WriteString(strings.Join(q.orderBy, ", "))
	}

	return b.String()
}

// Condition helpers for building WHERE clauses

// Eq returns an equality condition: [field] = 'value'
func Eq(field, value string) string {
	return fmt.Sprintf("[%s] = '%s'", field, escapeWIQL(value))
}

// EqMe returns a condition matching the current user: [field] = @Me
func EqMe(field string) string {
	return fmt.Sprintf("[%s] = @Me", field)
}

// Neq returns a not-equal condition: [field] <> 'value'
func Neq(field, value string) string {
	return fmt.Sprintf("[%s] <> '%s'", field, escapeWIQL(value))
}

// Contains returns a CONTAINS condition for text search.
func Contains(field, value string) string {
	return fmt.Sprintf("[%s] CONTAINS '%s'", field, escapeWIQL(value))
}

// In returns an IN condition: [field] IN ('v1', 'v2', ...)
func In(field string, values ...string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("'%s'", escapeWIQL(v))
	}
	return fmt.Sprintf("[%s] IN (%s)", field, strings.Join(quoted, ", "))
}

// And joins conditions with AND.
func And(conditions ...string) string {
	return "(" + strings.Join(conditions, " AND ") + ")"
}

// Or joins conditions with OR.
func Or(conditions ...string) string {
	return "(" + strings.Join(conditions, " OR ") + ")"
}

// UnderArea returns a condition for items under a specific area path.
func UnderArea(areaPath string) string {
	return fmt.Sprintf("[System.AreaPath] UNDER '%s'", escapeWIQL(areaPath))
}

// CurrentIteration returns a condition for items in the current iteration.
func CurrentIteration() string {
	return "[System.IterationPath] = @CurrentIteration"
}

func escapeWIQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
