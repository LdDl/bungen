package model

import (
	"strings"

	"github.com/LdDl/bungen/util"
)

// Relation stores relation
type Relation struct {
	PKFields []string
	FKFields []string
	GoName   string

	TargetPGName     string
	TargetPGSchema   string
	TargetPGFullName string

	TargetEntity *Entity

	GoType string
}

// NewRelation creates relation from Postgres info
func NewRelation(sourceColumns []string, targetSchema, targetTable string, targetColumns []string) Relation {
	names := make([]string, len(sourceColumns))
	for i, name := range sourceColumns {
		names[i] = util.ReplaceSuffix(util.ColumnName(name), util.ID, "")
	}

	typ := util.EntityName(targetTable)
	if targetSchema != util.PublicSchema {
		typ = util.CamelCased(targetSchema) + typ
	}

	return Relation{
		PKFields: targetColumns,
		FKFields: sourceColumns,
		GoName:   strings.Join(names, ""),

		TargetPGName:     targetTable,
		TargetPGSchema:   targetSchema,
		TargetPGFullName: util.JoinF(targetSchema, targetTable),

		GoType: typ,
	}
}

func (r *Relation) AddEntity(entity *Entity) {
	r.TargetEntity = entity
}
