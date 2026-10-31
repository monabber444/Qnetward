package airca

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type QueryFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // "=", ">", "<", ">=", "<=", "LIKE"
	Value    any    `json:"value"`
}

type StructuredQuery struct {
	TargetTable string        `json:"target_table"` // "probe_results", "alerts", "mesh_metrics"
	Filters     []QueryFilter `json:"filters"`
	TimeWindow  time.Duration `json:"time_window"`
	Limit       int           `json:"limit"`
	SQLWhere    string        `json:"sql_where"`
}

type NLQueryTranslator struct{}

func NewNLQueryTranslator() *NLQueryTranslator {
	return &NLQueryTranslator{}
}

// Translate converts an operator natural language string into a structured query and SQL WHERE clause.
func (t *NLQueryTranslator) Translate(query string) *StructuredQuery {
	q := strings.ToLower(strings.TrimSpace(query))
	res := &StructuredQuery{
		TargetTable: "probe_results",
		Filters:     make([]QueryFilter, 0),
		Limit:       100,
	}

	// 1. Determine target table
	if strings.Contains(q, "alert") || strings.Contains(q, "incident") {
		res.TargetTable = "alerts"
	} else if strings.Contains(q, "mesh") || strings.Contains(q, "matrix") {
		res.TargetTable = "mesh_metrics"
	}

	// 2. Parse Packet Loss threshold e.g. "loss > 5%" or "packet loss > 10"
	lossRe := regexp.MustCompile(`(?:loss|packet loss)\s*([><=]+)\s*(\d+)%?`)
	if m := lossRe.FindStringSubmatch(q); len(m) == 3 {
		val, _ := strconv.ParseFloat(m[2], 64)
		op := m[1]
		field := "packet_loss_percent"
		res.Filters = append(res.Filters, QueryFilter{
			Field:    field,
			Operator: op,
			Value:    val,
		})
	}

	// 3. Parse Latency threshold e.g. "latency > 100ms" or "rtt >= 50"
	latRe := regexp.MustCompile(`(?:latency|rtt|ping)\s*([><=]+)\s*(\d+)(?:ms)?`)
	if m := latRe.FindStringSubmatch(q); len(m) == 3 {
		val, _ := strconv.ParseFloat(m[2], 64)
		op := m[1]
		res.Filters = append(res.Filters, QueryFilter{
			Field:    "rtt_ms",
			Operator: op,
			Value:    val,
		})
	}

	// 4. Parse Severity e.g. "critical", "warning"
	if strings.Contains(q, "critical") {
		res.Filters = append(res.Filters, QueryFilter{
			Field:    "severity",
			Operator: "=",
			Value:    "critical",
		})
	} else if strings.Contains(q, "warning") {
		res.Filters = append(res.Filters, QueryFilter{
			Field:    "severity",
			Operator: "=",
			Value:    "warning",
		})
	}

	// 5. Parse Region / Tag e.g. "in us-east" or "in eu-west"
	regRe := regexp.MustCompile(`in\s+([a-z0-9\-]+)`)
	if m := regRe.FindStringSubmatch(q); len(m) == 2 {
		res.Filters = append(res.Filters, QueryFilter{
			Field:    "region",
			Operator: "=",
			Value:    m[1],
		})
	}

	// 6. Parse Target / Host substring e.g. "on core-router" or "target 8.8.8.8"
	targetRe := regexp.MustCompile(`(?:on|target)\s+([a-zA-Z0-9\.\-_]+)`)
	if m := targetRe.FindStringSubmatch(q); len(m) == 2 {
		res.Filters = append(res.Filters, QueryFilter{
			Field:    "target",
			Operator: "LIKE",
			Value:    "%" + m[1] + "%",
		})
	}

	// 7. Parse Time window e.g. "last 1 hour", "last 24 hours", "last 15 minutes"
	timeRe := regexp.MustCompile(`last\s+(\d+)\s*(hour|minute|day|h|m|d)s?`)
	if m := timeRe.FindStringSubmatch(q); len(m) == 3 {
		n, _ := strconv.Atoi(m[1])
		unit := m[2]
		switch {
		case strings.HasPrefix(unit, "h"):
			res.TimeWindow = time.Duration(n) * time.Hour
		case strings.HasPrefix(unit, "m"):
			res.TimeWindow = time.Duration(n) * time.Minute
		case strings.HasPrefix(unit, "d"):
			res.TimeWindow = time.Duration(n) * 24 * time.Hour
		}
	} else {
		res.TimeWindow = 1 * time.Hour // default 1 hour
	}

	// 8. Construct SQL WHERE clause
	var clauses []string
	if res.TimeWindow > 0 {
		clauses = append(clauses, fmt.Sprintf("timestamp >= NOW() - INTERVAL '%d seconds'", int(res.TimeWindow.Seconds())))
	}
	for _, f := range res.Filters {
		if f.Operator == "LIKE" {
			clauses = append(clauses, fmt.Sprintf("%s ILIKE '%s'", f.Field, f.Value))
		} else if str, ok := f.Value.(string); ok {
			clauses = append(clauses, fmt.Sprintf("%s %s '%s'", f.Field, f.Operator, str))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s %s %v", f.Field, f.Operator, f.Value))
		}
	}

	res.SQLWhere = strings.Join(clauses, " AND ")
	return res
}
